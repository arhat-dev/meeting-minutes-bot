package storage

import "context"

type Interface interface {
	Name() string

	Upload(ctx context.Context, filename string, data []byte) (url string, err error)
}
