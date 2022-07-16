package storage

import (
	"context"
	"fmt"

	"arhat.dev/meeting-minutes-bot/pkg/rt"
	"arhat.dev/rs"
)

type Interface interface {
	// Name of the storage backend
	Name() string

	// Upload content to the storage
	Upload(
		ctx context.Context, filename string, contentType rt.MIME, in *rt.Input,
	) (url string, err error)
}

type Result interface {
	// URL to fetch this result
	URL() string
}

// Config defines common config methods for storage backend
type Config interface {
	// Create storage based on this config
	Create() (Interface, error)

	MIMEMatch() string
	MaxSize() int64
}

type CommonConfig struct {
	rs.BaseField

	MIMEMatch string `yaml:"mimeMatch"`
	MaxSize   int64  `yaml:"maxSize"`
}

type configFactoryFunc = func() Config

var (
	supportedDrivers = map[string]configFactoryFunc{
		"": func() Config { return &nopConfig{} },
	}
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
