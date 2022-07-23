// Package chain implements a generator get multiple generators chained into one
package chain

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
)

var _ generator.Interface = (*Driver)(nil)

type Driver struct {
	underlay []generator.Interface
}

func forEach(
	underlay []generator.Interface,
	in rt.GeneratorInput,
	do func(impl generator.Interface, in *rt.GeneratorInput) (rt.GeneratorOutput, error),
) (out rt.GeneratorOutput, err error) {
	for _, impl := range underlay {
		out, err = do(impl, &in)
		if err != nil {
			return
		}

		in.Messages = out.Messages
	}

	return
}

// Peek implements generator.Interface
func (d *Driver) Peek(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return forEach(d.underlay, *in, func(impl generator.Interface, in *rt.GeneratorInput) (rt.GeneratorOutput, error) {
		return impl.Peek(con, in)
	})
}

// Continue implements generator.Interface
func (d *Driver) Continue(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return forEach(d.underlay, *in, func(impl generator.Interface, in *rt.GeneratorInput) (rt.GeneratorOutput, error) {
		return impl.Continue(con, in)
	})
}

// New implements generator.Interface
func (d *Driver) New(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return forEach(d.underlay, *in, func(impl generator.Interface, in *rt.GeneratorInput) (rt.GeneratorOutput, error) {
		return impl.New(con, in)
	})
}

// RenderBody implements generator.Interface
func (d *Driver) Generate(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return forEach(d.underlay, *in, func(impl generator.Interface, in *rt.GeneratorInput) (rt.GeneratorOutput, error) {
		return impl.Generate(con, in)
	})
}
