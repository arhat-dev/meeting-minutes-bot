package http

import (
	"fmt"
	"strings"
	"text/template"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/pkg/textquery"
	"arhat.dev/pkg/tlshelper"
	"arhat.dev/rs"
	"github.com/Masterminds/sprig/v3"
)

// nolint:revive
const (
	Name = "http"
)

func init() {
	publisher.Register(
		Name,
		func() publisher.Config { return &Config{} },
	)
}

type nameValuePair struct {
	rs.BaseField

	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type nameValueTemplatePair struct {
	nameTpl  *template.Template
	valueTpl *template.Template
}

func (p *nameValueTemplatePair) render(data interface{}) (name, value string, err error) {
	var (
		buf strings.Builder
	)

	err = p.nameTpl.Execute(&buf, data)
	if err != nil {
		return "", "", err
	}
	name = buf.String()

	buf.Reset()
	err = p.valueTpl.Execute(&buf, data)
	if err != nil {
		return "", "", err
	}
	value = buf.String()

	return
}

type Config struct {
	rs.BaseField

	URL string              `yaml:"url"`
	TLS tlshelper.TLSConfig `yaml:"tls"`

	Method  string          `yaml:"method"`
	Headers []nameValuePair `yaml:"headers"`

	ResponseTemplate string `yaml:"responseTemplate"`
}

func (c *Config) Create() (publisher.Interface, publisher.User, error) {
	urlTpl, err := parseTemplate(c.URL)
	if err != nil {
		return nil, nil, fmt.Errorf("invalud url template %q: %w", c.URL, err)
	}

	methodTpl, err := parseTemplate(c.Method)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid method template %q: %w", c.Method, err)
	}

	var respTpl *template.Template
	if len(c.ResponseTemplate) != 0 {
		respTpl, err = template.New("").
			Funcs(sprig.TxtFuncMap()).
			Funcs(map[string]interface{}{
				"jq":      textquery.JQ[byte, string],
				"jqBytes": textquery.JQ[byte, []byte],
			}).
			Parse(c.ResponseTemplate)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid template: %w", err)
		}
	}

	var headers []nameValueTemplatePair
	for _, p := range c.Headers {
		nameTpl, err2 := parseTemplate(p.Name)
		if err2 != nil {
			return nil, nil, fmt.Errorf("invalid header name template %q: %w", p.Name, err2)
		}

		valueTpl, err2 := parseTemplate(p.Value)
		if err2 != nil {
			return nil, nil, fmt.Errorf("invalid header value template %q: %w", p.Value, err2)
		}

		headers = append(headers, nameValueTemplatePair{
			nameTpl:  nameTpl,
			valueTpl: valueTpl,
		})
	}

	return &Driver{
		methodTpl: methodTpl,
		urlTpl:    urlTpl,

		headers: headers,
		respTpl: respTpl,
	}, &User{}, nil
}

var _ publisher.User = (*User)(nil)

type User struct{}

// SetPassword implements publisher.User
func (*User) SetPassword(string) {}

// SetTOTPCode implements publisher.User
func (*User) SetTOTPCode(string) {}

// SetToken implements publisher.User
func (*User) SetToken(string) {}

// SetUsername implements publisher.User
func (*User) SetUsername(string) {}
