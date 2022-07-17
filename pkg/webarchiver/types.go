package webarchiver

import (
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
	Archive(con rt.Conversation, url string) (Result, error)
}

// Result of a web archive operation
type Result struct {
	SizeWARC       int64
	SizeScreenshot int64

	// re:warc: https://en.wikipedia.org/wiki/Web_ARChive
	WARC       rt.CacheReader
	Screenshot rt.CacheReader
}

type configFactoryFunc = func() Config

var (
	supportedDrivers = map[string]configFactoryFunc{}
)

func Register(name string, cf configFactoryFunc) {
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
