// Package multipub implements a publisher for publishing generated content through multiple
// publishers
package multipub

import (
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
	"go.uber.org/multierr"
)

// TODO

type Driver struct {
	underlay []pair
}

func forEach(underlay []pair, do func(*pair) (rt.PublisherOutput, error)) (out rt.PublisherOutput, err error) {
	var (
		tmp  rt.PublisherOutput
		err2 error
	)

	for i := range underlay {
		tmp, err2 = do(&underlay[i])
		if err2 != nil {
			err = multierr.Append(err, err2)
			continue
		}

		out.Other = append(out.Other, tmp)
	}

	return
}

// AppendToExisting implements publisher.Interface
func (d *Driver) AppendToExisting(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	return forEach(d.underlay, func(p *pair) (rt.PublisherOutput, error) {
		return p.impl.AppendToExisting(con, cmd, params, in)
	})
}

// CreateNew implements publisher.Interface
func (d *Driver) CreateNew(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	return forEach(d.underlay, func(p *pair) (rt.PublisherOutput, error) {
		return p.impl.CreateNew(con, cmd, params, in)
	})
}

// Delete implements publisher.Interface
func (d *Driver) Delete(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	return forEach(d.underlay, func(p *pair) (rt.PublisherOutput, error) {
		return p.impl.Delete(con, cmd, params)
	})
}

// List implements publisher.Interface
func (d *Driver) List(con rt.Conversation) (out rt.PublisherOutput, err error) {
	return forEach(d.underlay, func(p *pair) (rt.PublisherOutput, error) {
		return p.impl.List(con)
	})
}

// Login implements publisher.Interface
func (d *Driver) Login(con rt.Conversation, user publisher.User) (out rt.PublisherOutput, err error) {
	return forEach(d.underlay, func(p *pair) (rt.PublisherOutput, error) {
		return p.impl.Login(con, p.user)
	})
}

// RequestExternalAccess implements publisher.Interface
func (d *Driver) RequestExternalAccess(con rt.Conversation) (out rt.PublisherOutput, err error) {
	return forEach(d.underlay, func(p *pair) (rt.PublisherOutput, error) {
		return p.impl.RequestExternalAccess(con)
	})
}

// RequireLogin implements publisher.Interface
func (d *Driver) RequireLogin(con rt.Conversation, cmd, params string) (_ rt.LoginFlow, _ error) {
	return
}

// Retrieve implements publisher.Interface
func (d *Driver) Retrieve(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	return forEach(d.underlay, func(p *pair) (rt.PublisherOutput, error) {
		return p.impl.Retrieve(con, cmd, params)
	})
}
