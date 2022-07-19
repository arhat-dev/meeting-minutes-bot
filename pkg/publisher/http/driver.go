package http

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"text/template"
	"time"

	"arhat.dev/pkg/stringhelper"
	"arhat.dev/pkg/textquery"
	"github.com/Masterminds/sprig/v3"
	"gopkg.in/yaml.v3"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	methodTpl *template.Template
	urlTpl    *template.Template

	headers []nameValueTemplatePair
	respTpl *template.Template
}

// CreateNew implements publisher.Interface
func (*Driver) CreateNew(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	return
}

// AppendToExisting implements publisher.Interface
//
// fromGenerator is expected to provide yaml spec of the http request
func (d *Driver) AppendToExisting(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
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
	err = yaml.Unmarshal(stringhelper.ToBytes[byte, byte](in.Data.Get()), &spec)
	if err != nil {
		err = fmt.Errorf("parse request spec: %w", err)
		return
	}

	method, err := executeTemplate(d.methodTpl, spec)
	if err != nil {
		err = fmt.Errorf("execute method template: %w", err)
		return
	}
	method = strings.ToUpper(method)

	url, err := executeTemplate(d.urlTpl, spec)
	if err != nil {
		err = fmt.Errorf("execute url template: %w", err)
		return
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

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		err = fmt.Errorf("create http request: %w", err)
		return
	}

	for _, h := range d.headers {
		name, value, err2 := h.render(spec)
		if err2 != nil {
			err = fmt.Errorf("render headers: %w", err2)
			return
		}

		req.Header.Add(name, value)
	}

	resp, err := client.Do(req)
	if err != nil {
		err = fmt.Errorf("do http request: %w", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		err = fmt.Errorf("read response body: %w", err)
		return
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
			return
		}

		out.SendMessage.Set(rt.SendMessageOptions{
			Body: []rt.Span{
				{
					Flags: rt.SpanFlag_Pre,
					Text:  buf.String(),
				},
			},
		})
		return
	}

	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{
				Flags: rt.SpanFlag_Pre,
				Text:  stringhelper.Convert[string, byte](data),
			},
		},
	})

	return
}

func (d *Driver) CheckLogin(con rt.Conversation, cmd, params string, user publisher.User) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) Login(con rt.Conversation, user publisher.User) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) RequestExternalAccess(con rt.Conversation) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) Retrieve(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) List(con rt.Conversation) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) Delete(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	return
}

func executeTemplate(tpl *template.Template, data interface{}) (string, error) {
	var buf strings.Builder
	err := tpl.Execute(&buf, data)
	return buf.String(), err
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
