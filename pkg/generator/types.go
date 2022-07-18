// Package generator
package generator

import (
	"fmt"

	"arhat.dev/mbot/pkg/rt"
)

type Config interface {
	// Create a generation based on this config
	Create() (Interface, error)
}

// Interface defines public methods of a generator
type Interface interface {
	// New generates content for new conversation
	//
	// cmd is the command triggered this conversation
	// params is the parameters to the cmd
	//
	// this method is called on BotCmd_Discuss
	New(con rt.Conversation, cmd, params string) (out rt.GeneratorOutput, err error)

	// Continue generates content for previously generated content
	//
	// cmd is the command triggered the continuation of the content generation
	// params is the parameters to the cmd
	//
	// this method is called on BotCmd_Continue
	Continue(con rt.Conversation, cmd, params string) (out rt.GeneratorOutput, err error)

	// Peek a newly received message
	Peek(con rt.Conversation, msg *rt.Message) (out rt.GeneratorOutput, err error)

	// Generate generates the body of this conversation
	//
	// this method is called on BotCmd_End
	Generate(con rt.Conversation, cmd, params string, msgs []*rt.Message) (out rt.GeneratorOutput, err error)
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

func NewConfig(name string) (Config, error) {
	cf, ok := drivers[name]
	if !ok {
		return nil, fmt.Errorf("unknown generator driver %q", name)
	}

	return cf(), nil
}
