package storage

import (
	"context"
	"io"
)

var _ Interface = (*Nop)(nil)

type nopConfig struct{}

func (nopConfig) Create() (Interface, error) { return Nop{}, nil }
func (nopConfig) MIMEMatch() string          { return "" }
func (nopConfig) MaxSize() int64             { return -1 }

type Nop struct{}

func (Nop) Name() string { return "nop" }

func (Nop) Upload(
	ctx context.Context, filename, contentType string, size int64, data io.Reader,
) (url string, err error) {
	return filename, nil
}
