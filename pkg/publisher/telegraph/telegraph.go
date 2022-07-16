package telegraph

import (
	"context"
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

	mu *sync.RWMutex
}

func (t *Driver) RequireLogin() bool { return true }

func (t *Driver) Login(config publisher.UserConfig) (string, error) {
	cfg, ok := config.(*userConfig)
	if !ok {
		return "", fmt.Errorf("unexpected user type %T", config)
	}

	var (
		account telegraphAccount
		err     error
	)

	if len(cfg.authToken) != 0 { // login
		const URL = apiBaseURL + "getAccountInfo"

		account, err = t.client.GetAccountInfo(cfg.authToken)
		if err == nil {
			account.AccessToken = cfg.authToken
		}
	} else { // no token, create a new account
		const URL = apiBaseURL + "createAccount"

		shortName := cfg.shortName
		if len(shortName) == 0 {
			shortName = t.defaultAccountShortName
		}
		account, err = t.client.CreateAccount(createAccountOptions{
			ShortName:  shortName,
			AuthorName: cfg.authorName,
			AuthorURL:  cfg.authorURL,
		})
	}
	if err != nil {
		return "", err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.account = account
	return account.AccessToken, nil
}

func (t *Driver) AuthURL() (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(t.account.AuthURL) == 0 {
		return "", fmt.Errorf("account not set")
	}

	return t.account.AuthURL, nil
}

func (t *Driver) Retrieve(url string) (_ []rt.Span, err error) {
	u, err := urlpkg.Parse(url)
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

func (t *Driver) Publish(title string, body *rt.Input) (_ []rt.Span, err error) {
	nodes, err := htmlToNodes(body.Reader())
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

		Title:   title,
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
			Text:  title,
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

// List all posts for this user
func (t *Driver) List() ([]publisher.PostInfo, error) {
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
func (t *Driver) Delete(urls ...string) error {
	if len(urls) == 0 {
		return nil
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.account.AccessToken) == 0 {
		return fmt.Errorf("account not set")
	}

	const LIMIT int64 = 20
	max := LIMIT

	toDelete := make(map[string]struct{})
	for _, u := range urls {
		toDelete[u] = struct{}{}
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
			_, ok := toDelete[p.URL]
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

func (t *Driver) Append(ctx context.Context, body *rt.Input) (_ []rt.Span, err error) {
	nodes, err := htmlToNodes(body.Reader())
	if err != nil {
		return nil, err
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
