package telegraph

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sync"

	"github.com/Masterminds/sprig/v3"
	"gitlab.com/toby3d/telegraph"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

// nolint:revive
const (
	Name = "telegraph"
)

var (
	//go:embed templates/*
	defaultTemplates embed.FS
)

const (
	defaultTemplatesPattern = "templates/*.tpl"
)

func init() {
	_, err := template.New("").
		Funcs(sprig.HtmlFuncMap()).
		Funcs(template.FuncMap(generator.CreateFuncMap())).
		ParseFS(defaultTemplates, defaultTemplatesPattern)
	if err != nil {
		panic(fmt.Errorf("%s: default templates not valid: %w", Name, err))
	}

	generator.Register(
		Name,
		func(config interface{}) (generator.Interface, generator.UserConfig, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, nil, fmt.Errorf("unexpected non %s config: %T", Name, config)
			}

			// TODO: move message template and page template to user config
			// 		 need to find a better UX for template editing

			templatesFS := fs.FS(defaultTemplates)
			pattern := defaultTemplatesPattern
			if len(c.TemplatesDir) != 0 {
				templatesFS = os.DirFS(c.TemplatesDir)
				if len(c.TemplatesPattern) != 0 {
					pattern = c.TemplatesPattern
				} else {
					pattern = path.Join(filepath.Base(c.TemplatesDir), "*")
				}
			}

			ret := &Telegraph{
				defaultAccountShortName: c.DefaultAccountShortName,

				mu: &sync.RWMutex{},
			}

			var err error
			ret.templates, err = template.New("").
				Funcs(sprig.HtmlFuncMap()).
				Funcs(template.FuncMap(generator.CreateFuncMap())).
				ParseFS(templatesFS, pattern)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid templates: %w", err)
			}

			return ret, &userConfig{}, nil
		},
		func() interface{} {
			return &Config{
				TemplatesDir: "",

				DefaultAccountShortName: "meeting-minutes-bot",
			}
		},
	)
}

type Config struct {
	TemplatesDir     string `json:"templatesDir" yaml:"templatesDir"`
	TemplatesPattern string `json:"templatesPattern" yaml:"templatesPattern"`

	DefaultAccountShortName string `json:"defaultAccountShortName" yaml:"defaultAccountShortName"`
}

var _ generator.Interface = (*Telegraph)(nil)

type Telegraph struct {
	templates *template.Template

	defaultAccountShortName string

	account *telegraph.Account
	page    *telegraph.Page

	mu *sync.RWMutex
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

// List all posts for this user
func (t *Telegraph) List() ([]generator.PostInfo, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.account == nil {
		return nil, fmt.Errorf("account not created")
	}

	const limit = 20
	max := limit
	var result []generator.PostInfo
	for i := 0; i < max; i += limit {
		list, err := t.account.GetPageList(i, limit)
		if err != nil {
			return result, fmt.Errorf("failed to get page list: %w", err)
		}

		max = list.TotalCount

		for _, p := range list.Pages {
			result = append(result, generator.PostInfo{
				Title: p.Title,
				URL:   p.URL,
			})
		}
	}

	return result, nil
}

// Delete one post according to the url, however we cannot delete telegraph posts
// we just make it empty
func (t *Telegraph) Delete(urls ...string) error {
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

func (t *Telegraph) FormatPageHeader() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := t.templates.ExecuteTemplate(buf, "page.header", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute page header template: %w", err)
	}

	return buf.Bytes(), nil
}

func (t *Telegraph) FormatPageBody(
	messages []message.Interface,
) ([]byte, error) {
	var (
		buf = &bytes.Buffer{}
		err error
	)

	t.mu.Lock()
	err = t.templates.ExecuteTemplate(
		buf,
		"page.body",
		&generator.TemplateData{
			Messages: messages,
		},
	)
	t.mu.Unlock()

	if err != nil {
		return nil, fmt.Errorf("failed to execute page template: %w", err)
	}

	return buf.Bytes(), nil
}
