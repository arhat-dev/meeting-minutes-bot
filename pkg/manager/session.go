package manager

import (
	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

func newSession[M message.Interface](wf *bot.Workflow, p publisher.Interface) *Session[M] {
	return &Session[M]{
		wf:        wf,
		publisher: p,

		msgs: make([]M, 0, 16),
	}
}

type Session[M message.Interface] struct {
	// immutable fileds
	wf        *bot.Workflow
	publisher publisher.Interface

	msgs []M
}

func (s *Session[M]) Workflow() *bot.Workflow           { return s.wf }
func (s *Session[M]) GetPublisher() publisher.Interface { return s.publisher }
func (s *Session[M]) RefMessages() *[]M                 { return &s.msgs }

func (s *Session[M]) AppendMessage(msg M) {
	s.msgs = append(s.msgs, msg)
}

func (s *Session[M]) DeleteMessage(msgID uint64) bool {
	// there won't be many messages
	for i := range s.msgs {
		if s.msgs[i].ID() == msgID {
			s.msgs = append(s.msgs[:i], s.msgs[i+1:]...)
			return true
		}
	}

	return false
}

func (s *Session[M]) TruncMessages(n int) {
	if sz := len(s.msgs); n < sz {
		copy(s.msgs, s.msgs[n:])
		s.msgs = s.msgs[:sz-n]
	} else {
		s.msgs = s.msgs[:]
	}
}

func (s *Session[M]) GetMessages() []M { return s.msgs }
