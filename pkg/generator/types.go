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
	// RenderNew generates content for new conversation
	//
	// cmd is the command triggered this conversation
	// params is the parameters to the cmd
	//
	// this is called on BotCmd_Discuss
	New(con rt.Conversation, cmd, params string) (string, error)

	// Continue generates content for previously generated content
	//
	// cmd is the command triggered the continuation of the content generation
	// params is the parameters to the cmd
	//
	// this is called on BotCmd_Continue
	Continue(con rt.Conversation, cmd, params string) (string, error)

	// RenderBody generates the body of this conversation
	//
	// this is called when the conversation ended
	//
	// this is called on BotCmd_End
	RenderBody(con rt.Conversation, msgs []*rt.Message) (string, error)
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
