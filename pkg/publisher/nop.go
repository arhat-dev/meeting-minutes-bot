package publisher

import (
	"context"
	"fmt"

	"arhat.dev/meeting-minutes-bot/pkg/message"
)

var _ UserConfig = (*nopUserConfig)(nil)

type nopUserConfig struct{}

func (c *nopUserConfig) SetAuthToken(token string) {}

type nopConfig struct{}

var _ Interface = (*nop)(nil)

type nop struct{}

func (a *nop) Name() string {
	return "nop"
}

func (a *nop) RequireLogin() bool {
	return false
}

func (a *nop) Login(config UserConfig) (token string, _ error) {
	return "", fmt.Errorf("unimplemented")
}

func (a *nop) AuthURL() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (a *nop) Retrieve(url string) ([]message.Entity, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (a *nop) Publish(title string, body []byte) ([]message.Entity, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (a *nop) Append(ctx context.Context, body []byte) ([]message.Entity, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (a *nop) List() ([]PostInfo, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (a *nop) Delete(urls ...string) error {
	return fmt.Errorf("unimplemented")
}
