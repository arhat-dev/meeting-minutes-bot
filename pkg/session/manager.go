package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"arhat.dev/pkg/queue"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

func NewManager[C Chat](ctx context.Context) Manager[C] {
	tq := queue.NewTimeoutQueue[rt.UserID, struct{}]()
	tq.Start(ctx.Done())

	pendingRequests := &sync.Map{}
	ch := tq.TakeCh()
	go func() {
		for td := range ch {
			pendingRequests.Delete(td.Key)
		}
	}()

	return Manager[C]{
		pendingRequests: pendingRequests,
		activeSessions:  &sync.Map{},

		tq: tq,

		mu: &sync.Mutex{},
	}
}

type Chat interface {
	ID() rt.ChatID
}

type Manager[C Chat] struct {
	// key: user_id
	// value: request
	pendingRequests *sync.Map

	// key: chat_id
	// value: Session
	activeSessions *sync.Map

	tq *queue.TimeoutQueue[rt.UserID, struct{}]

	mu *sync.Mutex
}

func (c *Manager[C]) scheduleDeleteTimeout(userID rt.UserID, timeout time.Duration) {
	_ = c.tq.OfferWithDelay(userID, struct{}{}, timeout)
}

func (c *Manager[C]) MarkPendingEditing(
	wf *bot.Workflow,
	userID rt.UserID,
	timeout time.Duration,
) (bot.BotCmd, bool) {
	pVal, loaded := c.pendingRequests.LoadOrStore(
		userID, &EditRequest{BaseRequest: BaseRequest{wf: wf}},
	)
	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}

	return GetCommandFromRequest[C](pVal), !loaded
}

func (c *Manager[C]) GetPendingEditing(userID rt.UserID) (*EditRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if ok {
		ret, isEditReq := pVal.(*EditRequest)
		return ret, isEditReq
	}
	return nil, false
}

func (c *Manager[C]) MarkPendingListing(
	wf *bot.Workflow,
	userID rt.UserID,
	timeout time.Duration,
) (bot.BotCmd, bool) {
	pVal, loaded := c.pendingRequests.LoadOrStore(
		userID, &ListRequest{BaseRequest: BaseRequest{wf: wf}},
	)

	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}
	return GetCommandFromRequest[C](pVal), !loaded
}

func (c *Manager[C]) GetPendingListing(userID rt.UserID) (*ListRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if ok {
		ret, isListReq := pVal.(*ListRequest)
		return ret, isListReq
	}
	return nil, false
}

// MarkPendingDeleting marks current session related to userID as DeleteRequest
func (c *Manager[C]) MarkPendingDeleting(
	wf *bot.Workflow,
	userID rt.UserID,
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
	return GetCommandFromRequest[C](pVal), !loaded
}

func (c *Manager[C]) GetPendingDeleting(userID rt.UserID) (*DeleteRequest, bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if ok {
		ret, isDeleteReq := pVal.(*DeleteRequest)
		return ret, isDeleteReq
	}
	return nil, false
}

func (c *Manager[C]) ResolvePendingRequest(userID rt.UserID) (any, bool) {
	c.tq.Remove(userID)

	pVal, loaded := c.pendingRequests.LoadAndDelete(userID)
	return pVal, loaded
}

// MarkSessionStandby prepare a new session
func (c *Manager[C]) MarkSessionStandby(
	wf *bot.Workflow,
	userID rt.UserID,
	chatID C,
	topic, url string,
	timeout time.Duration,
) bool {
	_, loaded := c.pendingRequests.LoadOrStore(
		userID, &SessionRequest[C]{
			BaseRequest: BaseRequest{wf: wf},

			Data: chatID,

			Topic: topic,
			URL:   url,
		},
	)
	if !loaded {
		c.scheduleDeleteTimeout(userID, timeout)
	}
	return !loaded
}

func (c *Manager[C]) GetStandbySession(userID rt.UserID) (*SessionRequest[C], bool) {
	pVal, ok := c.pendingRequests.Load(userID)
	if !ok {
		return nil, false
	}

	sr, ok := pVal.(*SessionRequest[C])
	if !ok {
		return nil, false
	}

	return sr, true
}

func (c *Manager[C]) MarkRequestExpectingInput(
	userID rt.UserID, shouldReplyToMsgID rt.MessageID,
) bool {
	reqVal, ok := c.pendingRequests.Load(userID)
	if !ok {
		return false
	}

	return reqVal.(Request).SetMessageIDShouldReplyTo(shouldReplyToMsgID)
}

func (c *Manager[C]) GetActiveSession(chatID rt.ChatID) (ret *Session, ok bool) {
	sVal, ok := c.activeSessions.Load(chatID)
	if !ok {
		return
	}

	ret = sVal.(*Session)
	return
}

func (c *Manager[C]) ActivateSession(
	wf *bot.Workflow, userID rt.UserID, chatID rt.ChatID, p publisher.Interface,
) (_ *Session, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	pVal, loaded := c.pendingRequests.LoadAndDelete(userID)
	if !loaded {
		return nil, fmt.Errorf("session request not found")
	}

	sr, ok := pVal.(*SessionRequest[C])
	if !ok {
		return nil, fmt.Errorf("conflict action, other pending request exists")
	}

	if sr.Data.ID() != chatID {
		return nil, fmt.Errorf("chat not match")
	}

	newS := newSession(wf, p)
	sVal, loaded := c.activeSessions.LoadOrStore(chatID, newS)
	if loaded {
		return sVal.(*Session), fmt.Errorf("already exists")
	}

	return newS, nil
}

func (c *Manager[C]) DeactivateSession(chatID rt.ChatID) (_ *Session, ok bool) {
	sVal, loaded := c.activeSessions.LoadAndDelete(chatID)
	if loaded {
		return sVal.(*Session), true
	}

	return nil, false
}
