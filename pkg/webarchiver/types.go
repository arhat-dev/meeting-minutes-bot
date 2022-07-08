package webarchiver

import (
	"context"
	"io"
)

type Config interface {
	// Create webarchiver based on this config
	Create() (Interface, error)
}

type Interface interface {
	// Archive web page
	// TODO: support full web request context
	Archive(ctx context.Context, url string) (Result, error)
}

// Result of a web archive operation
type Result interface {
	// WARC get archived .warc file
	//
	// ref: https://en.wikipedia.org/wiki/Web_ARChive
	WARC() (data io.ReadSeekCloser, size int64)

	// Screenshot get archived bitmap data
	Screenshot() (data io.ReadSeekCloser, size int64)
}
