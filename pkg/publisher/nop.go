package publisher

import (
	"context"
	"fmt"

	"arhat.dev/mbot/pkg/rt"
)

var _ UserConfig = (*nopUserConfig)(nil)

type nopUserConfig struct{}

func (nopUserConfig) SetAuthToken(token string) {}

type nopConfig struct{}

func (nopConfig) Create() (Interface, UserConfig, error) {
	return nop{}, nopUserConfig{}, nil
}

var _ Interface = (*nop)(nil)

type nop struct{}

func (nop) Name() string {
	return "nop"
}

func (nop) RequireLogin() bool {
	return false
}

func (nop) Login(config UserConfig) (token string, _ error) {
	return "", fmt.Errorf("unimplemented")
}

func (nop) AuthURL() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (nop) Retrieve(url string) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (nop) Publish(title string, body *rt.Input) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (nop) Append(ctx context.Context, body *rt.Input) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (nop) List() ([]PostInfo, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (nop) Delete(urls ...string) error {
	return fmt.Errorf("unimplemented")
}
