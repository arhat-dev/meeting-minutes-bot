package http

import (
	"net/http"
	"net/url"

	"arhat.dev/rs"
)

type Spec struct {
	rs.BaseField

	Params []nameValuePair `yaml:"params"`

	// base64 encoded request body
	Body string `yaml:"body"`
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
