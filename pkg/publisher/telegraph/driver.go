package telegraph

import (
	"fmt"
	urlpkg "net/url"
	"strings"
	"sync"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

// nolint:revive
const (
	Name = "telegraph"
)

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	client client

	defaultAccountShortName string

	account telegraphAccount
	page    telegraphPage

	mu sync.RWMutex
}

// CreateNew implements publisher.Interface
func (d *Driver) CreateNew(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	var rd strings.Reader
	rd.Reset(in.Data.Get())
	nodes, err := htmlToNodes(&rd)
	if err != nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.account.AccessToken) == 0 {
		err = fmt.Errorf("account not set")
		return
	}

	d.page, err = d.client.CreatePage(createPageOptions{
		AccessToken: d.account.AccessToken,
		AuthorName:  d.account.AuthorName,
		AuthorURL:   d.account.AuthorURL,

		Title:   params,
		Content: nodes,
	})
	if err != nil {
		return
	}

	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{
				Flags: rt.SpanFlag_PlainText,
				Text:  "The post for your session around ",
			},
			{
				Flags: rt.SpanFlag_Bold,
				Text:  params,
			},
			{
				Flags: rt.SpanFlag_PlainText,
				Text:  " created ",
			},
			{
				Flags: rt.SpanFlag_URL,
				Text:  "here",
				URL:   d.page.URL,
			},
		},
	})

	return
}

// AppendToExisting implements publisher.Interface
func (d *Driver) AppendToExisting(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	var rd strings.Reader
	rd.Reset(in.Data.Get())
	nodes, err := htmlToNodes(&rd)
	if err != nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.account.AccessToken) == 0 {
		err = fmt.Errorf("account not set")
		return
	}

	if len(d.page.Path) == 0 {
		err = fmt.Errorf("page not set")
		return
	}

	// backup old page content
	buf := make([]telegraphNode, len(d.page.Content)+len(nodes))
	copy(buf, d.page.Content)
	copy(buf[len(buf)-len(nodes):], nodes)

	updatedPage, err := d.client.EditPage(d.page.Path, createPageOptions{
		AccessToken: d.account.AccessToken,
		Title:       d.page.Title,
		AuthorName:  d.page.AuthorName,
		AuthorURL:   d.page.AuthorURL,
		Content:     buf,
	})
	if err != nil {
		d.page.Content = buf
		return
	}

	d.page = updatedPage

	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{
				Flags: rt.SpanFlag_PlainText,
				Text:  "Your session around ",
			},
			{
				Flags: rt.SpanFlag_Bold,
				Text:  updatedPage.Title,
			},
			{
				Flags: rt.SpanFlag_PlainText,
				Text:  " has ended, view and edit your post ",
			},
			{
				Flags: rt.SpanFlag_URL,
				Text:  "here",
				URL:   updatedPage.URL,
			},
		},
	})

	return
}

func (d *Driver) RequireLogin(con rt.Conversation, cmd, params string) (rt.LoginFlow, error) {
	if len(params) == 0 {
		return rt.LoginFlow_Token, nil
	}

	user := User{
		authToken: params,
	}

	_, err := d.Login(con, &user)
	if err == nil {
		return rt.LoginFlow_None, nil
	}

	return rt.LoginFlow_None, err
}

func (d *Driver) Login(con rt.Conversation, user publisher.User) (out rt.PublisherOutput, err error) {
	u, ok := user.(*User)
	if !ok {
		err = fmt.Errorf("unexpected user type %T", user)
		return
	}

	var (
		account telegraphAccount
	)

	if len(u.authToken) != 0 { // login
		account, err = d.client.GetAccountInfo(u.authToken)
		if err == nil {
			account.AccessToken = u.authToken
		}
	} else { // no token, create a new account
		shortName := u.shortName
		if len(shortName) == 0 {
			shortName = d.defaultAccountShortName
		}
		account, err = d.client.CreateAccount(createAccountOptions{
			ShortName:  shortName,
			AuthorName: u.authorName,
			AuthorURL:  u.authorURL,
		})
	}
	if err != nil {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	d.account = account
	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{Text: "Here is your "},
			{Flags: rt.SpanFlag_Bold, Text: Name},
			{Text: " token, token, keep it for future use:\n\n"},
			{Flags: rt.SpanFlag_Code, Text: account.AccessToken},
		},
	})
	return
}

// RequestExternalAccess implements publisher.Interface
func (d *Driver) RequestExternalAccess(con rt.Conversation) (out rt.PublisherOutput, err error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if len(d.account.AuthURL) == 0 {
		err = fmt.Errorf("account not set")
		return
	}

	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{Text: "Click this "},
			{
				Flags: rt.SpanFlag_URL,
				Text:  "link",
				URL:   d.account.AuthURL,
			},
			{Text: " to get authorized"},
		},
	})

	return
}

func (d *Driver) Retrieve(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	u, err := urlpkg.Parse(params)
	if err != nil {
		return
	}

	path := strings.TrimLeft(u.Path, "/")

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.account.AccessToken) == 0 {
		err = fmt.Errorf("account not created")
		return
	}

	const LIMIT int64 = 20
	max := LIMIT
	for offset := int64(0); offset < max; offset += LIMIT {
		const URL = apiBaseURL + "getPageList"

		var list telegraphPageList
		list, err = d.client.GetPageList(getPageListOptions{
			AccessToken: d.account.AccessToken,
			Offset:      offset,
			Limit:       LIMIT,
		})
		if err != nil {
			err = fmt.Errorf("get page list: %w", err)
			return
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			if path == p.Path {
				var page telegraphPage
				page, err = d.client.GetPage(p.Path)
				if err != nil {
					return
				}

				d.page = page

				out.SendMessage.Set(rt.SendMessageOptions{
					Body: []rt.Span{
						{
							Flags: rt.SpanFlag_PlainText,
							Text:  "You can continue your session now.",
						},
					},
				})

				return
			}
		}
	}

	err = fmt.Errorf("not found")
	return
}

// List all posts for this user
func (d *Driver) List(con rt.Conversation) (out rt.PublisherOutput, err error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.account.AccessToken) == 0 {
		err = fmt.Errorf("account not set")
		return
	}

	const LIMIT int64 = 20
	max := LIMIT
	var result rt.SendMessageOptions
	for i := int64(0); i < max; i += LIMIT {
		var list telegraphPageList
		list, err = d.client.GetPageList(getPageListOptions{
			AccessToken: d.account.AccessToken,
			Offset:      i,
			Limit:       LIMIT,
		})
		if err != nil {
			err = fmt.Errorf("failed to get page list: %w", err)
			return
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			result.Body = append(result.Body,
				rt.Span{Text: "- "},
				rt.Span{
					Flags: rt.SpanFlag_URL,
					Text:  p.Title,
					URL:   p.URL,
				},
				rt.Span{Text: "\n"},
			)
		}
	}

	for i, j := 1, len(result.Body)-2; i < j; i, j = i+3, j-3 {
		result.Body[i], result.Body[j] = result.Body[j], result.Body[i]
	}

	out.SendMessage.Set(result)
	return
}

// Delete one post according to the url, however we cannot delete telegraph posts
// we just make it empty
func (d *Driver) Delete(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	if len(params) == 0 {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	if len(d.account.AccessToken) == 0 {
		err = fmt.Errorf("account not set")
		return
	}

	const LIMIT int64 = 20
	max := LIMIT

	// TODO: fix url matching
	toDelete := make(map[string]struct{})
	for offset := 0; ; offset++ {
		i := strings.IndexAny(params[offset:], " \t\r\n")
		if i == -1 {
			break
		}

		if i == 0 {
			continue
		}

		url, err := urlpkg.Parse(params[offset:i])
		if err == nil {
			toDelete[url.Path] = struct{}{}
		} else {
			toDelete[params[offset:i]] = struct{}{}
		}
	}

	var list telegraphPageList
	for i := int64(0); i < max; i += LIMIT {
		list, err = d.client.GetPageList(getPageListOptions{
			AccessToken: d.account.AccessToken,
			Offset:      i,
			Limit:       LIMIT,
		})
		if err != nil {
			err = fmt.Errorf("get page list: %w", err)
			return
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			_, ok := toDelete[p.Path]
			if !ok {
				continue
			}

			_, err = d.client.EditPage(p.Path, createPageOptions{
				AccessToken: d.account.AccessToken,
				Title:       "[Removed]",
				AuthorName:  "",
				AuthorURL:   "",

				Content: []telegraphNode{{
					Elm: rt.NewOptionalValue(telegraphNodeElement{
						Tag: "p",
						Children: []telegraphNode{{
							Text: "[Content Removed]",
						}},
					}),
				}},
			})
			if err != nil {
				err = fmt.Errorf("remove post content: %w", err)
				return
			}
		}
	}

	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{Text: "Posts deleted."},
		},
	})

	return
}
