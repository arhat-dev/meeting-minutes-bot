package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"arhat.dev/pkg/queue"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

func NewSessionManager[M message.Interface](ctx context.Context) SessionManager[M] {
	tq := queue.NewTimeoutQueue[uint64, struct{}]()
	tq.Start(ctx.Done())

	pendingRequests := &sync.Map{}
	ch := tq.TakeCh()
	go func() {
		for td := range ch {
			pendingRequests.Delete(td.Key)
		}
	}()

	return SessionManager[M]{
		pendingRequests: pendingRequests,
		activeSessions:  &sync.Map{},

		tq: tq,

		mu: &sync.Mutex{},
	}
}

type SessionManager[M message.Interface] struct {
	// key: user_id
	// value: request
	pendingRequests *sync.Map

	// key: chat_id
	// value: Session
	activeSessions *sync.Map

	tq *queue.TimeoutQueue[uint64, struct{}]

	mu *sync.Mutex
}

func (c *SessionManager[M]) scheduleDeleteTimeout(userID uint64, timeout time.Duration) {
	_ = c.tq.OfferWithDelay(userID, struct{}{}, timeout)
}

func (c *SessionManager[M]) MarkPendingEditing(
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

func (c *SessionManager[M]) GetPendingEditing(userID uint64) (*EditRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if ok {
		ret, isEditReq := pVal.(*EditRequest)
		return ret, isEditReq
	}
	return nil, false
}

func (c *SessionManager[M]) MarkPendingListing(
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

func (c *SessionManager[M]) GetPendingListing(userID uint64) (*ListRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if ok {
		ret, isListReq := pVal.(*ListRequest)
		return ret, isListReq
	}
	return nil, false
}

// MarkPendingDeleting marks current session related to userID as DeleteRequest
func (c *SessionManager[M]) MarkPendingDeleting(
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

func (c *SessionManager[M]) GetPendingDeleting(userID uint64) (*DeleteRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if ok {
		ret, isDeleteReq := pVal.(*DeleteRequest)
		return ret, isDeleteReq
	}
	return nil, false
}

func (c *SessionManager[M]) ResolvePendingRequest(userID uint64) (interface{}, bool) {
	c.tq.Remove(userID)

	pVal, loaded := c.pendingRequests.LoadAndDelete(userID)
	return pVal, loaded
}

// MarkSessionStandby prepare a new session
func (c *SessionManager[M]) MarkSessionStandby(
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

func (c *SessionManager[M]) GetStandbySession(userID uint64) (*SessionRequest, bool) {
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

func (c *SessionManager[M]) MarkRequestExpectingInput(userID, shouldReplyToMsgID uint64) bool {
	reqVal, ok := c.pendingRequests.Load(userID)
	if !ok {
		return false
	}

	return reqVal.(Request).SetMessageIDShouldReplyTo(shouldReplyToMsgID)
}

func (c *SessionManager[M]) GetActiveSession(chatID uint64) (ret *Session[M], ok bool) {
	sVal, ok := c.activeSessions.Load(chatID)
	if !ok {
		return
	}

	ret = sVal.(*Session[M])
	return
}

func (c *SessionManager[M]) ActivateSession(
	wf *bot.Workflow, chatID, userID uint64, p publisher.Interface,
) (_ *Session[M], err error) {
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

	newS := newSession[M](wf, p)
	sVal, loaded := c.activeSessions.LoadOrStore(chatID, newS)
	if loaded {
		return sVal.(*Session[M]), fmt.Errorf("already exists")
	}

	return newS, nil
}

func (c *SessionManager[M]) DeactivateSession(chatID uint64) (_ *Session[M], ok bool) {
	sVal, loaded := c.activeSessions.LoadAndDelete(chatID)
	if loaded {
		return sVal.(*Session[M]), true
	}

	return nil, false
}
