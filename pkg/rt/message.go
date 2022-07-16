package rt

import (
	"sync/atomic"
	"time"
)

// Message defines a message
type Message struct {
	// ID of the message
	ID MessageID

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
	IsForwarded         bool
	OriginalChatName    string
	OriginalChatLink    string
	OriginalAuthor      string
	OriginalAuthorLink  string
	OriginalMessageLink string

	IsPrivateMessage bool

	IsReply          bool
	ReplyToMessageID MessageID

	Spans []Span

	// Timestamp when the message sent
	Timestamp time.Time

	ready     uint32
	mediaSpan *Span
	textSpans []Span

	cancelBgJob func()
	wait        <-chan struct{}
}

const (
	readyState_NotReady = iota
	readyState_Finalizing
	readyState_Ready
)

// Ready returns true if the message is ready for content generation
func (m *Message) Ready() bool {
	return atomic.LoadUint32(&m.ready) == readyState_Ready
}

// Wait returns true until this message is ready, false when canceled
//
// NOTE: can only be called when Ready() returns false
func (m *Message) Wait(until <-chan struct{}, cancelBgJobAfterUntil bool) (jobFinished bool) {
	select {
	case <-until:
		if cancelBgJobAfterUntil {
			m.cancelBgJob()
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

func (m *Message) MarkReady() {
	if atomic.CompareAndSwapUint32(&m.ready, 0, readyState_Finalizing) {
		if m.mediaSpan == nil {
			m.Spans = m.textSpans
		} else if len(m.textSpans) != 0 {
			m.Spans = append(m.textSpans, *m.mediaSpan)
		} else {
			m.Spans = []Span{*m.mediaSpan}
		}

		m.textSpans = nil
		m.mediaSpan = nil

		atomic.StoreUint32(&m.ready, readyState_Ready)
	}
}

// SetMediaSpan
//
// NOTE: MUST be called before MarkReady() if there is media span
func (m *Message) SetMediaSpan(media *Span) { m.mediaSpan = media }

// SetTextSpans
//
// NOTE: MUST be called before MarkReady() if there is text spans
func (m *Message) SetTextSpans(s []Span) { m.textSpans = s }

// NOTE: MUST be called before Wait()
func (m *Message) SetBackgroundJob(wait <-chan struct{}, cancel func()) {
	m.wait = wait
	m.cancelBgJob = cancel
}
