package generator

import (
	"fmt"
	"html"
	"sync"

	"gitlab.com/toby3d/telegraph"
)

const DefaultTelegraphAccountShortName = "meeting-minutes-bot"

var _ Interface = (*Telegraph)(nil)

type Telegraph struct {
	account *telegraph.Account
	page    *telegraph.Page

	mu *sync.RWMutex
}

type TelegraphLoginConfig struct {
	ShortName  string
	AuthorName string
	AuthorURL  string

	AuthToken string
}

func NewTelegraph() (*Telegraph, error) {
	return &Telegraph{
		mu: &sync.RWMutex{},
	}, nil
}

func (t *Telegraph) Name() string {
	return "Telegraph"
}

func (t *Telegraph) Login(config interface{}) (string, error) {
	baseAccount := &telegraph.Account{
		ShortName: DefaultTelegraphAccountShortName,
	}

	cfg, ok := config.(*TelegraphLoginConfig)
	if ok {
		baseAccount = &telegraph.Account{
			ShortName:  DefaultTelegraphAccountShortName,
			AuthorName: cfg.AuthorName,
			AuthURL:    cfg.AuthorURL,

			AccessToken: cfg.AuthToken,
		}
		if len(cfg.ShortName) != 0 {
			baseAccount.ShortName = cfg.ShortName
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

func (t *Telegraph) AuthURL() (string, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if t.account == nil {
		return "", fmt.Errorf("account not created")
	}

	return t.account.AuthURL, nil
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

func (t *Telegraph) Publish(title string, body []byte) (url string, _ error) {
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

func (t *Telegraph) Append(title string, body []byte) (url string, _ error) {
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

func (t *Telegraph) Format(kind FormatKind, data string, params ...string) string {
	switch kind {
	case KindText:
		return html.EscapeString(data)
	case KindBold:
		return `<strong>` + html.EscapeString(data) + `</strong>`
	case KindItalic:
		return `<em>` + html.EscapeString(data) + `</em>`
	case KindStrikethrough:
		return `<del>` + html.EscapeString(data) + `</del>`
	case KindUnderline:
		return `<u>` + html.EscapeString(data) + `</u>`
	case KindPre:
		return `<pre>` + html.EscapeString(data) + `</pre>`
	case KindCode:
		return `<code>` + data + `</code>`
	case KindNewLine:
		return data + `<br>`
	case KindParagraph:
		return `<p>` + data + `</p>`
	case KindThematicBreak:
		return data + `<hr>`
	case KindBlockquote:
		return `<blockquote>` + data + `</blockquote>`
	case KindEmail:
		return fmt.Sprintf(`<a href="mailto:%s">`, data) + html.EscapeString(data) + `</a>`
	case KindPhoneNumber:
		return fmt.Sprintf(`<a href="tel:%s">`, data) + html.EscapeString(data) + `</a>`
	case KindURL:
		// TODO: parse telegraph supported media url
		url := data
		if len(params) == 1 {
			url = params[0]
		}

		return fmt.Sprintf(`<a href="%s">`, url) + html.EscapeString(data) + `</a>`
	default:
		return html.EscapeString(data)
	}
}
