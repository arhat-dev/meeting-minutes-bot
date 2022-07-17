package file

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/pkg/fshelper"
	"arhat.dev/rs"
)

// nolint:revive
const (
	Name = "file"
)

func init() {
	generator.Register(Name, func() generator.Config { return &Config{} })
}

type Config struct {
	rs.BaseField

	Dir string `json:"dir" yaml:"dir"`
}

func (c *Config) Create() (_ generator.Interface, err error) {
	fs := fshelper.NewOSFS(false, func() (string, error) {
		return c.Dir, nil
	})

	err = fs.MkdirAll(".", 0755)
	if err != nil {
		return
	}

	return &Driver{fs: fs}, nil
}
