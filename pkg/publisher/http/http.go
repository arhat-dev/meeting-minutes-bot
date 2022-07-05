package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"text/template"
	"time"

	"arhat.dev/pkg/textquery"
	"arhat.dev/pkg/tlshelper"
	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
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
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type nameValueTemplatePair struct {
	nameTpl  *template.Template
	valueTpl *template.Template
}

func (p *nameValueTemplatePair) render(data interface{}) (name, value string, err error) {
	buf := &bytes.Buffer{}
	err = p.nameTpl.Execute(buf, data)
	if err != nil {
		return "", "", err
	}
	name = buf.String()

	buf.Reset()
	err = p.valueTpl.Execute(buf, data)
	if err != nil {
		return "", "", err
	}
	value = buf.String()

	return name, value, nil
}

type Config struct {
	URL string              `json:"url" yaml:"url"`
	TLS tlshelper.TLSConfig `json:"tls" yaml:"tls"`

	Method  string          `json:"method" yaml:"method"`
	Headers []nameValuePair `json:"headers" yaml:"headers"`

	ResponseTemplate string `json:"responseTemplate" yaml:"responseTemplate"`
}

func (c *Config) Create() (publisher.Interface, publisher.UserConfig, error) {
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
	}, &userConfig{}, nil
}

var _ publisher.UserConfig = (*userConfig)(nil)

type userConfig struct{}

func (u *userConfig) SetAuthToken(token string) {}

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	methodTpl *template.Template
	urlTpl    *template.Template

	headers []nameValueTemplatePair
	respTpl *template.Template
}

func (d *Driver) Name() string { return Name }

func (d *Driver) RequireLogin() bool { return false }

func (d *Driver) Login(config publisher.UserConfig) (token string, _ error) {
	return "", fmt.Errorf("unimplemented")
}

func (d *Driver) AuthURL() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (d *Driver) Retrieve(key string) ([]message.Entity, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) List() ([]publisher.PostInfo, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) Delete(urls ...string) error {
	return fmt.Errorf("unimplemented")
}

func (d *Driver) Append(ctx context.Context, yamlSpec []byte) ([]message.Entity, error) {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	spec := &Spec{}
	err := yaml.Unmarshal(yamlSpec, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse request spec: %w", err)
	}

	methodBytes, err := executeTemplate(d.methodTpl, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to execute method template: %w", err)
	}
	method := strings.ToUpper(string(methodBytes))

	urlBytes, err := executeTemplate(d.urlTpl, spec)
	if err != nil {
		return nil, fmt.Errorf("failed to execute url template: %w", err)
	}

	var (
		body io.Reader
	)
	switch method {
	case http.MethodGet:
	case http.MethodHead:
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		body = bytes.NewReader([]byte(spec.Body))
	case http.MethodDelete:
	case http.MethodOptions:
	}

	req, err := http.NewRequest(method, string(urlBytes), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	for _, h := range d.headers {
		name, value, err2 := h.render(spec)
		if err2 != nil {
			return nil, fmt.Errorf("failed to render headers: %w", err2)
		}

		req.Header.Add(name, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if d.respTpl != nil {
		buf := &bytes.Buffer{}

		tplData := &responseTemplateData{
			Code:    resp.StatusCode,
			Headers: resp.Header,
			Body:    data,
			Request: responseTemplateRequestData{
				URL:     req.URL,
				Headers: req.Header,
			},
		}

		err = d.respTpl.Execute(buf, tplData)
		if err != nil {
			return nil, err
		}

		return []message.Entity{
			{
				Kind: message.KindPre,
				Text: buf.String(),
			},
		}, nil
	}

	return []message.Entity{
		{
			Kind: message.KindPre,
			Text: string(data),
		},
	}, nil
}

func (d *Driver) Publish(title string, body []byte) ([]message.Entity, error) {
	return []message.Entity{
		{
			Kind: message.KindText,
			Text: "HTTP publisher ready",
		},
	}, nil
}

func executeTemplate(tpl *template.Template, data interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := tpl.Execute(buf, data)
	return buf.Bytes(), err
}

func parseTemplate(text string) (*template.Template, error) {
	return template.New("").
		Funcs(sprig.TxtFuncMap()).
		Funcs(map[string]interface{}{
			"jq":      textquery.JQ[byte, string],
			"jqBytes": textquery.JQ[byte, []byte],
		}).
		Parse(text)
}
