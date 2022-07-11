package message

import (
	"time"
)

// Interface defines a message
type Interface interface {
	// ID of the message
	ID() uint64

	// MessageLink is a link to this message
	MessageLink() string

	// Timestamp when the message sent
	Timestamp() time.Time

	// ChatName is the titile of the chat room/session
	ChatName() string

	// ChatLink is the url to the chat room/session
	ChatLink() string

	// Author is the sender of the message
	Author() string

	// AuthorLink is the link to the sender, usually a url (e.g. https://t.me/joe)
	AuthorLink() string

	// when the message was forwarded by the sender, following info describes the original sender
	IsForwarded() bool
	OriginalChatName() string
	OriginalChatLink() string
	OriginalAuthor() string
	OriginalAuthorLink() string
	OriginalMessageLink() string

	IsPrivateMessage() bool

	IsReply() bool
	ReplyToMessageID() uint64

	Entities() []Span

	// Ready returns true if the message is ready for content generation
	Ready() bool

	// Wait returns true until this message is ready, false when canceled
	Wait(cancel <-chan struct{}) (done bool)

	// Dispose this message
	//
	// - close cache file (if any)
	Dispose()
}
