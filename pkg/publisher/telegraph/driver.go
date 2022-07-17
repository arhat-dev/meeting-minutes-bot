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
func (t *Driver) CreateNew(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) (_ []rt.Span, err error) {
	nodes, err := htmlToNodes(fromGenerator.Reader())
	if err != nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if len(t.account.AccessToken) == 0 {
		return nil, fmt.Errorf("account not set")
	}

	t.page, err = t.client.CreatePage(createPageOptions{
		AccessToken: t.account.AccessToken,
		AuthorName:  t.account.AuthorName,
		AuthorURL:   t.account.AuthorURL,

		Title:   params,
		Content: nodes,
	})
	if err != nil {
		return nil, err
	}

	return []rt.Span{
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
			URL:   t.page.URL,
		},
	}, nil
}

// AppendToExisting implements publisher.Interface
func (t *Driver) AppendToExisting(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) (_ []rt.Span, err error) {
	nodes, err := htmlToNodes(fromGenerator.Reader())
	if err != nil {
		return
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.account.AccessToken) == 0 {
		return nil, fmt.Errorf("account not set")
	}

	if len(t.page.Path) == 0 {
		return nil, fmt.Errorf("page not set")
	}

	// backup old page content
	buf := make([]telegraphNode, len(t.page.Content)+len(nodes))
	copy(buf, t.page.Content)
	copy(buf[len(buf)-len(nodes):], nodes)

	updatedPage, err := t.client.EditPage(t.page.Path, createPageOptions{
		AccessToken: t.account.AccessToken,
		Title:       t.page.Title,
		AuthorName:  t.page.AuthorName,
		AuthorURL:   t.page.AuthorURL,
		Content:     buf,
	})
	if err != nil {
		t.page.Content = buf
		return nil, err
	}

	t.page = updatedPage

	return []rt.Span{
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
	}, nil
}

func (t *Driver) RequireLogin(con rt.Conversation, cmd, params string) (rt.LoginFlow, error) {
	if len(params) == 0 {
		return rt.LoginFlow_Token, nil
	}

	user := userConfig{
		authToken: params,
	}

	_, err := t.Login(con, &user)
	if err == nil {
		return rt.LoginFlow_None, nil
	}

	return rt.LoginFlow_None, err
}

func (t *Driver) Login(con rt.Conversation, config publisher.User) ([]rt.Span, error) {
	user, ok := config.(*userConfig)
	if !ok {
		return nil, fmt.Errorf("unexpected user type %T", config)
	}

	var (
		account telegraphAccount
		err     error
	)

	if len(user.authToken) != 0 { // login
		account, err = t.client.GetAccountInfo(user.authToken)
		if err == nil {
			account.AccessToken = user.authToken
		}
	} else { // no token, create a new account
		shortName := user.shortName
		if len(shortName) == 0 {
			shortName = t.defaultAccountShortName
		}
		account, err = t.client.CreateAccount(createAccountOptions{
			ShortName:  shortName,
			AuthorName: user.authorName,
			AuthorURL:  user.authorURL,
		})
	}
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.account = account
	return []rt.Span{
		{Text: "Here is your "},
		{Flags: rt.SpanFlag_Bold, Text: Name},
		{Text: " token, token, keep it for future use:\n\n"},
		{Flags: rt.SpanFlag_Code, Text: account.AccessToken},
	}, nil
}

// RequestExternalAccess implements publisher.Interface
func (t *Driver) RequestExternalAccess(con rt.Conversation) ([]rt.Span, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.account.AuthURL) == 0 {
		return nil, fmt.Errorf("account not set")
	}

	return []rt.Span{
		{Text: "Click this "},
		{
			Flags: rt.SpanFlag_URL,
			Text:  "link",
			URL:   t.account.AuthURL,
		},
		{Text: " to get authorized"},
	}, nil
}

func (t *Driver) Retrieve(con rt.Conversation, cmd, params string) (_ []rt.Span, err error) {
	u, err := urlpkg.Parse(params)
	if err != nil {
		return
	}

	path := strings.TrimLeft(u.Path, "/")

	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.account.AccessToken) == 0 {
		return nil, fmt.Errorf("account not created")
	}

	const LIMIT int64 = 20
	max := LIMIT
	for offset := int64(0); offset < max; offset += LIMIT {
		const URL = apiBaseURL + "getPageList"

		list, err := t.client.GetPageList(getPageListOptions{
			AccessToken: t.account.AccessToken,
			Offset:      offset,
			Limit:       LIMIT,
		})
		if err != nil {
			return nil, fmt.Errorf("get page list: %w", err)
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			if path == p.Path {
				page, err2 := t.client.GetPage(p.Path)
				if err2 != nil {
					return nil, err2
				}

				t.page = page
				return []rt.Span{
					{
						Flags: rt.SpanFlag_PlainText,
						Text:  "You can continue your session now.",
					},
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("not found")
}

// List all posts for this user
func (t *Driver) List(con rt.Conversation) ([]publisher.PostInfo, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.account.AccessToken) == 0 {
		return nil, fmt.Errorf("account not set")
	}

	const LIMIT int64 = 20
	max := LIMIT
	var result []publisher.PostInfo
	for i := int64(0); i < max; i += LIMIT {
		list, err := t.client.GetPageList(getPageListOptions{
			AccessToken: t.account.AccessToken,
			Offset:      i,
			Limit:       LIMIT,
		})
		if err != nil {
			return result, fmt.Errorf("failed to get page list: %w", err)
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			result = append(result, publisher.PostInfo{
				Title: p.Title,
				URL:   p.URL,
			})
		}
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result, nil
}

// Delete one post according to the url, however we cannot delete telegraph posts
// we just make it empty
func (t *Driver) Delete(con rt.Conversation, cmd, params string) error {
	if len(params) == 0 {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.account.AccessToken) == 0 {
		return fmt.Errorf("account not set")
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

	for i := int64(0); i < max; i += LIMIT {
		list, err := t.client.GetPageList(getPageListOptions{
			AccessToken: t.account.AccessToken,
			Offset:      i,
			Limit:       LIMIT,
		})
		if err != nil {
			return fmt.Errorf("failed to get page list: %w", err)
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			_, ok := toDelete[p.Path]
			if !ok {
				continue
			}

			_, err = t.client.EditPage(p.Path, createPageOptions{
				AccessToken: t.account.AccessToken,
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
				return fmt.Errorf("remove post content: %w", err)
			}
		}
	}

	return nil
}
