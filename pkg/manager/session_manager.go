package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"arhat.dev/pkg/queue"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

func NewSessionManager(ctx context.Context) SessionManager {
	tq := queue.NewTimeoutQueue[uint64, struct{}]()
	tq.Start(ctx.Done())

	pendingRequests := &sync.Map{}
	ch := tq.TakeCh()
	go func() {
		for td := range ch {
			pendingRequests.Delete(td.Key)
		}
	}()

	return SessionManager{
		pendingRequests: pendingRequests,
		activeSessions:  &sync.Map{},

		tq: tq,

		mu: &sync.Mutex{},
	}
}

type SessionManager struct {
	// key: user_id
	// value: request
	pendingRequests *sync.Map

	// key: chat_id
	// value: Session
	activeSessions *sync.Map

	tq *queue.TimeoutQueue[uint64, struct{}]

	mu *sync.Mutex
}

func (c *SessionManager) scheduleDeleteTimeout(userID uint64, timeout time.Duration) {
	_ = c.tq.OfferWithDelay(userID, struct{}{}, timeout)
}

func (c *SessionManager) MarkPendingEditing(
	wf *bot.Workflow,
	userID uint64,
	timeout time.Duration,
) (bot.BotCmd, bool) {
	pVal, loaded := c.pendingRequests.LoadOrStore(
		userID, &EditRequest{BaseRequest: BaseRequest{wf: wf}},
	)
	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}

	return GetCommandFromRequest(pVal), !loaded
}

func (c *SessionManager) GetPendingEditing(userID uint64) (*EditRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if ok {
		ret, isEditReq := pVal.(*EditRequest)
		return ret, isEditReq
	}
	return nil, false
}

func (c *SessionManager) MarkPendingListing(
	wf *bot.Workflow,
	userID uint64,
	timeout time.Duration,
) (bot.BotCmd, bool) {
	pVal, loaded := c.pendingRequests.LoadOrStore(
		userID, &ListRequest{BaseRequest: BaseRequest{wf: wf}},
	)

	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}
	return GetCommandFromRequest(pVal), !loaded
}

func (c *SessionManager) GetPendingListing(userID uint64) (*ListRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if ok {
		ret, isListReq := pVal.(*ListRequest)
		return ret, isListReq
	}
	return nil, false
}

// MarkPendingDeleting marks current session related to userID as DeleteRequest
func (c *SessionManager) MarkPendingDeleting(
	wf *bot.Workflow,
	userID uint64,
	urls []string,
	timeout time.Duration,
) (prevCmd bot.BotCmd, hasPrevCmd bool) {
	pVal, loaded := c.pendingRequests.LoadOrStore(
		userID, &DeleteRequest{
			BaseRequest: BaseRequest{
				wf: wf,
			},
			URLs: urls,
		},
	)
	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}
	return GetCommandFromRequest(pVal), !loaded
}

func (c *SessionManager) GetPendingDeleting(userID uint64) (*DeleteRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if ok {
		ret, isDeleteReq := pVal.(*DeleteRequest)
		return ret, isDeleteReq
	}
	return nil, false
}

func (c *SessionManager) ResolvePendingRequest(userID uint64) (interface{}, bool) {
	c.tq.Remove(userID)

	pVal, loaded := c.pendingRequests.LoadAndDelete(userID)
	return pVal, loaded
}

func (c *SessionManager) MarkSessionStandby(
	wf *bot.Workflow,
	userID, chatID uint64,
	topic, url string,
	timeout time.Duration,
) bool {
	_, loaded := c.pendingRequests.LoadOrStore(
		userID, &SessionRequest{
			BaseRequest: BaseRequest{wf: wf},

			ChatID: chatID,

			Topic: topic,
			URL:   url,
		},
	)
	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}
	return !loaded
}

func (c *SessionManager) GetStandbySession(userID uint64) (*SessionRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if !ok {
		return nil, false
	}

	sr, ok := pVal.(*SessionRequest)
	if !ok {
		return nil, false
	}

	return sr, true
}

func (c *SessionManager) MarkRequestExpectingInput(userID, shouldReplyToMsgID uint64) bool {
	reqVal, ok := c.pendingRequests.Load(userID)
	if !ok {
		return false
	}

	return reqVal.(Request).SetMessageIDShouldReplyTo(shouldReplyToMsgID)
}

func (c *SessionManager) GetActiveSession(chatID uint64) (*Session, bool) {
	sVal, ok := c.activeSessions.Load(chatID)
	if !ok {
		return nil, false
	}

	return sVal.(*Session), true
}

func (c *SessionManager) ActivateSession(
	wf *bot.Workflow, chatID, userID uint64, p publisher.Interface,
) (_ *Session, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pVal, loaded := c.pendingRequests.LoadAndDelete(userID)
	if !loaded {
		return nil, fmt.Errorf("session request not found")
	}

	sr, ok := pVal.(*SessionRequest)
	if !ok {
		return nil, fmt.Errorf("conflict action, other pending request exists")
	}

	if sr.ChatID != chatID {
		return nil, fmt.Errorf("chat not match")
	}

	newS := newSession(wf, p)
	sVal, loaded := c.activeSessions.LoadOrStore(chatID, newS)
	if loaded {
		return sVal.(*Session), fmt.Errorf("already exists")
	}

	return newS, nil
}

func (c *SessionManager) DeactivateSession(chatID uint64) (_ *Session, ok bool) {
	sVal, loaded := c.activeSessions.LoadAndDelete(chatID)
	if loaded {
		return sVal.(*Session), true
	}

	return nil, false
}
