package telegraph

import (
	"arhat.dev/mbot/pkg/storage"
	"arhat.dev/rs"
)

type Config struct {
	rs.BaseField

	// TODO
}

func (c *Config) Create() (storage.Interface, error) { return &Driver{}, nil }
