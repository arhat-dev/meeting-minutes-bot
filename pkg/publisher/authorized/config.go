package authorized

import (
	"fmt"
	"time"

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

	// Token for publishers expecting token
	Token string `yaml:"token"`

	// Username for publishers expecting username input
	Username string `yaml:"username"`
	// Password for publishers expecting password input
	Password string `yaml:"password"`

	// TOTPToken and TOTPCodeDigits for publishers expecting totp code input
	TOTPToken string `yaml:"totpToken"`
	// TOTPCodeDigits and TOTPToken for publishers expecting totp code input
	//
	// Defaults to 6
	TOTPCodeDigits int `yaml:"totpCodeDigits"`

	// For is the underlay publisher (expect exact one entry)
	For map[string]publisher.Config `yaml:"for"`
}

// Create implements publisher.Config
func (c *Config) Create() (_ publisher.Interface, user publisher.User, err error) {
	if len(c.For) != 1 {
		err = fmt.Errorf("unexpected count of config items %d (want exact one config)", len(c.For))
		return
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

	if len(c.TOTPToken) != 0 {
		var (
			totpCode string
			digits   = c.TOTPCodeDigits
		)
		if digits == 0 {
			digits = 6
		}

		totpCode, err = generateTOTPCode(c.TOTPToken, time.Now(), digits)
		if err != nil {
			return
		}

		user.SetTOTPCode(totpCode)
	}

	user.SetToken(c.Token)
	user.SetUsername(c.Username)
	user.SetPassword(c.Password)

	return &Driver{Interface: impl}, user, nil
}
