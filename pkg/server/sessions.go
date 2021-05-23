package server

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
)

func newSession(topic, defaultChatUsername string, gen generator.Interface) *session {
	return &session{
		Topic:    topic,
		Messages: make([]Message, 0, 16),

		defaultChatUsername: defaultChatUsername,
		generator:           gen,
		msgIdx:              make(map[string]int),
		mu:                  &sync.RWMutex{},
	}
}

type session struct {
	Topic    string
	Messages []Message

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

	return s.Messages[size-1]
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
		s.Messages = make([]Message, 0, 16)
		s.msgIdx = make(map[string]int)
	}
}

func (s *session) generateContent(fm generator.Formatter) (msgOutCount int, _ []byte) {
	s.mu.RLock()
	msgCopy := make([]Message, 0, len(s.Messages))
	msgCopy = append(msgCopy, s.Messages...)
	s.mu.RUnlock()

	buf := &bytes.Buffer{}
	for _, msg := range msgCopy {
		chatUsername := s.defaultChatUsername

		// TODO: should we try to find original link?

		// switch {
		// case msg.SenderChat != nil && msg.SenderChat.Username != nil:
		// 	chatUsername = *msg.SenderChat.Username
		// case msg.ForwardFromChat != nil && msg.ForwardFromChat.Username != nil:
		// 	chatUsername = *msg.ForwardFromChat.Username
		// 	msgID = uint64(*msg.ForwardFromMessageId)
		// }

		for !msg.Ready() {
			// TODO: support cancel
			println("waiting for message ready")
			time.Sleep(1 * time.Second)
		}

		_, _ = buf.WriteString(
			fm.Format(
				generator.KindThematicBreak,
				fm.Format(
					generator.KindParagraph,
					fm.Format(
						generator.KindBlockquote,
						fm.Format(
							generator.KindURL,
							"(link)",
							fmt.Sprintf("https://t.me/%s/%s", chatUsername, msg.ID()),
						)+" "+string(msg.Format(fm)),
					),
				),
			),
		)
	}

	return len(msgCopy), buf.Bytes()
}

func newStandbySession(chatID uint64, chatUsername, topic, url string) *standbySession {
	return &standbySession{
		ChatID:       chatID,
		ChatUsername: chatUsername,

		Topic: topic,
		URL:   url,
	}
}

type standbySession struct {
	ChatID       uint64
	ChatUsername string

	// only one of topic and url shall be specified
	Topic string
	URL   string
}

func newSessionManager() *SessionManager {
	return &SessionManager{
		standbySessions: &sync.Map{},

		expectingInputSessions: &sync.Map{},

		activeSessions: &sync.Map{},

		pendingEditing: &sync.Map{},

		mu: &sync.Mutex{},
	}
}

type SessionManager struct {
	// user_id -> *standbySession
	standbySessions *sync.Map

	// user_id -> message_id requested input
	expectingInputSessions *sync.Map

	// chat_id -> session
	activeSessions *sync.Map

	// user_id -> message_id requested input
	pendingEditing *sync.Map

	mu *sync.Mutex
}

func (c *SessionManager) markPendingEditing(userID, messageID uint64) bool {
	_, loaded := c.pendingEditing.LoadOrStore(userID, messageID)
	return !loaded
}

func (c *SessionManager) getPendingEditing(userID uint64) (uint64, bool) {
	msgIDVal, ok := c.pendingEditing.Load(userID)
	if ok {
		return msgIDVal.(uint64), true
	}
	return 0, false
}

func (c *SessionManager) resolvePendingEditing(userID uint64) {
	c.pendingEditing.Delete(userID)
}

func (c *SessionManager) markSessionStandby(userID, chatID uint64, chatUsername, topic, url string) bool {
	_, loaded := c.standbySessions.LoadOrStore(
		userID, newStandbySession(chatID, chatUsername, topic, url),
	)
	return !loaded
}

func (c *SessionManager) getStandbySession(userID uint64) (*standbySession, bool) {
	sVal, ok := c.standbySessions.Load(userID)
	if !ok {
		return nil, false
	}

	return sVal.(*standbySession), true
}

func (c *SessionManager) cancelSessionStandby(userID uint64) bool {
	_, loaded := c.standbySessions.LoadAndDelete(userID)
	return loaded
}

func (c *SessionManager) markSessionExpectingInput(userID, messageID uint64) bool {
	_, loaded := c.expectingInputSessions.LoadOrStore(userID, messageID)
	return !loaded
}

func (c *SessionManager) sessionIsExpectingInput(userID uint64) (msgID uint64, ok bool) {
	msgIDVal, ok := c.expectingInputSessions.Load(userID)
	if !ok {
		return 0, false
	}

	return msgIDVal.(uint64), true
}

func (c *SessionManager) resolveSessionInput(userID uint64) {
	c.expectingInputSessions.Delete(userID)
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

	standbyVal, loaded := c.standbySessions.LoadAndDelete(userID)
	if !loaded {
		return nil, fmt.Errorf("not found")
	}

	if standbyVal.(*standbySession).ChatID != chatID {
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
