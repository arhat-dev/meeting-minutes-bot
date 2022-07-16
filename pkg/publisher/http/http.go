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

	"arhat.dev/pkg/stringhelper"
	"arhat.dev/pkg/textquery"
	"arhat.dev/pkg/tlshelper"
	"arhat.dev/rs"
	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
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

func (d *Driver) Retrieve(key string) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) List() ([]publisher.PostInfo, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) Delete(urls ...string) error {
	return fmt.Errorf("unimplemented")
}

func (d *Driver) Append(ctx context.Context, yamlSpec *rt.Input) (_ []rt.Span, err error) {
	content, err := yamlSpec.Bytes()
	if err != nil {
		return
	}

	client := http.Client{
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

	var spec Spec
	err = yaml.Unmarshal(content, &spec)
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
		br   bytes.Reader
	)
	switch method {
	case http.MethodGet:
	case http.MethodHead:
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		br.Reset(stringhelper.ToBytes[byte, byte](spec.Body))
		body = &br
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
		var buf strings.Builder

		tplData := responseTemplateData{
			Code:    resp.StatusCode,
			Headers: resp.Header,
			Body:    data,
			Request: responseTemplateRequestData{
				URL:     req.URL,
				Headers: req.Header,
			},
		}

		err = d.respTpl.Execute(&buf, &tplData)
		if err != nil {
			return nil, err
		}

		return []rt.Span{
			{
				Flags: rt.SpanFlag_Pre,
				Text:  buf.String(),
			},
		}, nil
	}

	return []rt.Span{
		{
			Flags: rt.SpanFlag_Pre,
			Text:  stringhelper.Convert[string, byte](data),
		},
	}, nil
}

func (d *Driver) Publish(title string, body *rt.Input) ([]rt.Span, error) {
	return []rt.Span{
		{
			Flags: rt.SpanFlag_PlainText,
			Text:  "HTTP publisher ready",
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
