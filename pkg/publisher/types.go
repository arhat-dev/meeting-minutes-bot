package publisher

type Interface interface {
	Name() string

	RequireLogin() bool

	// Login to platform
	Login(config UserConfig) (token string, _ error)

	// AuthURL return a one click url for external authorization
	AuthURL() (string, error)

	// Retrieve post and cache it locally according to the url
	Retrieve(url string) (title string, _ error)

	// Publish a new post
	Publish(title string, body []byte) (url string, _ error)

	// List all posts for this user
	List() ([]PostInfo, error)

	// Delete one post according to the url
	Delete(urls ...string) error

	// Append content to local post cache
	Append(title string, body []byte) (url string, _ error)
}

type PostInfo struct {
	Title string
	URL   string
}

type UserConfig interface {
	SetAuthToken(token string)
}
