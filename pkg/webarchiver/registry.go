package webarchiver

import "fmt"

type (
	ConfigFactoryFunc func() Config
)

var (
	supportedDrivers = map[string]ConfigFactoryFunc{
		"": func() Config { return &NopConfig{} },
	}
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
