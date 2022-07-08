package manager

import (
	"sync/atomic"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
)

type Request interface {
	GetMessageIDShouldReplyTo() (uint64, bool)
	SetMessageIDShouldReplyTo(msgID uint64) bool
}

type BaseRequest struct {
	msgIDShouldReplyTo uint64

	wf *bot.Workflow
}

func (b *BaseRequest) Workflow() *bot.Workflow { return b.wf }

func (b *BaseRequest) GetMessageIDShouldReplyTo() (uint64, bool) {
	msgID := atomic.LoadUint64(&b.msgIDShouldReplyTo)
	return msgID, msgID != 0
}

func (b *BaseRequest) SetMessageIDShouldReplyTo(msgID uint64) bool {
	return atomic.CompareAndSwapUint64(&b.msgIDShouldReplyTo, 0, msgID)
}

type (
	SessionRequest struct {
		BaseRequest

		ChatID uint64

		// Topic and URL are mutually exclusive
		Topic string
		URL   string
	}

	EditRequest struct {
		BaseRequest
	}

	ListRequest struct {
		BaseRequest
	}
)

type DeleteRequest struct {
	BaseRequest

	URLs []string
}

func GetCommandFromRequest(req interface{}) bot.BotCmd {
	switch r := req.(type) {
	case *DeleteRequest:
		return bot.BotCmd_Delete
	case *ListRequest:
		return bot.BotCmd_List
	case *EditRequest:
		return bot.BotCmd_Edit
	case *SessionRequest:
		if len(r.Topic) != 0 {
			return bot.BotCmd_Discuss
		}

		return bot.BotCmd_Continue
	default:
		return bot.BotCmd_Unknown
	}
}
