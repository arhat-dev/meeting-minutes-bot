package webarchiver

import (
	"context"
	"fmt"

	"arhat.dev/mbot/pkg/rt"
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
	WARC() (data rt.CacheReader, size int64)

	// Screenshot get archived bitmap data
	Screenshot() (data rt.CacheReader, size int64)
}

type (
	ConfigFactoryFunc func() Config
)

var (
	supportedDrivers = map[string]ConfigFactoryFunc{}
)

func Register(name string, cf ConfigFactoryFunc) {
	// reserve empty name
	if name == "" {
		return
	}

	supportedDrivers[name] = cf
}

func NewConfig(name string) (interface{}, error) {
	cf, ok := supportedDrivers[name]
	if !ok {
		return nil, fmt.Errorf("driver %q not found", name)
	}

	return cf(), nil
}
