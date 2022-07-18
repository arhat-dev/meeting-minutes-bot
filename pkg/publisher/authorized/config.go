package authorized

import (
	"fmt"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/rs"
)

const (
	Name = "authorized"
)

func init() {
	publisher.Register(Name, func() publisher.Config { return &Config{} })
}

type Config struct {
	rs.BaseField

	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	TOTPSecret string `yaml:"totpSecret"`
	Token      string `yaml:"token"`

	// For sets a config
	For map[string]publisher.Config `yaml:"for"`
}

// Create implements publisher.Config
func (c *Config) Create() (_ publisher.Interface, user publisher.User, err error) {
	if len(c.For) != 1 {
		err = fmt.Errorf("unexpected count of config items %d (want exact one config)", len(c.For))
	}

	var (
		impl publisher.Interface
	)

	for _, cfg := range c.For {
		impl, user, err = cfg.Create()
		if err != nil {
			return
		}
	}

	user.SetUsername(c.Username)
	user.SetPassword(c.Password)
	// TODO: generate totp code from totp secret
	// user.SetTOTPCode("")
	user.SetToken(c.Token)

	return &Driver{Interface: impl}, user, nil
}
