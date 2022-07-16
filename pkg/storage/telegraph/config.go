package telegraph

import (
	"arhat.dev/mbot/pkg/storage"
	"arhat.dev/rs"
)

type Config struct {
	rs.BaseField

	storage.CommonConfig `yaml:",inline"`
}

func (c *Config) MIMEMatch() string { return c.CommonConfig.MIMEMatch }
func (c *Config) MaxSize() int64    { return c.CommonConfig.MaxSize }

func (c *Config) Create() (storage.Interface, error) { return &Driver{}, nil }
