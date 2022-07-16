package compositor

import (
	"fmt"
	"math"
	"regexp"

	"arhat.dev/mbot/pkg/storage"
	"arhat.dev/rs"
)

const (
	Name = "compositor"
)

func init() {
	storage.Register(Name, func() storage.Config { return &Config{} })
}

var _ storage.Config = (*Config)(nil)

type Config struct {
	rs.BaseField

	Specs []singleSpec `yaml:"specs"`
}

func (c *Config) Create() (_ storage.Interface, err error) {
	sz := len(c.Specs)

	mgr := &Driver{
		underlay: make([]impl, sz),
	}

	for i := 0; i < sz; i++ {
		mgr.underlay[i], err = c.Specs[i].resolve()
		if err != nil {
			return
		}
	}

	return mgr, nil
}

type singleSpec struct {
	rs.BaseField

	MIMEMatch string `yaml:"mimeMatch"`
	MaxSize   *int64 `yaml:"maxSize"`

	Config map[string]storage.Config `yaml:",inline"`
}

func (spec *singleSpec) resolve() (ret impl, err error) {
	if len(spec.Config) != 1 {
		err = fmt.Errorf("unexpected count of config items %d (want exact one config)", len(spec.Config))
		return
	}

	if len(spec.MIMEMatch) != 0 {
		ret.exp, err = regexp.CompilePOSIX(spec.MIMEMatch)
		if err != nil {
			err = fmt.Errorf("invalid mime match regular expression: %w", err)
			return
		}
	}

	for _, cfg := range spec.Config {
		ret.store, err = cfg.Create()
		if err != nil {
			return
		}
	}

	ret.maxSize = int64(math.MaxInt64)
	if spec.MaxSize != nil {
		ret.maxSize = *spec.MaxSize
	}

	return
}
