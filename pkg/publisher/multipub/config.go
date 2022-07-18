package multipub

import (
	"fmt"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/rs"
)

const (
	Name = "multigen"
)

func init() {
	publisher.Register(Name, func() publisher.Config { return &Config{} })
}

type Config struct {
	rs.BaseField

	Specs []singleSpec `yaml:"specs"`
}

type pair struct {
	impl publisher.Interface
	user publisher.User
}

// Create implements publisher.Config
func (c *Config) Create() (_ publisher.Interface, _ publisher.User, err error) {
	underlay := make([]pair, len(c.Specs))
	for i := range underlay {
		underlay[i].impl, underlay[i].user, err = c.Specs[i].resolve()
		if err != nil {
			return
		}
	}

	return &Driver{underlay: underlay}, nil, nil
}

type singleSpec struct {
	rs.BaseField

	Config map[string]publisher.Config `yaml:",inline"`
}

func (spec *singleSpec) resolve() (_ publisher.Interface, _ publisher.User, err error) {
	if len(spec.Config) != 1 {
		err = fmt.Errorf("unexpected count of config items %d (want exact one config)", len(spec.Config))
		return
	}

	for _, cfg := range spec.Config {
		return cfg.Create()
	}

	panic("unreachable")
}
