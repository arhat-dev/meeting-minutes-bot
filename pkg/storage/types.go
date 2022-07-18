package storage

import (
	"fmt"

	"arhat.dev/mbot/pkg/rt"
)

// Config defines common config methods for storage backend
type Config interface {
	// Create storage based on this config
	Create() (Interface, error)
}

type Interface interface {
	// Upload content to the storage
	Upload(con rt.Conversation, in *rt.StorageInput) (out rt.StorageOutput, err error)
}

type configFactoryFunc = func() Config

var (
	drivers = map[string]configFactoryFunc{}
)

func Register(name string, cf configFactoryFunc) {
	// reserve empty name
	if name == "" {
		return
	}

	drivers[name] = cf
}

func NewConfig(name string) (any, error) {
	cf, ok := drivers[name]
	if !ok {
		return nil, fmt.Errorf("unknown storage driver %q", name)
	}

	return cf(), nil
}
