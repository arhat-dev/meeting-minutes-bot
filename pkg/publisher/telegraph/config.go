package telegraph

import (
	"sync"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/rs"
)

const (
	apiBaseURL = "https://api.telegra.ph/"
)

func init() {
	publisher.Register(
		Name,
		func() publisher.Config {
			return &Config{
				DefaultAccountShortName: "mbot",
			}
		},
	)
}

var _ publisher.UserConfig = (*userConfig)(nil)

type userConfig struct {
	// TODO: support overriding
	shortName  string
	authorName string
	authorURL  string

	authToken string
}

func (c *userConfig) SetAuthToken(token string) {
	c.authToken = token
}

type Config struct {
	rs.BaseField

	DefaultAccountShortName string `yaml:"defaultAccountShortName"`
}

func (c *Config) Create() (_ publisher.Interface, _ publisher.UserConfig, err error) {
	client, err := newDefaultClient()
	if err != nil {
		return
	}

	return &Driver{
		client: client,

		defaultAccountShortName: c.DefaultAccountShortName,

		mu: &sync.RWMutex{},
	}, &userConfig{}, nil
}
