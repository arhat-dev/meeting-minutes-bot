// Package chain implements a generator get multiple generators chained into one
package chain

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
)

var _ generator.Interface = (*Driver)(nil)

type Driver struct{}

// Peek implements generator.Interface
func (*Driver) Peek(con rt.Conversation, msg *rt.Message) (out rt.GeneratorOutput, err error) {
	return
}

// Continue implements generator.Interface
func (*Driver) Continue(con rt.Conversation, cmd string, params string) (out rt.GeneratorOutput, err error) {
	return
}

// New implements generator.Interface
func (*Driver) New(con rt.Conversation, cmd string, params string) (out rt.GeneratorOutput, err error) {
	return
}

// RenderBody implements generator.Interface
func (*Driver) Generate(con rt.Conversation, cmd string, params string, msgs []*rt.Message) (out rt.GeneratorOutput, err error) {
	return
}
