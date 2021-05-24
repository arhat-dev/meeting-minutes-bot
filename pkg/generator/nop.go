package generator

var _ UserConfig = (*NopUserConfig)(nil)

type NopUserConfig struct{}

func (c *NopUserConfig) SetAuthToken(token string) {}

type NopConfig struct{}

var _ Interface = (*Nop)(nil)

type Nop struct{}

func (a *Nop) Name() string { return "Nop" }

// Login to platform
func (a *Nop) Login(config UserConfig) (token string, _ error) { return "", nil }

// AuthURL return a one click url for external authorization
func (a *Nop) AuthURL() (string, error) { return "", nil }

// Retrieve post and cache it locally according to the url
func (a *Nop) Retrieve(url string) (title string, _ error) { return "", nil }

// Publish a new post
func (a *Nop) Publish(title string, body []byte) (url string, _ error) { return "", nil }

// Append content to local post cache
func (a *Nop) Append(title string, body []byte) (url string, _ error) { return "", nil }

func (a *Nop) Format(kind FormatKind, text string, params ...string) string { return "" }
