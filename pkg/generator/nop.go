package generator

var _ UserConfig = (*nopUserConfig)(nil)

type nopUserConfig struct{}

func (c *nopUserConfig) SetAuthToken(token string) {}

type nopConfig struct{}

var _ Interface = (*nop)(nil)

type nop struct{}

func (a *nop) Name() string { return "nop" }

// Login to platform
func (a *nop) Login(config UserConfig) (token string, _ error) { return "", nil }

// AuthURL return a one click url for external authorization
func (a *nop) AuthURL() (string, error) { return "", nil }

// Retrieve post and cache it locally according to the url
func (a *nop) Retrieve(url string) (title string, _ error) { return "", nil }

// Publish a new post
func (a *nop) Publish(title string, body []byte) (url string, _ error) { return "", nil }

// Append content to local post cache
func (a *nop) Append(title string, body []byte) (url string, _ error) { return "", nil }

func (a *nop) FormatPagePrefix() ([]byte, error) {
	return nil, nil
}

func (a *nop) FormatPageContent(messages []Message, funcMap FuncMap) ([]byte, error) {
	return nil, nil
}
