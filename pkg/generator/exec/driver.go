// Package exec implements a generator generating content by running local executable
package exec

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
)

var _ generator.Interface = (*Driver)(nil)

type Driver struct{}

// New implements generator.Interface
func (*Driver) New(con rt.Conversation, cmd string, params string) (out rt.GeneratorOutput, err error) {
	return
}

// Continue implements generator.Interface
func (*Driver) Continue(con rt.Conversation, cmd string, params string) (out rt.GeneratorOutput, err error) {
	return
}

// Generate implements generator.Interface
func (*Driver) Generate(con rt.Conversation, cmd string, params string, msgs []*rt.Message) (out rt.GeneratorOutput, err error) {
	return
}

// Peek implements generator.Interface
func (*Driver) Peek(con rt.Conversation, msg *rt.Message) (out rt.GeneratorOutput, err error) {
	return
}
