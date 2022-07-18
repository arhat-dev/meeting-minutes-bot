package archiver

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/rs"
)

const (
	Name = "archiver"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
}

type Config struct {
	rs.BaseField

	// ExcludeDomains excludes urls matched domain in this list
	ExcludeDomains []string `yaml:"excludeDomains"`
}

// Create implements generator.Config
func (*Config) Create() (generator.Interface, error) {
	return &Driver{}, nil
}
