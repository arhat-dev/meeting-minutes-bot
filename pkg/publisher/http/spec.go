package http

type Spec struct {
	QueryParams []nameValuePair `json:"queryParams" yaml:"queryParams"`

	// base64 encoded request body
	Body string `json:"body" yaml:"body"`
}
