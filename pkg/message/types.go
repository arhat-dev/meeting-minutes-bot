package message

import (
	"time"
)

type Interface interface {
	ID() string
	MessageURL() string

	Timestamp() time.Time

	ChatName() string
	ChatURL() string

	Author() string
	AuthorURL() string

	// message forwarded from other chat, use following info
	IsForwarded() bool
	OriginalChatName() string
	OriginalChatURL() string
	OriginalAuthor() string
	OriginalAuthorURL() string
	OriginalMessageURL() string

	IsPrivateMessage() bool

	IsReply() bool
	ReplyToMessageID() string

	Entities() []Entity

	// Ready returns true if the message has been pre-processed
	Ready() bool

	Messages() []Interface
}
