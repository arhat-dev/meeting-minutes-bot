package exec

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/rs"
)

const (
	Name = "exec"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
}

type Config struct {
	rs.BaseField

	// WorkDir when exec start
	WorkDir string `yaml:"workdir"`

	// Executable is the path to the executable file
	Executable string `yaml:"executable"`

	// Args
	Args []string `yaml:"args"`
}

// Create implements generator.Config
func (*Config) Create() (generator.Interface, error) {
	return &Driver{}, nil
}
