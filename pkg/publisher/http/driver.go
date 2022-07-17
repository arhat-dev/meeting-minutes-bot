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
func (*Driver) CreateNew(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) ([]rt.Span, error) {
	panic("unimplemented")
}

// AppendToExisting implements publisher.Interface
//
// fromGenerator is expected to provide yaml spec of the http request
func (d *Driver) AppendToExisting(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) (_ []rt.Span, err error) {
	content, err := fromGenerator.Bytes()
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

func (d *Driver) RequireLogin(con rt.Conversation, cmd, params string) (rt.LoginFlow, error) {
	return rt.LoginFlow_None, nil
}

func (d *Driver) Login(con rt.Conversation, user publisher.User) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) RequestExternalAccess(con rt.Conversation) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) Retrieve(con rt.Conversation, cmd, params string) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) List(con rt.Conversation) ([]publisher.PostInfo, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) Delete(con rt.Conversation, cmd, params string) error {
	return fmt.Errorf("unimplemented")
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
