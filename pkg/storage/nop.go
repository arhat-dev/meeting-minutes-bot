package storage

import "context"

var _ Interface = (*Nop)(nil)

type nopConfig struct{}

func (nopConfig) Create() (Interface, error) { return Nop{}, nil }

type Nop struct{}

func (Nop) Name() string { return "nop" }

func (Nop) Upload(
	ctx context.Context,
	filename string,
	contentType string,
	data []byte,
) (url string, err error) {
	return filename, nil
}
