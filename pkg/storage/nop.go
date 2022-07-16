package storage

import (
	"context"

	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

var _ Interface = (*Nop)(nil)

type nopConfig struct{}

func (nopConfig) Create() (Interface, error) { return Nop{}, nil }
func (nopConfig) MIMEMatch() string          { return "" }
func (nopConfig) MaxSize() int64             { return -1 }

type Nop struct{}

func (Nop) Name() string { return "nop" }

func (Nop) Upload(
	ctx context.Context, filename string, contentType rt.MIME, in *rt.Input,
) (url string, err error) {
	return filename, nil
}
