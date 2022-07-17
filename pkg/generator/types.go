package generator

import (
	"fmt"

	"arhat.dev/mbot/pkg/rt"
)

type Config interface {
	// Create a generation based on this config
	Create() (Interface, error)
}

// Output is the type handle for output of a generator
type Output interface {
	// TODO: add methods
}

type Interface interface {
	// RenderPageHeader render page.header
	RenderPageHeader() (string, error)

	// RenderPageBody render page.body
	RenderPageBody(messages []*rt.Message) (string, error)
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

func NewConfig(name string) (Config, error) {
	cf, ok := supportedDrivers[name]
	if !ok {
		return nil, fmt.Errorf("unknown generator driver %q", name)
	}

	return cf(), nil
}
