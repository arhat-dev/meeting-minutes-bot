package generator

import (
	"fmt"

	"arhat.dev/meeting-minutes-bot/pkg/message"
)

type Config interface {
	// Create a generation based on this config
	Create() (Interface, error)
}

type Interface interface {
	Name() string

	// RenderPageHeader render page.header
	RenderPageHeader() ([]byte, error)

	// RenderPageBody render page.body
	RenderPageBody(messages []message.Interface) ([]byte, error)
}

type TemplateData struct {
	Messages []message.Interface
}

type FuncMap map[string]interface{}

// Result serves as type handle for arhat.dev/rs
type Result interface {
	// TODO: add methods
}

type configFactoryFunc = func() Config

var (
	supportedDrivers = map[string]configFactoryFunc{
		"": func() Config { return nopConfig{} },
	}
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
