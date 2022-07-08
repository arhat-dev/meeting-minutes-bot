package message

import (
	"time"
)

// Interface defines the set of public methods for a message
type Interface interface {
	ID() string
	MessageURL() string

	// Timestamp when the message sent
	Timestamp() time.Time

	// ChatName is the titile of the chat room/session
	ChatName() string

	// ChatURL is the url to the chat room/session
	ChatURL() string

	// Author is the sender of the message
	Author() string

	// AuthorURL is the link to the sender, usually a url (e.g. https://t.me/joe)
	AuthorURL() string

	// when the message was forwarded by the sender, following info describes the original sender
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
