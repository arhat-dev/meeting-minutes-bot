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

		Params string
		IsNew  bool
	}

	EditRequest struct {
		BaseRequest
	}

	ListRequest struct {
		BaseRequest
	}

	DeleteRequest struct {
		BaseRequest

		Params string
	}
)

func GetCommandFromRequest[T any](req any) rt.BotCmd {
	switch r := req.(type) {
	case *DeleteRequest:
		return rt.BotCmd_Delete
	case *ListRequest:
		return rt.BotCmd_List
	case *EditRequest:
		return rt.BotCmd_Edit
	case *SessionRequest[T]:
		if len(r.Params) != 0 {
			return rt.BotCmd_New
		}

		return rt.BotCmd_Resume
	default:
		return rt.BotCmd_Unknown
	}
}
