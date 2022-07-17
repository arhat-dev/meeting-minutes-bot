package publisher

import (
	"context"
	"fmt"

	"arhat.dev/mbot/pkg/rt"
)

type Config interface {
	// Create a publisher base on this config
	Create() (Interface, UserConfig, error)
}

type Feature uint32

const (
	Feature_Login  Feature = 1 << iota // needs login
	Feature_SignUp                     // needs sign up
	Feature_Edit                       // supports editing
	Feature_List                       // supports listing content generated by user
	Feature_Delete                     // supports deleting
)

type Interface interface {
	// RequireLogin return true when the publisher requires login, if false
	// there will be no login process presented to user
	RequireLogin() bool

	// Login to platform
	Login(config UserConfig) (token string, _ error)

	// AuthURL returns a clickable url for external authorization
	AuthURL() (string, error)

	// Retrieve post and cache it locally according to the url
	Retrieve(url string) ([]rt.Span, error)

	// Publish a new post
	Publish(title string, body *rt.Input) ([]rt.Span, error)

	// List all posts for this user
	List() ([]PostInfo, error)

	// Delete one post according to the url
	Delete(urls ...string) error

	// Append content to local post cache
	Append(ctx context.Context, body *rt.Input) ([]rt.Span, error)
}

type PostInfo struct {
	Title string
	URL   string
}

type UserConfig interface {
	SetAuthToken(token string)
}

// Result serves as type handle for arhat.dev/rs
type Result interface {
	// TODO: add methods
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
		return nil, fmt.Errorf("unknown publisher driver %q", name)
	}

	return cf(), nil
}
