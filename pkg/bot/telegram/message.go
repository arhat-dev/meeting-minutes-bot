package telegram

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	api "arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

var _ message.Interface = (*telegramMessage)(nil)

func newTelegramMessage(msg *api.Message, msgs *[]message.Interface) *telegramMessage {
	return &telegramMessage{
		id: formatMessageID(msg.MessageId),

		msg:  msg,
		msgs: msgs,

		entities: nil,

		ready: 0,

		mu: &sync.Mutex{},
	}
}

type telegramMessage struct {
	id string

	msg      *api.Message
	entities []message.Entity
	msgs     *[]message.Interface

	ready uint32

	mu *sync.Mutex
}

func (m *telegramMessage) ID() string {
	return m.id
}

func (m *telegramMessage) MessageURL() string {
	url := m.ChatURL()
	if len(url) == 0 {
		return ""
	}

	return url + "/" + formatMessageID(m.msg.MessageId)
}

func (m *telegramMessage) Timestamp() time.Time {
	return time.Unix(int64(m.msg.Date), 0).Local()
}

func (m *telegramMessage) ChatName() string {
	var name string

	if cfn := m.msg.Chat.FirstName; cfn != nil {
		name = *cfn
	}

	if cln := m.msg.Chat.LastName; cln != nil {
		name += " " + *cln
	}

	return name
}

func (m *telegramMessage) ChatURL() string {
	if m.IsPrivateMessage() {
		return ""
	}

	if cu := m.msg.Chat.Username; cu != nil {
		return "https://t.me/" + *cu
	}

	return ""
}

func (m *telegramMessage) Author() string {
	if m.msg.From == nil {
		return ""
	}

	name := m.msg.From.FirstName
	if fln := m.msg.From.LastName; fln != nil {
		name += " " + *fln
	}

	return name
}

func (m *telegramMessage) AuthorURL() string {
	if m.msg.From == nil {
		return ""
	}

	if fu := m.msg.From.Username; fu != nil {
		return "https://t.me/" + *fu
	}

	return ""
}

func (m *telegramMessage) IsForwarded() bool {
	return m.msg.ForwardFrom != nil ||
		m.msg.ForwardFromChat != nil ||
		m.msg.ForwardSenderName != nil ||
		m.msg.ForwardFromMessageId != nil
}

func (m *telegramMessage) OriginalMessageURL() string {
	chatURL := m.OriginalChatURL()
	if len(chatURL) == 0 {
		return ""
	}

	if ffmi := m.msg.ForwardFromMessageId; ffmi != nil {
		return chatURL + "/" + strconv.FormatInt(int64(*ffmi), 10)
	}

	return ""
}

func (m *telegramMessage) OriginalChatName() string {
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

func (m *telegramMessage) OriginalChatURL() string {
	if fc := m.msg.ForwardFromChat; fc != nil && fc.Username != nil {
		return "https://t.me/" + *fc.Username
	}

	return ""
}

func (m *telegramMessage) OriginalAuthor() string {
	if ff := m.msg.ForwardFrom; ff != nil {
		name := ff.FirstName
		if ff.LastName != nil {
			name += " " + *ff.LastName
		}

		return name
	}

	return ""
}

func (m *telegramMessage) OriginalAuthorURL() string {
	if ff := m.msg.ForwardFrom; ff != nil && ff.Username != nil {
		return "https://t.me/" + *m.msg.ForwardFrom.Username
	}

	return ""
}

func (m *telegramMessage) IsPrivateMessage() bool {
	return m.msg.Chat.Type == api.ChatTypePrivate
}

func (m *telegramMessage) IsReply() bool {
	return m.msg.ReplyToMessage != nil
}

func (m *telegramMessage) ReplyToMessageID() string {
	if m.msg.ReplyToMessage != nil {
		return formatMessageID(m.msg.ReplyToMessage.MessageId)
	}

	return ""
}

func (m *telegramMessage) Entities() []message.Entity {
	return m.entities
}

func (m *telegramMessage) Messages() []message.Interface {
	if m.msgs != nil {
		return *m.msgs
	}

	return nil
}

// ready for content generation
func (m *telegramMessage) Ready() bool {
	return atomic.LoadUint32(&m.ready) == 1
}

func (m *telegramMessage) markReady() {
	atomic.StoreUint32(&m.ready, 1)
}

func (m *telegramMessage) update(do func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	do()
}
