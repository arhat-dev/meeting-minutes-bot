package rttest

import (
	"context"

	"arhat.dev/mbot/pkg/rt"
)

func FakeConversation(ctx context.Context) rt.Conversation {
	return &fakeConversation{ctx: ctx}
}

var _ rt.Conversation = (*fakeConversation)(nil)

type fakeConversation struct {
	ctx context.Context
}

// Context implements rt.Conversation
func (c *fakeConversation) Context() context.Context {
	return c.ctx
}

// SendMessage implements rt.Conversation
func (c *fakeConversation) SendMessage(ctx context.Context, opts rt.SendMessageOptions) ([]rt.MessageID, error) {
	return nil, nil
}
