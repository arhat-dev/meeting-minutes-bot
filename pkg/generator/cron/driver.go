// Package cron implements a generator works in background to generate content periodically.
package cron

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
)

var _ generator.Interface = (*Driver)(nil)

type Driver struct{}

// New implements generator.Interface
func (*Driver) New(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return
}

// Continue implements generator.Interface
func (*Driver) Continue(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return
}

// Peek implements generator.Interface
func (*Driver) Peek(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return
}

// RenderBody implements generator.Interface
func (*Driver) Generate(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return
}
