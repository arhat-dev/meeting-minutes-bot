package bot

import (
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
}

// AppendToExisting implements publisher.Interface
func (*Driver) AppendToExisting(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) ([]rt.Span, error) {
	panic("unimplemented")
}

// CreateNew implements publisher.Interface
func (*Driver) CreateNew(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) ([]rt.Span, error) {
	panic("unimplemented")
}

func (d Driver) RequireLogin(con rt.Conversation, cmd, params string) (rt.LoginFlow, error) {
	return rt.LoginFlow_None, nil
}

func (d *Driver) Login(con rt.Conversation, user publisher.User) (_ []rt.Span, _ error) { return }

func (d *Driver) RequestExternalAccess(con rt.Conversation) (_ []rt.Span, err error) { return }

func (d *Driver) Retrieve(con rt.Conversation, cmd, params string) (_ []rt.Span, err error) { return }

func (d *Driver) List(con rt.Conversation) (_ []publisher.PostInfo, err error) { return }

func (d *Driver) Delete(con rt.Conversation, cmd, params string) (err error) { return }
