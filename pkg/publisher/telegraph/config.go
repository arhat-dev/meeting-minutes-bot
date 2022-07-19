package telegraph

import (
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
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

type Config struct {
	rs.BaseField

	DefaultAccountShortName string `yaml:"defaultAccountShortName"`
}

func (c *Config) Create() (_ publisher.Interface, _ publisher.User, err error) {
	client, err := newDefaultClient()
	if err != nil {
		return
	}

	return &Driver{
		client: client,

		defaultAccountShortName: c.DefaultAccountShortName,
	}, &User{}, nil
}

var _ publisher.User = (*User)(nil)

type User struct {
	// TODO: support overriding
	shortName  string
	authorName string
	authorURL  string

	authToken string
}

// NextCredential implements publisher.User
func (u *User) NextCredential() (flow rt.LoginFlow) {
	switch {
	case len(u.authToken) == 0:
		return rt.LoginFlow_Token
	}

	return
}

// SetPassword implements publisher.User
func (*User) SetPassword(string) {}

// SetTOTPCode implements publisher.User
func (*User) SetTOTPCode(string) {}

// SetToken implements publisher.User
func (u *User) SetToken(token string) { u.authToken = token }

// SetUsername implements publisher.User
func (*User) SetUsername(string) {}
