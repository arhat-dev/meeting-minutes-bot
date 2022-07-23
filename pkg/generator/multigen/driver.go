// Package multigen implements a generator
package multigen

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
	"go.uber.org/multierr"
)

var _ generator.Interface = (*Driver)(nil)

type Driver struct {
	underlay []generator.Interface
}

func (d *Driver) forEach(do func(generator.Interface) (rt.GeneratorOutput, error)) (out rt.GeneratorOutput, err error) {
	var (
		tmp  rt.GeneratorOutput
		err2 error
	)

	for _, impl := range d.underlay {
		tmp, err2 = do(impl)
		if err2 != nil {
			err = multierr.Append(err, err2)
			continue
		}

		out.Other = append(out.Other, tmp)
	}

	return
}

// New implements generator.Interface
func (d *Driver) New(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return d.forEach(func(impl generator.Interface) (rt.GeneratorOutput, error) {
		return impl.New(con, in)
	})
}

// Continue implements generator.Interface
func (d *Driver) Continue(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return d.forEach(func(impl generator.Interface) (rt.GeneratorOutput, error) {
		return impl.Continue(con, in)
	})
}

// Generate implements generator.Interface
func (d *Driver) Generate(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return d.forEach(func(impl generator.Interface) (rt.GeneratorOutput, error) {
		return impl.Generate(con, in)
	})
}

// Peek implements generator.Interface
func (d *Driver) Peek(con rt.Conversation, in *rt.GeneratorInput) (out rt.GeneratorOutput, err error) {
	return d.forEach(func(impl generator.Interface) (rt.GeneratorOutput, error) {
		return impl.Peek(con, in)
	})
}
