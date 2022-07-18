package rt

import (
	"sync/atomic"
	"time"
)

func NewMessage() *Message {
	return &Message{}
}

type MessageFlag uint32

const (
	MessageFlag_Private MessageFlag = 1 << iota
	MessageFlag_Forwarded
	MessageFlag_Reply
)

func (f MessageFlag) IsPrivate() bool   { return f&MessageFlag_Private != 0 }
func (f MessageFlag) IsForwarded() bool { return f&MessageFlag_Forwarded != 0 }
func (f MessageFlag) IsReply() bool     { return f&MessageFlag_Reply != 0 }

// Message defines a message
type Message struct {
	// ID of the message
	ID      MessageID
	ReplyTo MessageID

	// Flags
	Flags MessageFlag

	// MessageLink is a link to this message
	MessageLink string

	// ChatName is the titile of the chat room/session
	ChatName string

	// ChatLink is the url to the chat room/session
	ChatLink string

	// Author is the sender of the message
	Author string

	// AuthorLink is the link to the sender, usually a url (e.g. https://t.me/joe)
	AuthorLink string

	// when the message was forwarded by the sender, following info describes the original sender
	OriginalChatName    string
	OriginalChatLink    string
	OriginalAuthor      string
	OriginalAuthorLink  string
	OriginalMessageLink string

	Spans []Span

	// Timestamp when the message sent
	Timestamp time.Time

	// Text of all text spans
	Text string

	workers int32

	wait <-chan struct{}
}

func (m *Message) IsForwarded() bool { return m.Flags.IsForwarded() }
func (m *Message) IsPrivate() bool   { return m.Flags.IsPrivate() }
func (m *Message) IsReply() bool     { return m.Flags.IsReply() }

// Ready returns true if the message is ready for content generation
func (m *Message) Ready() bool {
	return atomic.LoadInt32(&m.workers) == 0
}

// Wait returns true until this message is ready, false when canceled
//
// NOTE: can only be called when Ready() returns false
func (m *Message) Wait(until <-chan struct{}, cancelBgJobAfterUntil bool) (jobFinished bool) {
	select {
	case <-until:
		if cancelBgJobAfterUntil {
			// m.cancelBgJob()
		}
		return false
	case <-m.wait:
		return true
	}
}

// Dispose this message
//
// - close cache file (if any)
func (m *Message) Dispose() {
	for i := range m.Spans {
		if m.Spans[i].Data != nil {
			_ = m.Spans[i].Data.Close()
		}
	}
}

type Signal <-chan struct{}

func (m *Message) AddWorker(do func(cancel Signal, m *Message)) {
	atomic.AddInt32(&m.workers, 1)

	go func() {
		defer atomic.AddInt32(&m.workers, -1)

		do(nil, m)
	}()
}
