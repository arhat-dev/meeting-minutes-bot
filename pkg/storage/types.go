package storage

import (
	"fmt"

	"arhat.dev/mbot/internal/mime"
	"arhat.dev/mbot/pkg/rt"
)

type Interface interface {
	// Upload content to the storage
	Upload(con rt.Conversation, filename string, contentType mime.MIME, in *rt.Input) (url string, err error)
}

type Result interface {
	// URL to fetch this result
	URL() string
}

// Config defines common config methods for storage backend
type Config interface {
	// Create storage based on this config
	Create() (Interface, error)
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

func NewConfig(name string) (any, error) {
	cf, ok := supportedDrivers[name]
	if !ok {
		return nil, fmt.Errorf("unknown storage driver %q", name)
	}

	return cf(), nil
}
