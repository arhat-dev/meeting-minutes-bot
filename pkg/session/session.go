package session

import (
	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

func newSession(wf *bot.Workflow, p publisher.Interface) *Session {
	return &Session{
		wf:        wf,
		publisher: p,

		msgs: make([]*rt.Message, 0, 16),
	}
}

type Session struct {
	// immutable fileds
	wf        *bot.Workflow
	publisher publisher.Interface

	msgs []*rt.Message
}

func (s *Session) Workflow() *bot.Workflow           { return s.wf }
func (s *Session) GetPublisher() publisher.Interface { return s.publisher }
func (s *Session) RefMessages() *[]*rt.Message       { return &s.msgs }

func (s *Session) AppendMessage(msg *rt.Message) {
	s.msgs = append(s.msgs, msg)
}

func (s *Session) DeleteMessage(msgID rt.MessageID) bool {
	// there won't be many messages
	for i := range s.msgs {
		if s.msgs[i].ID == msgID {
			s.msgs = append(s.msgs[:i], s.msgs[i+1:]...)
			return true
		}
	}

	return false
}

func (s *Session) TruncMessages(n int) {
	if sz := len(s.msgs); n < sz {
		copy(s.msgs, s.msgs[n:])
		s.msgs = s.msgs[:sz-n]
	} else {
		s.msgs = s.msgs[:]
	}
}

func (s *Session) GetMessages() []*rt.Message { return s.msgs }
