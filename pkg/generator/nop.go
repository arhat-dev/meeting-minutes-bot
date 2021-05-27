package generator

import "arhat.dev/meeting-minutes-bot/pkg/message"

var _ UserConfig = (*nopUserConfig)(nil)

type nopUserConfig struct{}

func (c *nopUserConfig) SetAuthToken(token string) {}

type nopConfig struct{}

var _ Interface = (*nop)(nil)

type nop struct{}

func (a *nop) Name() string                                            { return "nop" }
func (a *nop) Login(config UserConfig) (token string, _ error)         { return "", nil }
func (a *nop) AuthURL() (string, error)                                { return "", nil }
func (a *nop) Retrieve(url string) (title string, _ error)             { return "", nil }
func (a *nop) Publish(title string, body []byte) (url string, _ error) { return "", nil }
func (a *nop) Append(title string, body []byte) (url string, _ error)  { return "", nil }
func (a *nop) List() ([]PostInfo, error)                               { return nil, nil }
func (a *nop) Delete(urls ...string) error                             { return nil }
func (a *nop) FormatPageHeader() ([]byte, error)                       { return nil, nil }

func (a *nop) FormatPageBody(messages []message.Interface) ([]byte, error) {
	return nil, nil
}
