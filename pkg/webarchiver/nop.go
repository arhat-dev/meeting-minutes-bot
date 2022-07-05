package webarchiver

import "context"

var _ Interface = (*nop)(nil)

type NopConfig struct{}

func (NopConfig) Create() (Interface, error) { return nop{}, nil }

type nop struct{}

func (nop) Archive(
	ctx context.Context,
	url string,
) (
	archiveURL string,
	screenshot []byte,
	err error,
) {
	return "", nil, nil
}
