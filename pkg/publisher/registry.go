package publisher

import "fmt"

type (
	ConfigFactoryFunc func() interface{}
	FactoryFunc       func(config interface{}) (Interface, UserConfig, error)
)

type bundle struct {
	f  FactoryFunc
	cf ConfigFactoryFunc
}

var (
	supportedDrivers = map[string]*bundle{
		"": {
			f: func(interface{}) (Interface, UserConfig, error) {
				return &nop{}, &nopUserConfig{}, nil
			},
			cf: func() interface{} {
				return &nopConfig{}
			},
		},
	}
)

func Register(name string, f FactoryFunc, cf ConfigFactoryFunc) {
	if f == nil || cf == nil {
		return
	}

	// reserve empty name
	if name == "" {
		return
	}

	supportedDrivers[name] = &bundle{
		f:  f,
		cf: cf,
	}
}

func NewConfig(name string) (interface{}, error) {
	b, ok := supportedDrivers[name]
	if !ok {
		return nil, fmt.Errorf("driver %q not found", name)
	}

	return b.cf(), nil
}

func NewDriver(name string, config interface{}) (Interface, UserConfig, error) {
	b, ok := supportedDrivers[name]
	if !ok {
		return nil, nil, fmt.Errorf("driver %q not found", name)
	}

	return b.f(config)
}
