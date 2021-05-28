package http

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"text/template"
	"time"

	"arhat.dev/pkg/tlshelper"
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
		func(config interface{}) (publisher.Interface, publisher.UserConfig, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, nil, fmt.Errorf("unexpected non http config")
			}

			_, err := url.Parse(c.URL)
			if err != nil {
				return nil, nil, fmt.Errorf("invalid url: %w", err)
			}

			var respTpl *template.Template
			if len(c.ResponseTemplate) != 0 {
				respTpl, err = template.New("").Parse(c.ResponseTemplate)
				if err != nil {
					return nil, nil, fmt.Errorf("invalid template: %w", err)
				}
			}

			return &Driver{
				method: strings.ToUpper(c.Method),
				url:    c.URL,

				respTpl: respTpl,
			}, &userConfig{}, nil
		},
		func() interface{} {
			return &Config{}
		},
	)
}

type nameValuePair struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

type Config struct {
	URL string              `json:"url" yaml:"url"`
	TLS tlshelper.TLSConfig `json:"tls" yaml:"tls"`

	Method  string          `json:"method" yaml:"method"`
	Headers []nameValuePair `json:"headers" yaml:"headers"`

	ResponseTemplate string `json:"responseTemplate" yaml:"responseTemplate"`
}

var _ publisher.UserConfig = (*userConfig)(nil)

type userConfig struct{}

func (u *userConfig) SetAuthToken(token string) {}

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	method string // upper case
	url    string

	respTpl *template.Template
}

func (d *Driver) Name() string                                              { return Name }
func (d *Driver) RequireLogin() bool                                        { return false }
func (d *Driver) Login(config publisher.UserConfig) (token string, _ error) { return "", nil }
func (d *Driver) AuthURL() (string, error)                                  { return "", nil }
func (d *Driver) Retrieve(key string) error                                 { return nil }
func (d *Driver) List() ([]publisher.PostInfo, error)                       { return nil, nil }
func (d *Driver) Delete(urls ...string) error                               { return nil }

func (d *Driver) Append(yamlSpec []byte) ([]message.Entity, error) {
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

	var (
		body io.Reader
	)
	switch d.method {
	case http.MethodGet:
	case http.MethodHead:
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		body = bytes.NewReader([]byte(spec.Body))
	case http.MethodDelete:
	case http.MethodOptions:
	}

	u, err := url.Parse(d.url)
	if err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	query := u.Query()
	for _, kv := range spec.QueryParams {
		query.Add(kv.Name, kv.Value)
	}

	u.RawQuery = query.Encode()

	req, err := http.NewRequest(d.method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to do http request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if d.respTpl != nil {
		buf := &bytes.Buffer{}
		err = d.respTpl.Execute(buf, string(data))
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
			Text: "You are going to interact with ",
		},
		{
			Kind: message.KindPre,
			Text: d.url,
		},
		{
			Kind: message.KindText,
			Text: fmt.Sprintf("using http.%q", strings.ToUpper(d.method)),
		},
	}, nil
}
