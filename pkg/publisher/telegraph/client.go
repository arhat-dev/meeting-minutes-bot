package telegraph

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	urlpkg "net/url"
	pathpkg "path"
	"strings"

	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/pkg/stringhelper"
)

func newDefaultClient() (_ client, err error) {
	u, err := url.Parse(apiBaseURL)
	if err != nil {
		return
	}

	return client{
		baseURL: *u,
		client:  http.Client{},
	}, nil
}

type client struct {
	baseURL urlpkg.URL
	client  http.Client
}

func (c *client) setURL(url *urlpkg.URL, path string, query urlpkg.Values) {
	*url = c.baseURL
	url.Path = pathpkg.Join(c.baseURL.Path, path)
	url.RawQuery = query.Encode()
}

type createAccountOptions struct {
	// Required. Account name, helps users with several accounts remember which they are currently using. Displayed to the user above the "Edit/Publish" button on Telegra.ph, other users don't see this name.
	ShortName string `json:"short_name,omitempty"` // 1-32 characters
	// Default author name used when creating new articles.
	AuthorName string `json:"author_name,omitempty"` // 0-128 characters
	// Default profile link, opened when users click on the author's name below the title. Can be any link, not necessarily to a Telegram profile or channel.
	AuthorURL string `json:"author_url,omitempty"` // 0-512 characters
}

func (c *client) CreateAccount(opts createAccountOptions) (ret telegraphAccount, err error) {
	var (
		url urlpkg.URL
	)

	c.setURL(&url, "createAccount", nil)
	return post[telegraphAccount](&c.client, &url, &opts)
}

func (c *client) GetAccountInfo(accessToken string) (ret telegraphAccount, err error) {
	type Request struct {
		AccessToken string          `json:"access_token"`
		Fields      json.RawMessage `json:"fields"`
	}

	var (
		url    urlpkg.URL
		fields = `["short_name","author_name","author_url","auth_url","page_count"]`
	)

	req := Request{
		AccessToken: accessToken,
		Fields:      stringhelper.ToBytes[byte, byte](fields),
	}

	c.setURL(&url, "getAccountInfo", nil)
	return post[telegraphAccount](&c.client, &url, &req)
}

type getPageListOptions struct {
	AccessToken string `json:"access_token"`
	Offset      int64  `json:"offset"`
	Limit       int64  `json:"limit"`
}

func (c *client) GetPageList(opts getPageListOptions) (ret telegraphPageList, err error) {
	var url urlpkg.URL
	c.setURL(&url, "getPageList", nil)
	return post[telegraphPageList](&c.client, &url, &opts)
}

func (c *client) GetPage(path string) (ret telegraphPage, err error) {
	type Request struct {
		Path          string `json:"path"`
		ReturnContent bool   `json:"return_content"`
	}

	var url urlpkg.URL
	c.setURL(&url, "getPage", nil)

	opts := Request{
		Path:          path,
		ReturnContent: true,
	}

	return post[telegraphPage](&c.client, &url, &opts)
}

type createPageOptions struct {
	// Required. Access token of the Telegraph account.
	AccessToken string `json:"access_token,omitempty"`
	// Required. Page Title.
	Title string `json:"title,omitempty"` // 1-256 characters
	// Author name, displayed below the article's title.
	AuthorName string `json:"author_name,omitempty"` // 0-128 characters
	// Profile link, opened when users click on the author's name below the title. Can be any link, not necessarily to a Telegram profile or channel.
	AuthorURL string `json:"author_url,omitempty"` // 0-512 characters
	// Required. Content of the page.
	Content []telegraphNode `json:"content,omitempty"` // up to 64 KB
	// If true, a content field will be returned in the Page object (see: Content format).
	ReturnContent bool `json:"return_content,omitempty"`
}

func (c *client) CreatePage(opts createPageOptions) (ret telegraphPage, err error) {
	var url urlpkg.URL
	c.setURL(&url, "createPage", nil)
	return post[telegraphPage](&c.client, &url, &opts)
}

func (c *client) EditPage(path string, opts createPageOptions) (ret telegraphPage, err error) {
	type Request struct {
		// Required. Path to the page.
		Path string `json:"path,omitempty"`

		// same as createPageOptions

		// Required. Access token of the Telegraph account.
		AccessToken string `json:"access_token,omitempty"`
		// Required. Page Title.
		Title string `json:"title,omitempty"` // 1-256 characters
		// Author name, displayed below the article's title.
		AuthorName string `json:"author_name,omitempty"` // 0-128 characters
		// Profile link, opened when users click on the author's name below the title. Can be any link, not necessarily to a Telegram profile or channel.
		AuthorURL string `json:"author_url,omitempty"` // 0-512 characters
		// Required. Content of the page.
		Content []telegraphNode `json:"content,omitempty"` // up to 64 KB
		// If true, a content field will be returned in the Page object (see: Content format).
		ReturnContent bool `json:"return_content,omitempty"`
	}

	var url urlpkg.URL
	c.setURL(&url, "editPage", nil)

	req := Request{
		Path:          path,
		AccessToken:   opts.AccessToken,
		Title:         opts.Title,
		AuthorName:    opts.AuthorName,
		AuthorURL:     opts.AuthorURL,
		Content:       opts.Content,
		ReturnContent: opts.ReturnContent,
	}
	return post[telegraphPage](&c.client, &url, &req)
}

type response[T any] struct {
	Ok     bool   `json:"ok"`
	Result T      `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

type requestOptions struct {
	method  string
	headers http.Header
	url     *urlpkg.URL
	body    io.Reader
}

var headerValues = [headerValueCount]string{
	headerValueIndex_MIMETypeJSON: "application/json",
	headerValueIndex_Referer:      "https://telegra.ph",
	headerValueIndex_UserAgent:    "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.5060.114 Safari/537.36",
}

const (
	headerValueIndex_MIMETypeJSON = iota
	headerValueIndex_Referer
	headerValueIndex_UserAgent

	headerValueCount
)

// func get[R any](client *http.Client, url *urlpkg.URL) (out R, err error) {
// 	opts := requestOptions{
// 		method: http.MethodGet,
// 		url:    url,
// 		headers: http.Header{
// 			"Accept": headerValues[headerValueIndex_ApplicationJSON : headerValueIndex_ApplicationJSON+1],
// 		},
// 		body: nil,
// 	}
//
// 	return request[R](client, &opts)
// }

func post[R any](client *http.Client, url *urlpkg.URL, data any) (out R, err error) {
	var body bytes.Buffer
	enc := json.NewEncoder(&body)
	err = enc.Encode(data)
	if err != nil {
		return
	}

	opts := requestOptions{
		method: http.MethodPost,
		url:    url,
		headers: http.Header{
			"Origin":       headerValues[headerValueIndex_Referer : headerValueIndex_Referer+1],
			"Referer":      headerValues[headerValueIndex_Referer : headerValueIndex_Referer+1],
			"Content-Type": headerValues[headerValueIndex_MIMETypeJSON : headerValueIndex_MIMETypeJSON+1],
			"Accept":       headerValues[headerValueIndex_MIMETypeJSON : headerValueIndex_MIMETypeJSON+1],
			"User-Agent":   headerValues[headerValueIndex_UserAgent : headerValueIndex_UserAgent+1],
		},
		body: &body,
	}

	return request[R](client, &opts)
}

type errString string

func (e errString) Error() string { return string(e) }

func request[R any](client *http.Client, opts *requestOptions) (out R, err error) {
	var req http.Request
	req.Method = opts.method
	req.URL = opts.url
	req.Host = removeEmptyPort(opts.url.Host)
	req.Header = opts.headers

	if rc, ok := opts.body.(io.ReadCloser); ok {
		req.Body = rc
	} else {
		req.Body = io.NopCloser(opts.body)
	}

	resp, err := client.Do(&req)
	if err != nil {
		return
	}

	dec := json.NewDecoder(resp.Body)
	var body response[R]
	err = dec.Decode(&body)
	_ = resp.Body.Close()
	if err != nil {
		return
	}

	if body.Ok {
		out = body.Result
	} else {
		err = errString(body.Error)
	}

	return
}

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }

// removeEmptyPort strips the empty port in ":port" to ""
// as mandated by RFC 3986 Section 6.2.3.
func removeEmptyPort(host string) string {
	if hasPort(host) {
		return strings.TrimSuffix(host, ":")
	}
	return host
}

type telegraphPageList struct {
	Pages      []telegraphPage `json:"pages"`
	TotalCount int64           `json:"total_count"`
}

type telegraphPage struct {
	// Path to the page.
	Path string `json:"path,omitempty"`
	// URL of the page.
	URL string `json:"url,omitempty"`
	// Title of the page.
	Title string `json:"title,omitempty"`
	// Description of the page.
	Description string `json:"description,omitempty"`
	// Optional. Name of the author, displayed below the title.
	AuthorName string `json:"author_name,omitempty"`
	// Optional. Profile link, opened when users click on the author's name below the title.  Can be any link, not necessarily to a Telegram profile or channel.
	AuthorURL string `json:"author_url,omitempty"`
	// Optional. Image URL of the page.
	ImageURL string `json:"image_url,omitempty"`
	// Optional. Content of the page.
	Content []telegraphNode `json:"content,omitempty"`
	// Number of page views for the page.
	Views int64 `json:"views,omitempty"`
	// Optional. Only returned if access_token passed. True, if the target Telegraph account can edit the page.
	CanEdit bool `json:"can_edit,omitempty"`
}

type telegraphNode struct {
	Elm  rt.Optional[telegraphNodeElement]
	Text string
}

var (
	_ json.Unmarshaler = (*telegraphNode)(nil)
	_ json.Marshaler   = (*telegraphNode)(nil)
)

func (m *telegraphNode) UnmarshalJSON(data []byte) (err error) {
	if len(data) == 0 {
		return nil
	}

	if data[0] == '{' {
		var elm telegraphNodeElement
		err = json.Unmarshal(data, &elm)
		m.Elm.Set(elm)
		return
	}

	return json.Unmarshal(data, &m.Text)
}

func (m *telegraphNode) MarshalJSON() ([]byte, error) {
	if m.Elm.IsNil() {
		return json.Marshal(m.Text)
	}

	return json.Marshal(m.Elm.GetPtr())
}

type telegraphNodeElement struct {
	// Name of the DOM element. Available tags: a, aside, b, blockquote, br, code, em, figcaption, figure, h3, h4, hr, i, iframe, img, li, ol, p, pre, s, strong, u, ul, video.
	Tag string `json:"tag,omitempty"`
	// Optional. Attributes of the DOM element. Key of object represents name of attribute, value represents value of attribute. Available attributes: href, src.
	Attrs map[string]string `json:"attrs,omitempty"`
	// Optional. List of child nodes for the DOM element.
	Children []telegraphNode `json:"children,omitempty"`
}

type telegraphAccount struct {
	// Account name, helps users with several accounts remember which they are currently using. Displayed to the user above the "Edit/Publish" button on Telegra.ph, other users don't see this name.
	ShortName string `json:"short_name,omitempty"`
	// Default author name used when creating new articles.
	AuthorName string `json:"author_name,omitempty"`
	// Profile link, opened when users click on the author's name below the title. Can be any link, not necessarily to a Telegram profile or channel.
	AuthorURL string `json:"author_url,omitempty"`
	// Optional. Only returned by the createAccount and revokeAccessToken method. Access token of the Telegraph account.
	AccessToken string `json:"access_token,omitempty"`
	// Optional. URL to authorize a browser on telegra.ph and connect it to a Telegraph account. This URL is valid for only one use and for 5 minutes only.
	AuthURL string `json:"auth_url,omitempty"`
	// Optional. Number of pages belonging to the Telegraph account.
	PageCount int64 `json:"page_count,omitempty"`
}
