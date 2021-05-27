package telegraph

import (
	"fmt"
	"sync"

	"gitlab.com/toby3d/telegraph"

	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

// nolint:revive
const (
	Name = "telegraph"
)

func init() {
	publisher.Register(
		Name,
		func(config interface{}) (publisher.Interface, publisher.UserConfig, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, nil, fmt.Errorf("unexpected non %s config: %T", Name, config)
			}

			return &Driver{
				defaultAccountShortName: c.DefaultAccountShortName,

				mu: &sync.RWMutex{},
			}, &userConfig{}, nil
		},
		func() interface{} {
			return &Config{
				DefaultAccountShortName: "meeting-minutes-bot",
			}
		},
	)
}

type Config struct {
	DefaultAccountShortName string `json:"defaultAccountShortName" yaml:"defaultAccountShortName"`
}

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	defaultAccountShortName string

	account *telegraph.Account
	page    *telegraph.Page

	mu *sync.RWMutex
}

func (t *Driver) Name() string {
	return Name
}

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

func (t *Driver) Retrieve(url string) (title string, _ error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.account == nil {
		return "", fmt.Errorf("account not created")
	}

	const limit = 20
	max := limit
	for i := 0; i < max; i += limit {
		list, err := t.account.GetPageList(i, limit)
		if err != nil {
			return "", fmt.Errorf("failed to get page list: %w", err)
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			if p.URL == url {
				page, err2 := telegraph.GetPage(p.Path, true)
				if err2 != nil {
					return "", err2
				}

				t.page = page
				return page.Title, nil
			}
		}
	}

	return "", fmt.Errorf("not found")
}

func (t *Driver) Publish(title string, body []byte) (url string, _ error) {
	content, err := telegraph.ContentFormat(body)
	if err != nil {
		return "", err
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if t.account == nil {
		return "", fmt.Errorf("account not created")
	}

	page, err := t.account.CreatePage(telegraph.Page{
		Title:   title,
		Content: content,
	}, true)
	if err != nil {
		return "", err
	}

	t.page = page

	return page.URL, nil
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

func (t *Driver) Append(title string, body []byte) (url string, _ error) {
	content, err := telegraph.ContentFormat(body)
	if err != nil {
		return "", err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.account == nil {
		return "", fmt.Errorf("account not created")
	}

	if t.page == nil {
		return "", fmt.Errorf("page not created")
	}

	// backup old page content
	prevContent := make([]telegraph.Node, len(t.page.Content))
	prevContent = append(prevContent, t.page.Content...)

	t.page.Content = append(t.page.Content, content...)
	updatedPage, err := t.account.EditPage(*t.page, true)
	if err != nil {
		t.page.Content = prevContent
		return "", err
	}

	t.page = updatedPage
	return updatedPage.URL, nil
}