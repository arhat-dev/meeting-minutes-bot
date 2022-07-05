package storage

import "context"

type Config interface {
	// Create storage based on this config
	Create() (Interface, error)
}

type Interface interface {
	Name() string

	Upload(
		ctx context.Context,
		filename string,
		contentType string,
		data []byte,
	) (url string, err error)
}
