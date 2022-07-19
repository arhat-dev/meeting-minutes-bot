package bot

import (
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

var _ publisher.Interface = (*Driver)(nil)

type Driver struct{}

// Delete implements publisher.Interface
func (*Driver) Delete(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	return
}

// List implements publisher.Interface
func (*Driver) List(con rt.Conversation) (out rt.PublisherOutput, err error) {
	return
}

// Login implements publisher.Interface
func (*Driver) Login(con rt.Conversation, user publisher.User) (out rt.PublisherOutput, err error) {
	return
}

// RequestExternalAccess implements publisher.Interface
func (*Driver) RequestExternalAccess(con rt.Conversation) (out rt.PublisherOutput, err error) {
	return
}

// RequireLogin implements publisher.Interface
func (*Driver) CheckLogin(con rt.Conversation, cmd, params string, user publisher.User) (out rt.PublisherOutput, err error) {
	return
}

// Retrieve implements publisher.Interface
func (*Driver) Retrieve(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	return
}

// AppendToExisting implements publisher.Interface
func (*Driver) AppendToExisting(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	return
}

// CreateNew implements publisher.Interface
func (*Driver) CreateNew(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	return
}
