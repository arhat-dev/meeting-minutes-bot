package session

import (
	"sync/atomic"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
)

type Request interface {
	GetMessageIDShouldReplyTo() (rt.MessageID, bool)
	SetMessageIDShouldReplyTo(msgID rt.MessageID) bool
}

var (
	_ Request = (*BaseRequest)(nil)
	_ Request = (*SessionRequest[struct{}])(nil)
	_ Request = (*EditRequest)(nil)
	_ Request = (*ListRequest)(nil)
	_ Request = (*DeleteRequest)(nil)
)

type BaseRequest struct {
	msgIDShouldReplyTo uint64

	wf *bot.Workflow
}

func (b *BaseRequest) Workflow() *bot.Workflow { return b.wf }

func (b *BaseRequest) GetMessageIDShouldReplyTo() (rt.MessageID, bool) {
	msgID := atomic.LoadUint64(&b.msgIDShouldReplyTo)
	return rt.MessageID(msgID), msgID != 0
}

func (b *BaseRequest) SetMessageIDShouldReplyTo(msgID rt.MessageID) bool {
	return atomic.CompareAndSwapUint64(&b.msgIDShouldReplyTo, 0, uint64(msgID))
}

type (
	SessionRequest[T any] struct {
		BaseRequest

		Data T

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

	DeleteRequest struct {
		BaseRequest

		URLs []string
	}
)

func GetCommandFromRequest[T any](req any) bot.BotCmd {
	switch r := req.(type) {
	case *DeleteRequest:
		return bot.BotCmd_Delete
	case *ListRequest:
		return bot.BotCmd_List
	case *EditRequest:
		return bot.BotCmd_Edit
	case *SessionRequest[T]:
		if len(r.Topic) != 0 {
			return bot.BotCmd_Discuss
		}

		return bot.BotCmd_Continue
	default:
		return bot.BotCmd_Unknown
	}
}
