package multigen

import (
	"fmt"

	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/rs"
)

const (
	Name = "multigen"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
}

type Config struct {
	rs.BaseField

	Specs []singleSpec `yaml:"specs"`
}

// Create implements generator.Config
func (c *Config) Create() (_ generator.Interface, err error) {
	underlay := make([]generator.Interface, len(c.Specs))
	for i := range underlay {
		underlay[i], err = c.Specs[i].resolve()
		if err != nil {
			return
		}
	}

	return &Driver{underlay: underlay}, nil
}

type singleSpec struct {
	rs.BaseField

	Config map[string]generator.Config `yaml:",inline"`
}

func (spec *singleSpec) resolve() (_ generator.Interface, err error) {
	if len(spec.Config) != 1 {
		err = fmt.Errorf("unexpected count of config items %d (want exact one config)", len(spec.Config))
		return
	}

	for _, cfg := range spec.Config {
		return cfg.Create()
	}

	panic("unreachable")
}
