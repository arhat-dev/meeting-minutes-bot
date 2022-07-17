package file

import (
	"fmt"
	"os"
	"path/filepath"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/rs"
)

// nolint:revive
const (
	Name = "file"
)

func init() {
	publisher.Register(
		Name,
		func() publisher.Config { return &Config{} },
	)
}

type Config struct {
	rs.BaseField

	Dir string `json:"dir" yaml:"dir"`
}

func (c *Config) Create() (publisher.Interface, publisher.User, error) {
	dir, err := filepath.Abs(c.Dir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to determin absolute dir path: %w", err)
	}

	err = os.MkdirAll(dir, 0755)
	if err != nil && !os.IsExist(err) {
		return nil, nil, fmt.Errorf("failed to ensure dir: %w", err)
	}

	return &Driver{dir: dir}, &UserConfig{}, nil
}

var _ publisher.User = (*UserConfig)(nil)

type UserConfig struct{}

func (u *UserConfig) SetAuthToken(token string) {}
