package chain

import (
	"fmt"

	"arhat.dev/mbot/pkg/generator"
)

const (
	Name = "chain"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
}

type Config []generator.Config

func (c *Config) Create() (_ generator.Interface, err error) {
	sz := len(*c)
	if sz != 1 {
		err = fmt.Errorf("unexpected count of config items %d (want exact one config)", sz)
		return
	}

	ret := &Driver{
		underlay: make([]generator.Interface, sz),
	}

	for i, cfg := range *c {
		ret.underlay[i], err = cfg.Create()
		if err != nil {
			return
		}
	}

	return ret, nil
}
