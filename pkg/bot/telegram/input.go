package telegram

import (
	"fmt"
	"strings"
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (c *tgBot) getChatByID(chatID int64) (cs chatSpec, err error) {
	var (
		resp  tg.MessagesChatsClass
		chats []tg.ChatClass
	)

	resp, err = c.client.API().MessagesGetChats(c.Context(), []int64{chatID})
	if err != nil {
		return
	}

	switch cts := resp.(type) {
	case *tg.MessagesChats:
		chats = cts.GetChats()
	case *tg.MessagesChatsSlice:
		chats = cts.GetChats()
	}

	return expectExactOneChat(chats)
}

func expectExactOneChat(chats []tg.ChatClass) (cs chatSpec, err error) {
	if len(chats) != 0 {
		err = fmt.Errorf("not single chat (%d)", len(chats))
		return
	}

	return resolveChatSpec(chats[0]), nil
}

type tokenInputHandleFunc = func(src *messageSource, msg *tg.Message, replyTo int) (bool, error)

func (c *tgBot) tryToHandleInputForDiscussOrContinue(
	src *messageSource, msg *tg.Message, replyTo int,
) (handled bool, err error) {
	chatID := uint64(src.Chat.ID())
	userID := uint64(src.From.GetPtr().ID())

	standbySession, hasStandbySession := c.GetStandbySession(userID)
	if !hasStandbySession {
		return false, nil
	}

	msgIDShouldReplyTo, isExpectingInput := standbySession.GetMessageIDShouldReplyTo()
	if !isExpectingInput ||
		uint64(replyTo) != msgIDShouldReplyTo {
		return false, nil
	}

	origPeer := src.Chat.InputPeer()
	if chatID != standbySession.ChatID {
		origChat, err := c.getChatByID(int64(standbySession.ChatID))
		if err != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold("failed to fetch original chat"),
			)
			return true, err
		}

		origPeer = origChat.InputPeer()
	}

	pub, userConfig, err := standbySession.Workflow().CreatePublisher()
	defer func() {
		if err != nil {
			_, _ = c.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("The session was canceled due to error, please retry later: "),
				styling.Bold(err.Error()),
			)

			if standbySession.ChatID != chatID {
				_, _ = c.sendTextMessage(
					c.sender.To(origPeer).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
					styling.Plain("The session was canceled due to error, please retry later"),
				)
			}
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return true, nil
	}

	userConfig.SetAuthToken(strings.TrimSpace(msg.GetMessage()))
	_, err = pub.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Bold(pub.Name()),
			styling.Plain(" auth error: "),
			styling.Bold(err.Error()),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	var (
		note []message.Span
	)
	switch {
	case len(standbySession.Topic) != 0:
		// is /discuss, create a new post
		content, err2 := standbySession.Workflow().Generator.RenderPageHeader()
		if err2 != nil {
			return true, fmt.Errorf("failed to generate initial page: %w", err2)
		}

		note, err2 = pub.Publish(standbySession.Topic, content)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Bold(pub.Name()),
				styling.Plain(" pre-publish failed: "),
				styling.Bold(err2.Error()),
			)
			return true, err2
		}
	case len(standbySession.URL) != 0:
		// is /continue, find existing post to edit

		note, err = pub.Retrieve(standbySession.URL)
		if err != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Retrieve "),
				styling.Bold(pub.Name()),
				styling.Plain(" post failed: "),
				styling.Bold(err.Error()),
			)

			// we may not find the post if user provided a wrong url, don't count this error
			// as session error
			return true, nil
		}
	}

	_, err = c.ActivateSession(standbySession.Workflow(), standbySession.ChatID, userID, pub)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return true, err
	}

	c.ResolvePendingRequest(userID)
	defer func() {
		if err != nil {
			// bset effort
			_, _ = c.DeactivateSession(standbySession.ChatID)
		}
	}()

	msgID, _ := c.sendTextMessage(
		c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
		styling.Plain("Success!"),
	)

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&src.Chat,
		10*time.Second,
		uint64(msgID),
		uint64(msg.GetID()),
		msgIDShouldReplyTo,
	)

	_, err = c.sendTextMessage(
		c.sender.To(origPeer).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
		translateEntities(note)...,
	)
	if err != nil {
		return true, nil
	}

	return true, nil
}

func (c *tgBot) tryToHandleInputForEditing(
	src *messageSource, msg *tg.Message, replyTo int,
) (bool, error) {
	userID := uint64(src.From.GetPtr().ID())

	// check if it's a reply for conversion started by links to /edit, /delete, /list
	req, isPendingEditing := c.GetPendingEditing(userID)
	if !isPendingEditing {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isExpectingInput ||
		uint64(replyTo) != msgIDshouldReplyTo {
		return false, nil
	}

	pub, userConfig, err := req.Workflow().CreatePublisher()
	defer func() {
		if err != nil {
			c.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("The edit was canceled due to error, please retry later: "),
				styling.Bold(err.Error()),
			)
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return true, nil
	}

	userConfig.SetAuthToken(strings.TrimSpace(msg.GetMessage()))
	_, err = pub.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Bold(pub.Name()),
			styling.Plain(" auth error: "),
			styling.Bold(err.Error()),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	authURL, err := pub.AuthURL()
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Bold(pub.Name()),
			styling.Plain(" unable to get auth url: "),
			styling.Bold(err.Error()),
		)
		return true, err
	}

	c.ResolvePendingRequest(userID)

	msgID, _ := c.sendTextMessage(
		c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().
			Silent().Reply(msg.GetID()).Row(&tg.KeyboardButtonURL{
			Text: "Edit on this device",
			URL:  authURL,
		}),
		styling.Plain("Login Success!"),
	)

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&src.Chat, 10*time.Second,
		uint64(msg.GetID()),
		msgIDshouldReplyTo,
	)

	// delete auth url later
	c.scheduleMessageDelete(&src.Chat, 5*time.Minute, uint64(msgID))

	return true, nil
}

func (c *tgBot) tryToHandleInputForListing(
	src *messageSource, msg *tg.Message, replyTo int,
) (bool, error) {
	userID := uint64(src.From.GetPtr().ID())

	// check if it's a reply for conversion started by links to /list
	req, isPendingListing := c.GetPendingListing(userID)
	if !isPendingListing {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isExpectingInput ||
		uint64(replyTo) != msgIDshouldReplyTo {
		return false, nil
	}

	pub, userConfig, err := req.Workflow().CreatePublisher()
	defer func() {
		if err != nil {
			c.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("The list request was canceled due to error, please retry later: "),
				styling.Bold(err.Error()),
			)
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return true, nil
	}

	userConfig.SetAuthToken(strings.TrimSpace(msg.GetMessage()))
	_, err = pub.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Bold(pub.Name()),
			styling.Plain(" auth error: "),
			styling.Bold(err.Error()),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	posts, err := pub.List()
	if err != nil && len(posts) == 0 {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Bold(pub.Name()),
			styling.Plain(" unable to list posts: "),
			styling.Bold(err.Error()),
		)
		return true, err
	}

	c.ResolvePendingRequest(userID)

	entities := make([]styling.StyledTextOption, 0, len(posts)*3)
	for _, p := range posts {
		entities = append(entities,
			styling.Plain("- "),
			styling.TextURL(p.Title, p.URL),
			styling.Plain("\n"),
		)
	}

	_, _ = c.sendTextMessage(
		c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
		entities...,
	)

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&src.Chat,
		10*time.Second,
		uint64(msg.GetID()),
		msgIDshouldReplyTo,
	)

	return true, nil
}

func (c *tgBot) tryToHandleInputForDeleting(
	src *messageSource, msg *tg.Message, replyTo int,
) (bool, error) {
	userID := uint64(src.From.GetPtr().ID())

	// check if it's a reply for conversion started by links to /delete
	req, isPendingDeleting := c.GetPendingDeleting(userID)
	if !isPendingDeleting {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isPendingDeleting || !isExpectingInput ||
		uint64(replyTo) != msgIDshouldReplyTo {
		return false, nil
	}

	pub, userConfig, err := req.Workflow().CreatePublisher()
	defer func() {
		if err != nil {
			c.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("The delete request was canceled due to error, please retry later: "),
				styling.Bold(err.Error()),
			)
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return true, nil
	}

	userConfig.SetAuthToken(strings.TrimSpace(msg.GetMessage()))
	_, err = pub.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Bold(pub.Name()),
			styling.Plain(" auth error: "),
			styling.Bold(err.Error()),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	err = pub.Delete(req.URLs...)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Bold(pub.Name()),
			styling.Plain(": "),
			styling.Plain(err.Error()),
		)
		return true, err
	}

	c.ResolvePendingRequest(userID)

	_, _ = c.sendTextMessage(
		c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent(),
		styling.Plain("The post(s) has been deleted"),
	)

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&src.Chat,
		10*time.Second,
		uint64(msg.GetID()),
		msgIDshouldReplyTo,
	)

	return true, nil
}
