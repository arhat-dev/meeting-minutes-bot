package file

import (
	"fmt"
	"os"
	"path/filepath"

	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

// nolint:revive
const (
	Name = "file"
)

func init() {
	publisher.Register(
		Name,
		func(config interface{}) (publisher.Interface, publisher.UserConfig, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, nil, fmt.Errorf("unexpected non file config")
			}

			return &Driver{
				dir: c.Dir,
			}, &UserConfig{}, nil
		},
		func() interface{} {
			return &Config{}
		},
	)
}

type Config struct {
	Dir string `json:"dir" yaml:"dir"`
}

var _ publisher.UserConfig = (*UserConfig)(nil)

type UserConfig struct{}

func (u *UserConfig) SetAuthToken(token string) {}

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	dir string
}

func (d *Driver) Name() string                                              { return Name }
func (d *Driver) RequireLogin() bool                                        { return false }
func (d *Driver) Login(config publisher.UserConfig) (token string, _ error) { return "", nil }
func (d *Driver) AuthURL() (string, error)                                  { return "", nil }
func (d *Driver) Retrieve(url string) (title string, _ error)               { return "", nil }
func (d *Driver) List() ([]publisher.PostInfo, error)                       { return nil, nil }
func (d *Driver) Delete(urls ...string) error                               { return nil }

func (d *Driver) Publish(title string, body []byte) (url string, _ error) {
	path := filepath.Join(d.dir, title)
	return path, os.WriteFile(path, body, 0640)
}

func (d *Driver) Append(title string, body []byte) (url string, _ error) {
	path := filepath.Join(d.dir, title)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0640)
	if err != nil {
		return "", err
	}

	defer func() { _ = f.Close() }()

	_, err = f.Write(body)
	if err != nil {
		return "", err
	}

	return path, nil
}
