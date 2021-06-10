package webarchiver

import "context"

var _ Interface = (*Nop)(nil)

type NopConfig struct{}

type Nop struct{}

func (a *Nop) Login(config interface{}) error {
	return nil
}

func (a *Nop) Archive(
	ctx context.Context,
	url string,
) (
	archiveURL string,
	screenshot []byte,
	err error,
) {
	return "", nil, nil
}
