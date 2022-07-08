package telegram

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	api "arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

var _ message.Interface = (*Message)(nil)

func newTelegramMessage(msg *api.Message, msgs *[]message.Interface) Message {
	return Message{
		ready: 0,

		id: formatMessageID(msg.MessageId),

		msg:  msg,
		msgs: msgs,

		entities: nil,

		mu: &sync.Mutex{},
	}
}

// Message represents a single telegram message
type Message struct {
	ready uint32

	id string

	msg      *api.Message
	entities []message.Entity
	msgs     *[]message.Interface

	mu *sync.Mutex
}

func (m *Message) ID() string {
	return m.id
}

func (m *Message) MessageURL() string {
	url := m.ChatURL()
	if len(url) == 0 {
		return ""
	}

	return url + "/" + formatMessageID(m.msg.MessageId)
}

func (m *Message) Timestamp() time.Time {
	return time.Unix(int64(m.msg.Date), 0).Local()
}

func (m *Message) ChatName() string {
	var name string

	if cfn := m.msg.Chat.FirstName; cfn != nil {
		name = *cfn
	}

	if cln := m.msg.Chat.LastName; cln != nil {
		name += " " + *cln
	}

	return name
}

func (m *Message) ChatURL() string {
	if m.IsPrivateMessage() {
		return ""
	}

	if cu := m.msg.Chat.Username; cu != nil {
		return "https://t.me/" + *cu
	}

	return ""
}

func (m *Message) Author() string {
	if m.msg.From == nil {
		return ""
	}

	name := m.msg.From.FirstName
	if fln := m.msg.From.LastName; fln != nil {
		name += " " + *fln
	}

	return name
}

func (m *Message) AuthorURL() string {
	if m.msg.From == nil {
		return ""
	}

	if fu := m.msg.From.Username; fu != nil {
		return "https://t.me/" + *fu
	}

	return ""
}

func (m *Message) IsForwarded() bool {
	return m.msg.ForwardFrom != nil ||
		m.msg.ForwardFromChat != nil ||
		m.msg.ForwardSenderName != nil ||
		m.msg.ForwardFromMessageId != nil
}

func (m *Message) OriginalMessageURL() string {
	chatURL := m.OriginalChatURL()
	if len(chatURL) == 0 {
		return ""
	}

	if ffmi := m.msg.ForwardFromMessageId; ffmi != nil {
		return chatURL + "/" + strconv.FormatInt(int64(*ffmi), 10)
	}

	return ""
}

func (m *Message) OriginalChatName() string {
	if fc := m.msg.ForwardFromChat; fc != nil {
		var name string
		if fc.FirstName != nil {
			name += *fc.FirstName
		}

		if fc.LastName != nil {
			name += " " + *fc.LastName
		}

		return name
	}

	return ""
}

func (m *Message) OriginalChatURL() string {
	if fc := m.msg.ForwardFromChat; fc != nil && fc.Username != nil {
		return "https://t.me/" + *fc.Username
	}

	return ""
}

func (m *Message) OriginalAuthor() string {
	if ff := m.msg.ForwardFrom; ff != nil {
		name := ff.FirstName
		if ff.LastName != nil {
			name += " " + *ff.LastName
		}

		return name
	}

	return ""
}

func (m *Message) OriginalAuthorURL() string {
	if ff := m.msg.ForwardFrom; ff != nil && ff.Username != nil {
		return "https://t.me/" + *m.msg.ForwardFrom.Username
	}

	return ""
}

func (m *Message) IsPrivateMessage() bool {
	return m.msg.Chat.Type == api.ChatTypePrivate
}

func (m *Message) IsReply() bool {
	return m.msg.ReplyToMessage != nil
}

func (m *Message) ReplyToMessageID() string {
	if m.msg.ReplyToMessage != nil {
		return formatMessageID(m.msg.ReplyToMessage.MessageId)
	}

	return ""
}

func (m *Message) Entities() []message.Entity {
	return m.entities
}

func (m *Message) Messages() []message.Interface {
	if m.msgs != nil {
		return *m.msgs
	}

	return nil
}

// ready for content generation
func (m *Message) Ready() bool {
	return atomic.LoadUint32(&m.ready) == 1
}

func (m *Message) markReady() {
	atomic.StoreUint32(&m.ready, 1)
}

func (m *Message) update(do func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	do()
}
