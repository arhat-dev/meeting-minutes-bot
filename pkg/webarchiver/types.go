package webarchiver

import "context"

type Config interface {
	// Create webarchiver based on this config
	Create() (Interface, error)
}

type Interface interface {
	// Archive web page, return url of the archived page and screenshot
	Archive(ctx context.Context, url string) (
		archiveURL string,
		screenshot []byte,
		err error,
	)
}
