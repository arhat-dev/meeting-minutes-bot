package cron

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/rs"
)

const (
	Name = "cron"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
}

type Config struct {
	rs.BaseField

	// TODO
}

func (c *Config) Create() (generator.Interface, error) {
	return &Driver{}, nil
}
