package http

import (
	"net/http"
	"net/url"
)

type Spec struct {
	Params []nameValuePair `json:"params" yaml:"params"`

	// base64 encoded request body
	Body string `json:"body" yaml:"body"`
}

type responseTemplateRequestData struct {
	URL     *url.URL
	Headers http.Header
}

type responseTemplateData struct {
	Code    int
	Headers http.Header
	Body    []byte
	Request responseTemplateRequestData
}
