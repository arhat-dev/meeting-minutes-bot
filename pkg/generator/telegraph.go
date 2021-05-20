package generator

import (
	"fmt"
	"sync"

	"gitlab.com/toby3d/telegraph"
)

const DefaultTelegraphAccountShortName = "meeting-minutes-bot"

type Telegraph struct {
	account *telegraph.Account
	page    *telegraph.Page

	mu *sync.RWMutex
}

type TelegraphLoginConfig struct {
	ShortName  string
	AuthorName string
	AuthorURL  string
}

func NewTelegraph() (*Telegraph, error) {
	return &Telegraph{
		mu: &sync.RWMutex{},
	}, nil
}

func (t *Telegraph) Login(config interface{}) error {
	baseAccount := &telegraph.Account{
		ShortName: DefaultTelegraphAccountShortName,
	}

	cfg, ok := config.(*TelegraphLoginConfig)
	if ok {
		baseAccount = &telegraph.Account{
			ShortName:  DefaultTelegraphAccountShortName,
			AuthorName: cfg.AuthorName,
			AuthURL:    cfg.AuthorURL,
		}
		if len(cfg.ShortName) != 0 {
			baseAccount.ShortName = cfg.ShortName
		}
	}

	account, err := telegraph.CreateAccount(*baseAccount)
	if err != nil {
		return err
	}

	t.mu.Lock()
	defer t.mu.Unlock()

	t.account = account
	return nil
}

func (t *Telegraph) Retrieve(url string) (title string, _ error) {
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
				page := p
				t.page = &page

				return p.Title, nil
			}
		}
	}

	return "", fmt.Errorf("not found")
}

func (t *Telegraph) Publish(title string, htmlContent []byte) (url string, _ error) {
	content, err := telegraph.ContentFormat(htmlContent)
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

func (t *Telegraph) Append(title string, htmlContent []byte) (url string, _ error) {
	content, err := telegraph.ContentFormat(htmlContent)
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
