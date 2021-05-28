package manager

import (
	"sync"
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

func newSession(p publisher.Interface) *Session {
	return &Session{
		msgs: make([]message.Interface, 0, 16),

		publisher: p,
		msgIdx:    make(map[string]int),
		mu:        &sync.RWMutex{},
	}
}

type Session struct {
	msgs []message.Interface

	publisher publisher.Interface
	msgIdx    map[string]int
	mu        *sync.RWMutex
}

func (s *Session) GetPublisher() publisher.Interface {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.publisher
}

func (s *Session) RefMessages() *[]message.Interface {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return &s.msgs
}

func (s *Session) AppendMessage(msg message.Interface) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.msgs = append(s.msgs, msg)
	s.msgIdx[msg.ID()] = len(s.msgs) - 1
}

func (s *Session) DeleteMessage(msgID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, ok := s.msgIdx[msgID]
	if !ok {
		// no such id, ignore
		return false
	}

	delete(s.msgIdx, msgID)
	s.msgs = append(s.msgs[:idx], s.msgs[idx+1:]...)
	return true
}

func (s *Session) DeleteFirstNMessage(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if n < len(s.msgs) {
		for _, msg := range s.msgs[:n] {
			delete(s.msgIdx, msg.ID())
		}

		s.msgs = s.msgs[n:]
	} else {
		s.msgs = make([]message.Interface, 0, 16)
		s.msgIdx = make(map[string]int)
	}
}

func (s *Session) GenerateContent(gen generator.Interface) (msgOutCount int, _ []byte, _ error) {
	s.mu.RLock()
	msgOutCount = len(s.msgs)

	msgCopy := make([]message.Interface, msgOutCount)
	_ = copy(msgCopy, s.msgs)

	s.mu.RUnlock()

	// ensure every message is ready
	for _, m := range msgCopy {
		if !m.Ready() {
			// TODO: log output
			time.Sleep(1 * time.Second)
		}
	}

	result, err := gen.RenderPageBody(msgCopy)

	return msgOutCount, result, err
}
