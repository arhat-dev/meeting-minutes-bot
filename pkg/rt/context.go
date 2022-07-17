package rt

import (
	"context"

	"arhat.dev/pkg/log"
)

func NewContext(ctx context.Context, logger log.Interface, cache Cache) RTContext {
	return RTContext{
		ctx:    ctx,
		logger: logger,
		cache:  cache,
	}
}

type RTContext struct {
	ctx    context.Context
	logger log.Interface
	cache  Cache
}

func (r *RTContext) Context() context.Context { return r.ctx }
func (r *RTContext) Logger() log.Interface    { return r.logger }
func (r *RTContext) Cache() Cache             { return r.cache }

type MessageCallbackSpec struct {
	Text string

	URL     Optional[string]
	OnClick Optional[func() error]
}

type SendMessageOptions struct {
	ReplyTo MessageID

	NoForward      bool
	NoNotification bool
	NoWebPreview   bool
	_              bool

	// MessageBody
	MessageBody []Span

	// Callbacks for user interactions
	Callbacks [][]MessageCallbackSpec
}

// Conversation is the chat context currently triggering bot reaction
type Conversation interface {
	// Context of this conversation
	Context() context.Context

	// SendMessage to this conversation
	SendMessage(ctx context.Context, opts SendMessageOptions) ([]MessageID, error)
}
