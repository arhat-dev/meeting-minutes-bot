package bot

import (
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/rs"
)

const (
	Name = "bot"
)

func init() {
	publisher.Register(Name, func() publisher.Config { return &Config{} })
}

type Config struct {
	rs.BaseField
}

func (c *Config) Create() (_ publisher.Interface, _ publisher.User, err error) {
	return
}
