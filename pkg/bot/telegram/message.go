package telegram

import (
	"strings"
	"sync/atomic"
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
	"github.com/gotd/td/tg"
)

var _ message.Interface = (*Message)(nil)

type msgFlag uint32

const (
	msgFlag_Private = 1 << iota
	msgFlag_Reply
	msgFlag_Fwd
)

func (m msgFlag) IsPrivateMessage() bool { return m&msgFlag_Private != 0 }
func (m msgFlag) IsReply() bool          { return m&msgFlag_Reply != 0 }
func (m msgFlag) IsForwarded() bool      { return m&msgFlag_Fwd != 0 }

func newTelegramMessage(src *messageSource, msg *tg.Message, msgs *[]*Message) (ret Message) {
	var (
		buf strings.Builder
	)

	ret.id = uint64(msg.GetID())
	// TODO: set tz by user location
	ret.timestamp = time.Unix(int64(msg.GetDate()), 0).Local()
	ret.msgs = msgs

	if src.Chat.IsPrivateChat() {
		ret.msgFlag |= msgFlag_Private
	}

	if replyTo, ok := msg.GetReplyTo(); ok {
		ret.msgFlag |= msgFlag_Reply
		ret.replyToMessageID = uint64(replyTo.GetReplyToMsgID())
	}

	if !src.FwdFrom.IsNil() {
		ret.msgFlag |= msgFlag_Fwd

		ret.origAuthor = src.FwdFrom.GetPtr().Title()
		if len(ret.origAuthor) == 0 {
			buf.Reset()
			buf.WriteString(src.FwdFrom.GetPtr().Firstname())
			if buf.Len() != 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(src.FwdFrom.GetPtr().Lastname())
			ret.origAuthor = buf.String()
		}

		if len(src.FwdFrom.GetPtr().Username()) != 0 {
			buf.Reset()
			buf.WriteString("https://t.me/")
			buf.WriteString(src.FwdFrom.GetPtr().Username())
			ret.origAuthorLink = buf.String()
		}
	}

	ret.chatName = src.Chat.Title()
	if len(ret.chatName) == 0 {
		buf.Reset()
		buf.WriteString(src.Chat.Firstname())
		if buf.Len() != 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(src.Chat.Lastname())

		ret.chatName = buf.String()
	}

	if !ret.IsPrivateMessage() && len(src.Chat.Username()) != 0 {
		buf.Reset()
		buf.WriteString("https://t.me/")
		buf.WriteString(src.Chat.Username())
		ret.chatURL = buf.String()

		buf.WriteString("/")
		buf.WriteString(formatMessageID(ret.id))
		ret.msgURL = buf.String()
	}

	if !src.From.IsNil() {
		ret.author = src.From.GetPtr().Title()

		if len(ret.author) != 0 {
			buf.Reset()
			buf.WriteString(src.From.GetPtr().Firstname())
			if buf.Len() != 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(src.From.GetPtr().Lastname())
			ret.author = buf.String()
		}
	}

	if !src.From.IsNil() && len(src.From.GetPtr().Username()) != 0 {
		buf.Reset()
		buf.WriteString("https://t.me/")
		buf.WriteString(src.From.GetPtr().Username())
		ret.authorURL = buf.String()
	}

	if !src.FwdChat.IsNil() {
		ret.origChatName = src.FwdChat.GetPtr().Title()
		if len(ret.origChatName) == 0 {
			buf.Reset()
			buf.WriteString(src.FwdChat.GetPtr().Firstname())
			if buf.Len() != 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(src.FwdChat.GetPtr().Lastname())
			ret.origChatName = buf.String()
		}

		if len(src.FwdChat.GetPtr().Username()) != 0 {
			buf.Reset()
			buf.WriteString("https://t.me/")
			buf.WriteString(src.FwdChat.GetPtr().Username())
			ret.origChatLink = buf.String()

			if fwdHdr, ok := msg.GetFwdFrom(); ok {
				fwdMsgID, ok := fwdHdr.GetSavedFromMsgID()
				if ok {
					buf.WriteString("/")
					buf.WriteString(formatMessageID(fwdMsgID))
					ret.origMsgURL = buf.String()
				}
			}
		}
	}

	return
}

// Message represents a single telegram message
type Message struct {
	ready uint32
	msgFlag

	id               uint64
	replyToMessageID uint64

	chatName, chatURL string
	author, authorURL string
	msgURL            string

	origChatName, origChatLink string
	origAuthor, origAuthorLink string
	origMsgURL                 string

	timestamp time.Time

	// temporary buffer for background worker, merged into entities when markReady() called
	nonTextEntity rt.Optional[message.Span]
	textEntities  []message.Span

	entities []message.Span

	msgs *[]*Message

	cancelBgJob func()
	wait        chan struct{}
}

func (m *Message) ID() uint64           { return m.id }
func (m *Message) Timestamp() time.Time { return m.timestamp }

func (m *Message) ChatName() string    { return m.chatName }
func (m *Message) ChatLink() string    { return m.chatURL }
func (m *Message) MessageLink() string { return m.msgURL }
func (m *Message) Author() string      { return m.author }
func (m *Message) AuthorLink() string  { return m.authorURL }

func (m *Message) OriginalChatName() string    { return m.origChatName }
func (m *Message) OriginalChatLink() string    { return m.origChatLink }
func (m *Message) OriginalMessageLink() string { return m.origMsgURL }
func (m *Message) OriginalAuthor() string      { return m.origAuthor }
func (m *Message) OriginalAuthorLink() string  { return m.origAuthorLink }

func (m *Message) ReplyToMessageID() uint64 { return m.replyToMessageID }

func (m *Message) Entities() []message.Span { return m.entities }

func (m *Message) Ready() bool { return atomic.LoadUint32(&m.ready) == 2 }

func (m *Message) Wait(cancel <-chan struct{}) (done bool) {
	select {
	case <-cancel:
		m.cancelBgJob()
		return false
	case <-m.wait:
		return true
	}
}

func (m *Message) Dispose() {
	for i := range m.entities {
		if m.entities[i].Data != nil {
			_ = m.entities[i].Data.Close()
		}
	}
}

func (m *Message) markReady() {
	if atomic.CompareAndSwapUint32(&m.ready, 0, 1) {
		if m.nonTextEntity.IsNil() {
			m.entities = m.textEntities
		} else {
			m.entities = append(m.textEntities, m.nonTextEntity.Get())
		}

		m.textEntities = nil
		atomic.StoreUint32(&m.ready, 2)
	}
}
