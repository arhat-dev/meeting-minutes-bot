package server

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"arhat.dev/pkg/queue"

	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

func newSession(topic, defaultChatUsername string, gen generator.Interface) *session {
	return &session{
		Topic:    topic,
		Messages: make([]message.Interface, 0, 16),

		defaultChatUsername: defaultChatUsername,
		generator:           gen,
		msgIdx:              make(map[string]int),
		mu:                  &sync.RWMutex{},
	}
}

type session struct {
	Topic    string
	Messages []message.Interface

	defaultChatUsername string
	generator           generator.Interface
	msgIdx              map[string]int
	mu                  *sync.RWMutex
}

func (s *session) peekLastMessage() Message {
	s.mu.RLock()
	defer s.mu.RUnlock()

	size := len(s.Messages)
	if size == 0 {
		return nil
	}

	return s.Messages[size-1].(Message)
}

func (s *session) appendMessage(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Messages = append(s.Messages, msg)
	s.msgIdx[msg.ID()] = len(s.Messages) - 1
}

func (s *session) deleteMessage(msgID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, ok := s.msgIdx[msgID]
	if !ok {
		// no such id, ignore
		return false
	}

	delete(s.msgIdx, msgID)
	s.Messages = append(s.Messages[:idx], s.Messages[idx+1:]...)
	return true
}

func (s *session) deleteFirstNMessage(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if n < len(s.Messages) {
		for _, msg := range s.Messages[n:] {
			delete(s.msgIdx, msg.ID())
		}

		s.Messages = s.Messages[n:]
	} else {
		s.Messages = make([]message.Interface, 0, 16)
		s.msgIdx = make(map[string]int)
	}
}

func (s *session) generateContent() (msgOutCount int, _ []byte, _ error) {
	s.mu.RLock()
	msgOutCount = len(s.Messages)

	msgCopy := make([]message.Interface, msgOutCount)
	_ = copy(msgCopy, s.Messages)

	s.mu.RUnlock()

	// ensure every message is ready
	for _, m := range msgCopy {
		if !m.(Message).Ready() {
			// TODO: log output
			time.Sleep(1 * time.Second)
		}
	}

	result, err := s.generator.FormatPageBody(msgCopy)

	return msgOutCount, result, err
}

type (
	baseRequest struct {
		msgIDShouldReplyTo uint64
	}

	request interface {
		GetMessageIDShouldReplyTo() (uint64, bool)
		SetMessageIDShouldReplyTo(msgID uint64) bool
	}
)

func (b *baseRequest) GetMessageIDShouldReplyTo() (uint64, bool) {
	msgID := atomic.LoadUint64(&b.msgIDShouldReplyTo)
	return msgID, msgID != 0
}

func (b *baseRequest) SetMessageIDShouldReplyTo(msgID uint64) bool {
	return atomic.CompareAndSwapUint64(&b.msgIDShouldReplyTo, 0, msgID)
}

type (
	sessionRequest struct {
		baseRequest

		ChatID       uint64
		ChatUsername string

		// only one of topic and url shall be specified
		Topic string
		URL   string
	}

	editRequest struct {
		baseRequest
	}

	deleteRequest struct {
		baseRequest

		urls []string
	}

	listRequest struct {
		baseRequest
	}
)

func newSessionManager(ctx context.Context) *SessionManager {
	tq := queue.NewTimeoutQueue()
	tq.Start(ctx.Done())

	pendingRequests := &sync.Map{}
	ch := tq.TakeCh()
	go func() {
		for td := range ch {
			pendingRequests.Delete(td.Key.(uint64))
		}
	}()

	return &SessionManager{
		activeSessions: &sync.Map{},

		pending: pendingRequests,

		tq: tq,

		mu: &sync.Mutex{},
	}
}

type SessionManager struct {
	// chat_id -> session
	activeSessions *sync.Map

	// user_id -> request
	pending *sync.Map

	tq *queue.TimeoutQueue

	mu *sync.Mutex
}

func (c *SessionManager) scheduleDeleteTimeout(userID uint64, timeout time.Duration) {
	_ = c.tq.OfferWithDelay(userID, struct{}{}, timeout)
}

func (c *SessionManager) markPendingEditing(userID uint64, timeout time.Duration) (string, bool) {
	pVal, loaded := c.pending.LoadOrStore(
		userID, &editRequest{},
	)
	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}

	return getCommandFromRequest(pVal), !loaded
}

func (c *SessionManager) getPendingEditing(userID uint64) (*editRequest, bool) {
	pVal, ok := c.pending.Load(userID)
	if ok {
		ret, isEditReq := pVal.(*editRequest)
		return ret, isEditReq
	}
	return nil, false
}

func (c *SessionManager) markPendingListing(userID uint64, timeout time.Duration) (string, bool) {
	pVal, loaded := c.pending.LoadOrStore(
		userID, &listRequest{},
	)
	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}
	return getCommandFromRequest(pVal), !loaded
}

func (c *SessionManager) getPendingListing(userID uint64) (*listRequest, bool) {
	pVal, ok := c.pending.Load(userID)
	if ok {
		ret, isListReq := pVal.(*listRequest)
		return ret, isListReq
	}
	return nil, false
}

func (c *SessionManager) markPendingDeleting(userID uint64, urls []string, timeout time.Duration) (string, bool) {
	pVal, loaded := c.pending.LoadOrStore(
		userID, &deleteRequest{urls: urls},
	)
	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}
	return getCommandFromRequest(pVal), !loaded
}

func (c *SessionManager) getPendingDeleting(userID uint64) (*deleteRequest, bool) {
	pVal, ok := c.pending.Load(userID)
	if ok {
		ret, isDeleteReq := pVal.(*deleteRequest)
		return ret, isDeleteReq
	}
	return nil, false
}

func (c *SessionManager) resolvePendingRequest(userID uint64) (interface{}, bool) {
	c.tq.Remove(userID)

	pVal, loaded := c.pending.LoadAndDelete(userID)
	return pVal, loaded
}

func (c *SessionManager) markSessionStandby(
	userID, chatID uint64,
	chatUsername, topic, url string,
	timeout time.Duration,
) bool {
	_, loaded := c.pending.LoadOrStore(
		userID, &sessionRequest{
			ChatID:       chatID,
			ChatUsername: chatUsername,

			Topic: topic,
			URL:   url,
		},
	)
	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}
	return !loaded
}

func (c *SessionManager) getStandbySession(userID uint64) (*sessionRequest, bool) {
	pVal, ok := c.pending.Load(userID)
	if !ok {
		return nil, false
	}

	sr, ok := pVal.(*sessionRequest)
	if !ok {
		return nil, false
	}

	return sr, true
}

func (c *SessionManager) markRequestExpectingInput(userID, msgID uint64) bool {
	reqVal, ok := c.pending.Load(userID)
	if !ok {
		return false
	}

	return reqVal.(request).SetMessageIDShouldReplyTo(msgID)
}

func (c *SessionManager) getActiveSession(chatID uint64) (*session, bool) {
	sVal, ok := c.activeSessions.Load(chatID)
	if !ok {
		return nil, false
	}

	return sVal.(*session), true
}

func (c *SessionManager) activateSession(
	chatID, userID uint64,
	topic, defaultChatUsername string,
	gen generator.Interface,
) (*session, error) {

	c.mu.Lock()
	defer c.mu.Unlock()

	pVal, loaded := c.pending.LoadAndDelete(userID)
	if !loaded {
		return nil, fmt.Errorf("not found")
	}

	sr, ok := pVal.(*sessionRequest)
	if !ok {
		return nil, fmt.Errorf("conflict action, having other pending request")
	}

	if sr.ChatID != chatID {
		return nil, fmt.Errorf("chat not match")
	}

	newS := newSession(topic, defaultChatUsername, gen)
	sVal, loaded := c.activeSessions.LoadOrStore(chatID, newS)
	if !loaded {
		return newS, nil
	}

	return sVal.(*session), fmt.Errorf("already exists")
}

func (c *SessionManager) deactivateSession(chatID uint64) (_ *session, ok bool) {
	sVal, loaded := c.activeSessions.LoadAndDelete(chatID)
	if loaded {
		return sVal.(*session), true
	}

	return nil, false
}

func getCommandFromRequest(req interface{}) string {
	switch r := req.(type) {
	case *deleteRequest:
		return constant.CommandDelete
	case *listRequest:
		return constant.CommandList
	case *editRequest:
		return constant.CommandEdit
	case *sessionRequest:
		if len(r.Topic) != 0 {
			return constant.CommandDiscuss
		}

		return constant.CommandContinue
	}

	return ""
}
