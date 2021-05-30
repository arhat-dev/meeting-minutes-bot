package interpreter

import (
	"context"
	"fmt"
	"os/exec"

	"arhat.dev/meeting-minutes-bot/pkg/message"
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

func (d *Driver) Name() string {
	return Name
}

func (d *Driver) RequireLogin() bool {
	return false
}

func (d *Driver) Login(config publisher.UserConfig) (token string, _ error) {
	return "", fmt.Errorf("unimplemented")
}

func (d *Driver) AuthURL() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (d *Driver) Retrieve(key string) ([]message.Entity, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) List() ([]publisher.PostInfo, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) Delete(urls ...string) error {
	return fmt.Errorf("unimplemented")
}

func (d *Driver) Append(body []byte) ([]message.Entity, error) {
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
			return []message.Entity{
				{
					Kind: message.KindPre,
					Text: fmt.Sprintf("%s\n%v", output, err),
				},
			}, nil
		}

		return []message.Entity{
			{
				Kind: message.KindPre,
				Text: err.Error(),
			},
		}, nil
	}

	return []message.Entity{
		{
			Kind: message.KindPre,
			Text: string(output),
		},
	}, nil
}

func (d *Driver) Publish(title string, body []byte) ([]message.Entity, error) {
	return []message.Entity{
		{
			Kind: message.KindText,
			Text: fmt.Sprintf("You are using %s interpreter.", d.bin),
		},
	}, nil
}
