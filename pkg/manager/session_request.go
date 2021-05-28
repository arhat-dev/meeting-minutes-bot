package manager

import (
	"sync/atomic"

	"arhat.dev/meeting-minutes-bot/pkg/constant"
)

type Request interface {
	GetMessageIDShouldReplyTo() (uint64, bool)
	SetMessageIDShouldReplyTo(msgID uint64) bool
}

type BaseRequest struct {
	msgIDShouldReplyTo uint64
}

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

		// only one of topic and url shall be specified
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

func GetCommandFromRequest(req interface{}) string {
	switch r := req.(type) {
	case *DeleteRequest:
		return constant.CommandDelete
	case *ListRequest:
		return constant.CommandList
	case *EditRequest:
		return constant.CommandEdit
	case *SessionRequest:
		if len(r.Topic) != 0 {
			return constant.CommandDiscuss
		}

		return constant.CommandContinue
	}

	return ""
}
