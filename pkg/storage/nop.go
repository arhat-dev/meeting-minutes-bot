package storage

import "context"

var _ Interface = (*Nop)(nil)

type NopConfig struct{}

type Nop struct{}

func (u *Nop) Name() string {
	return "nop"
}

func (u *Nop) Upload(ctx context.Context, filename string, data []byte) (url string, err error) {
	return filename, nil
}
