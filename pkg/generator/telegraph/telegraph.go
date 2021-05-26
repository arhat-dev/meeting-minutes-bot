package telegraph

import (
	"bytes"
	"fmt"
	"html/template"
	"strings"
	"sync"

	"gitlab.com/toby3d/telegraph"

	"arhat.dev/meeting-minutes-bot/pkg/generator"

	_ "embed"
)

// nolint:revive
const (
	Name = "telegraph"
)

var (
	//go:embed page.tpl
	defaultPageTemplate string

	//go:embed page-prefix.tpl
	defaultPagePrefixTemplate string
)

func init() {
	generator.Register(
		Name,
		func(config interface{}) (generator.Interface, generator.UserConfig, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, nil, fmt.Errorf("unexpected non telegraph config: %T", config)
			}

			// TODO: move message template and page template to user config
			// 		 need to find a better UX for template editing

			pageTpl := c.PageTemplate
			if len(strings.TrimSpace(pageTpl)) == 0 {
				pageTpl = defaultPageTemplate
			}

			_, err := template.New("").
				Funcs(template.FuncMap(generator.CreateFuncMap(nil))).Parse(pageTpl)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid page template: %w", err)
			}

			pagePrefixTpl := c.PagePrefixTemplate
			if len(strings.TrimSpace(pagePrefixTpl)) == 0 {
				pagePrefixTpl = defaultPagePrefixTemplate
			}

			_, err = template.New("").
				Funcs(template.FuncMap(generator.CreateFuncMap(nil))).Parse(pagePrefixTpl)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid page prefix template: %w", err)
			}

			return &Telegraph{
				pageTpl:       pageTpl,
				pagePrefixTpl: pagePrefixTpl,

				defaultAccountShortName: c.DefaultAccountShortName,

				mu: &sync.RWMutex{},
			}, &userConfig{}, nil
		},
		func() interface{} {
			return &Config{
				PageTemplate: defaultPageTemplate,

				DefaultAccountShortName: "meeting-minutes-bot",
			}
		},
	)
}

type Config struct {
	PageTemplate       string `json:"pageTemplate" yaml:"pageTemplate"`
	PagePrefixTemplate string `json:"pagePrefixTemplate" yaml:"pagePrefixTemplate"`

	DefaultAccountShortName string `json:"defaultAccountShortName" yaml:"defaultAccountShortName"`
}

var _ generator.Interface = (*Telegraph)(nil)

type Telegraph struct {
	pageTpl, pagePrefixTpl string

	defaultAccountShortName string

	account *telegraph.Account
	page    *telegraph.Page

	mu *sync.RWMutex
}

var _ generator.UserConfig = (*userConfig)(nil)

type userConfig struct {
	shortName  string
	authorName string
	authorURL  string

	authToken string
}

func (c *userConfig) SetAuthToken(token string) {
	c.authToken = token
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

func (t *Telegraph) FormatPagePrefix() ([]byte, error) {
	pagePrefixTpl, err := template.New("").
		Funcs(template.FuncMap(generator.CreateFuncMap(nil))).Parse(t.pagePrefixTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse page prefix template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = pagePrefixTpl.Execute(buf, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute page prefix template: %w", err)
	}

	return buf.Bytes(), nil
}

func (t *Telegraph) FormatPageContent(
	messages []generator.Message, fm generator.FuncMap,
) ([]byte, error) {
	var (
		buf = &bytes.Buffer{}
		err error
	)

	pageTpl, err := template.New("").
		Funcs(template.FuncMap(generator.CreateFuncMap(nil))).Parse(t.pageTpl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse page template: %w", err)
	}

	buf.Reset()
	err = pageTpl.Execute(buf, &generator.TemplateData{
		Messages: messages,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute page template: %w", err)
	}

	return buf.Bytes(), nil
}
