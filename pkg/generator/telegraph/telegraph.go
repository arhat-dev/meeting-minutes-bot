package generator

import (
	"fmt"
	"html"
	"sync"

	"gitlab.com/toby3d/telegraph"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
)

// nolint:revive
const (
	Name = "telegraph"
)

func init() {
	generator.Register(
		Name,
		func(config interface{}) (generator.Interface, generator.UserConfig, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, nil, fmt.Errorf("Unexpected non telegraph config: %T", config)
			}

			return &Telegraph{
				defaultAccountShortName: c.DefaultAccountShortName,

				mu: &sync.RWMutex{},
			}, &UserConfig{}, nil
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

var _ generator.Interface = (*Telegraph)(nil)

type Telegraph struct {
	defaultAccountShortName string

	account *telegraph.Account
	page    *telegraph.Page

	mu *sync.RWMutex
}

var _ generator.UserConfig = (*UserConfig)(nil)

type UserConfig struct {
	ShortName  string
	AuthorName string
	AuthorURL  string

	AuthToken string
}

func (c *UserConfig) SetAuthToken(token string) {
	c.AuthToken = token
}

func NewTelegraph() (*Telegraph, error) {
	return &Telegraph{
		mu: &sync.RWMutex{},
	}, nil
}

func (t *Telegraph) Name() string {
	return Name
}

func (t *Telegraph) Login(config generator.UserConfig) (string, error) {
	baseAccount := &telegraph.Account{
		ShortName: t.defaultAccountShortName,
	}

	cfg, ok := config.(*UserConfig)
	if ok {
		baseAccount = &telegraph.Account{
			ShortName:  t.defaultAccountShortName,
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

func (t *Telegraph) Format(kind generator.FormatKind, data string, params ...string) string {
	switch kind {
	case generator.KindText:
		return html.EscapeString(data)
	case generator.KindBold:
		return `<strong>` + html.EscapeString(data) + `</strong>`
	case generator.KindItalic:
		return `<em>` + html.EscapeString(data) + `</em>`
	case generator.KindStrikethrough:
		return `<del>` + html.EscapeString(data) + `</del>`
	case generator.KindUnderline:
		return `<u>` + html.EscapeString(data) + `</u>`
	case generator.KindPre:
		return `<pre>` + html.EscapeString(data) + `</pre>`
	case generator.KindCode:
		return `<code>` + data + `</code>`
	case generator.KindNewLine:
		return data + `<br>`
	case generator.KindParagraph:
		return `<p>` + data + `</p>`
	case generator.KindThematicBreak:
		return data + `<hr>`
	case generator.KindBlockquote:
		return `<blockquote>` + data + `</blockquote>`
	case generator.KindEmail:
		return fmt.Sprintf(`<a href="mailto:%s">`, data) + html.EscapeString(data) + `</a>`
	case generator.KindPhoneNumber:
		return fmt.Sprintf(`<a href="tel:%s">`, data) + html.EscapeString(data) + `</a>`
	case generator.KindVideo, generator.KindAudio:
		result := `<figure>` + fmt.Sprintf(`<video src="%s"></video>`, data)
		caption := ""
		if len(params) != 0 {
			caption = params[0]
		}

		return result + fmt.Sprintf(`<figcaption>%s</figcaption></figure>`, html.EscapeString(caption))
	case generator.KindImage:
		result := `<figure>` + fmt.Sprintf(`<img src="%s">`, data)
		caption := ""
		if len(params) != 0 {
			caption = params[0]
		}

		return result + fmt.Sprintf(`<figcaption>%s</figcaption></figure>`, html.EscapeString(caption))
	case generator.KindURL:
		// TODO: parse telegraph supported media url (e.g. youtube)
		url := data
		if len(params) == 1 {
			url = params[0]
		}

		return fmt.Sprintf(`<a href="%s">`, url) + html.EscapeString(data) + `</a>`
	default:
		return html.EscapeString(data)
	}
}
