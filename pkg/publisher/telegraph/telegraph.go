package telegraph

import (
	"context"
	"fmt"
	"sync"

	"gitlab.com/toby3d/telegraph"

	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

// nolint:revive
const (
	Name = "telegraph"
)

func init() {
	publisher.Register(
		Name,
		func() publisher.Config {
			return &Config{
				DefaultAccountShortName: "meeting-minutes-bot",
			}
		},
	)
}

type Config struct {
	DefaultAccountShortName string `json:"defaultAccountShortName" yaml:"defaultAccountShortName"`
}

func (c *Config) Create() (publisher.Interface, publisher.UserConfig, error) {
	return &Driver{
		defaultAccountShortName: c.DefaultAccountShortName,

		mu: &sync.RWMutex{},
	}, &userConfig{}, nil
}

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	defaultAccountShortName string

	account *telegraph.Account
	page    *telegraph.Page

	mu *sync.RWMutex
}

func (t *Driver) Name() string       { return Name }
func (t *Driver) RequireLogin() bool { return true }

func (t *Driver) Login(config publisher.UserConfig) (string, error) {
	baseAccount := &telegraph.Account{
		ShortName: t.defaultAccountShortName,
	}

	cfg, ok := config.(*userConfig)
	if ok {
		baseAccount = &telegraph.Account{
			ShortName:  t.defaultAccountShortName,
			AuthorName: cfg.authorName,
			AuthURL:    cfg.authorURL,

			AccessToken: cfg.authToken,
		}
		if len(cfg.shortName) != 0 {
			baseAccount.ShortName = cfg.shortName
		}
	}

	var (
		account *telegraph.Account
		err     error
	)
	if len(baseAccount.AccessToken) != 0 {
		account, err = baseAccount.GetAccountInfo(
			telegraph.FieldAuthURL,
			telegraph.FieldAuthorName,
			telegraph.FieldAuthorURL,
			telegraph.FieldShortName,
			telegraph.FieldPageCount,
		)
		if err == nil {
			account.AccessToken = baseAccount.AccessToken
		}
	} else {
		account, err = telegraph.CreateAccount(*baseAccount)
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

	if t.account == nil {
		return "", fmt.Errorf("account not created")
	}

	return t.account.AuthURL, nil
}

func (t *Driver) Retrieve(url string) ([]message.Entity, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.account == nil {
		return nil, fmt.Errorf("account not created")
	}

	const limit = 20
	max := limit
	for i := 0; i < max; i += limit {
		list, err := t.account.GetPageList(i, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to get page list: %w", err)
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			if p.URL == url {
				page, err2 := telegraph.GetPage(p.Path, true)
				if err2 != nil {
					return nil, err2
				}

				t.page = page
				return []message.Entity{
					{
						Kind: message.KindText,
						Text: "You can continue your session now.",
					},
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("not found")
}

func (t *Driver) Publish(title string, body []byte) ([]message.Entity, error) {
	content, err := telegraph.ContentFormat(body)
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.account == nil {
		return nil, fmt.Errorf("account not created")
	}

	page, err := t.account.CreatePage(telegraph.Page{
		Title:   title,
		Content: content,
	}, true)
	if err != nil {
		return nil, err
	}

	t.page = page
	return []message.Entity{
		{
			Kind: message.KindText,
			Text: "The post for your session around ",
		},
		{
			Kind: message.KindBold,
			Text: fmt.Sprintf("%q", title),
		},
		{
			Kind: message.KindText,
			Text: " has been created ",
		},
		{
			Kind: message.KindURL,
			Text: "here",
			Params: map[message.EntityParamKey]interface{}{
				message.EntityParamURL: page.URL,
			},
		},
	}, nil
}

// List all posts for this user
func (t *Driver) List() ([]publisher.PostInfo, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.account == nil {
		return nil, fmt.Errorf("account not created")
	}

	const limit = 20
	max := limit
	var result []publisher.PostInfo
	for i := 0; i < max; i += limit {
		list, err := t.account.GetPageList(i, limit)
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

	if t.account == nil {
		return fmt.Errorf("account not created")
	}

	const limit = 20
	max := limit

	toDelete := make(map[string]struct{})
	for _, u := range urls {
		toDelete[u] = struct{}{}
	}

	for i := 0; i < max; i += limit {
		list, err := t.account.GetPageList(i, limit)
		if err != nil {
			return fmt.Errorf("failed to get page list: %w", err)
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			_, ok := toDelete[p.URL]
			if !ok {
				continue
			}

			p.AuthorName = ""
			p.AuthorURL = ""
			p.Description = ""
			p.ImageURL = ""
			p.Title = "[Removed]"

			p.Content, err = telegraph.ContentFormat("<p>[Content Removed]</p>")
			if err != nil {
				panic(err)
			}

			_, err2 := t.account.EditPage(p, false)
			if err2 != nil {
				return fmt.Errorf("failed to remove post content: %w", err2)
			}
		}
	}

	return nil
}

func (t *Driver) Append(ctx context.Context, body []byte) ([]message.Entity, error) {
	content, err := telegraph.ContentFormat(body)
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.account == nil {
		return nil, fmt.Errorf("account not created")
	}

	if t.page == nil {
		return nil, fmt.Errorf("page not created")
	}

	// backup old page content
	prevContent := make([]telegraph.Node, len(t.page.Content))
	prevContent = append(prevContent, t.page.Content...)

	t.page.Content = append(t.page.Content, content...)
	updatedPage, err := t.account.EditPage(*t.page, true)
	if err != nil {
		t.page.Content = prevContent
		return nil, err
	}

	t.page = updatedPage

	return []message.Entity{
		{
			Kind: message.KindText,
			Text: "Your session around ",
		},
		{
			Kind: message.KindBold,
			Text: fmt.Sprintf("%q", updatedPage.Title),
		},
		{
			Kind: message.KindText,
			Text: " has been ended, view and edit your post ",
		},
		{
			Kind: message.KindURL,
			Text: "here",
			Params: map[message.EntityParamKey]interface{}{
				message.EntityParamURL: updatedPage.URL,
			},
		},
	}, nil
}
