package tengo

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/rs"
)

const (
	Name = "tengo"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
}

type Config struct {
	rs.BaseField
}

// Create implements generator.Config
func (*Config) Create() (generator.Interface, error) {
	return &Driver{}, nil
}
