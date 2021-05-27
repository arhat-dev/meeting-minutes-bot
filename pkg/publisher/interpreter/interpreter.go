package interpreter

import (
	"context"
	"fmt"
	"os/exec"

	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

// nolint:revive
const (
	Name = "interpreter"
)

func init() {
	publisher.Register(
		Name,
		func(config interface{}) (publisher.Interface, publisher.UserConfig, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, nil, fmt.Errorf("unexpected non interpreter config")
			}

			return &Driver{
				bin:      c.Bin,
				baseArgs: c.BaseArgs,
			}, &UserConfig{}, nil
		},
		func() interface{} {
			return &Config{}
		},
	)
}

type Config struct {
	Bin      string   `json:"bin" yaml:"bin"`
	BaseArgs []string `json:"baseArgs" yaml:"baseArgs"`
}

var _ publisher.UserConfig = (*UserConfig)(nil)

type UserConfig struct{}

func (u *UserConfig) SetAuthToken(token string) {}

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	bin      string
	baseArgs []string
}

func (d *Driver) Name() string                                              { return Name }
func (d *Driver) RequireLogin() bool                                        { return false }
func (d *Driver) Login(config publisher.UserConfig) (token string, _ error) { return "", nil }
func (d *Driver) AuthURL() (string, error)                                  { return "", nil }
func (d *Driver) Retrieve(url string) (title string, _ error)               { return "", nil }
func (d *Driver) List() ([]publisher.PostInfo, error)                       { return nil, nil }
func (d *Driver) Delete(urls ...string) error                               { return nil }

func (d *Driver) Append(title string, body []byte) (url string, _ error) {
	var args []string
	args = append(args, d.baseArgs...)
	args = append(args, string(body))

	cmd := exec.CommandContext(
		context.TODO(),
		d.bin,
		args...,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		if len(output) != 0 {
			return fmt.Sprintf("%s\n%v", output, err), nil
		} else {
			return err.Error(), nil
		}
	}

	return string(output), nil
}

func (d *Driver) Publish(title string, body []byte) (url string, _ error) { return "", nil }
