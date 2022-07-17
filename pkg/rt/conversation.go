package rt

import "context"

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

	// Body is the message body
	Body []Span

	// Callbacks for user interaction
	Callbacks [][]MessageCallbackSpec
}

// Conversation is the chat context currently triggering bot reaction
type Conversation interface {
	// Context of this conversation
	Context() context.Context

	// SendMessage to this conversation
	SendMessage(ctx context.Context, opts SendMessageOptions) ([]MessageID, error)
}
