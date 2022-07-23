package multigen

import (
	"arhat.dev/mbot/pkg/generator"
)

const (
	Name = "multigen"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
}

type Config []generator.Config

// Create implements generator.Config
func (c *Config) Create() (_ generator.Interface, err error) {
	underlay := make([]generator.Interface, len(*c))
	for i, cfg := range *c {
		underlay[i], err = cfg.Create()
		if err != nil {
			return
		}
	}

	return &Driver{underlay: underlay}, nil
}
