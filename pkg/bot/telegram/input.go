package telegram

import (
	"fmt"
	"strings"
	"time"

	"arhat.dev/mbot/pkg/rt"
	"github.com/gotd/td/telegram/message/styling"
)

// input handling func MUST be ONLY called in private chat

type inputHandleFunc = func(mc *messageContext, replyTo rt.MessageID) (bool, error)

func (c *tgBot) tryToHandleInputForDiscussOrContinue(mc *messageContext, replyTo rt.MessageID) (handled bool, err error) {
	chatID := mc.src.Chat.ID()
	userID := mc.src.From.ID()

	mc.logger.V("try handle input for discuss or continue")
	standbySession, hasStandbySession := c.sessions.GetStandbySession(userID)
	if !hasStandbySession {
		return false, nil
	}

	msgIDShouldReplyTo, isExpectingInput := standbySession.GetMessageIDShouldReplyTo()
	if !isExpectingInput || replyTo != msgIDShouldReplyTo {
		return false, nil
	}

	handled = true
	origPeer := mc.src.Chat.InputPeer()
	if chatID != standbySession.Data.ID() {
		origPeer = standbySession.Data.chat
	}

	pub, userConfig, err := standbySession.Workflow().CreatePublisher()
	defer func() {
		if err != nil {
			_, _ = c.sessions.ResolvePendingRequest(userID)

			// best effort

			_, _ = c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Plain("The session was canceled due to error, please retry later: "),
				styling.Bold(err.Error()),
			)

			if standbySession.Data.ID() != chatID {
				_, _ = c.sendTextMessage(
					c.sender.To(origPeer).NoWebpage().Silent().Reply(mc.msg.GetID()),
					styling.Plain("The session was canceled due to error, please retry later"),
				)
			}
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return true, nil
	}

	userConfig.SetToken(strings.TrimSpace(mc.msg.GetMessage()))
	_, err = pub.Login(&mc.con, userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Bold(standbySession.Workflow().PublisherName()),
			styling.Plain(" auth error: "),
			styling.Bold(err.Error()),
		)
		// usually not our fault, let user try again
		err = nil
		return
	}

	var (
		note rt.PublisherOutput
	)

	if standbySession.IsDiscuss {
		// is /discuss, create a new post

		content, err2 := standbySession.Workflow().Generator.New(
			&mc.con,
			standbySession.Workflow().BotCommands.TextOf(rt.BotCmd_Discuss),
			standbySession.Params,
		)
		if err2 != nil {
			return true, fmt.Errorf("failed to generate initial page: %w", err2)
		}

		note, err2 = pub.CreateNew(
			&mc.con,
			standbySession.Workflow().BotCommands.TextOf(rt.BotCmd_Discuss),
			standbySession.Params,
			&content,
		)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Bold(standbySession.Workflow().PublisherName()),
				styling.Plain(" pre-publish failed: "),
				styling.Bold(err2.Error()),
			)
			return true, err2
		}
	} else {
		// is /continue, find existing post to edit

		note, err = pub.Retrieve(
			&mc.con,
			standbySession.Workflow().BotCommands.TextOf(rt.BotCmd_Continue),
			standbySession.Params,
		)
		if err != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Plain("Retrieve "),
				styling.Bold(standbySession.Workflow().PublisherName()),
				styling.Plain(" post failed: "),
				styling.Bold(err.Error()),
			)

			// we may not find the post if user provided a wrong url, don't count this error
			// as session error
			return true, nil
		}
	}

	mc.logger.V("activate session")
	_, err = c.sessions.ActivateSession(standbySession.Workflow(), userID, standbySession.Data.ID(), pub)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return true, err
	}

	c.sessions.ResolvePendingRequest(userID)
	defer func(chatID rt.ChatID) {
		if err != nil {
			// bset effort
			_, _ = c.sessions.DeactivateSession(chatID)
		}
	}(standbySession.Data.ID())

	msgID, _ := c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
		styling.Plain("Success!"),
	)

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&mc.src.Chat,
		10*time.Second,
		msgID,
		rt.MessageID(mc.msg.GetID()),
		msgIDShouldReplyTo,
	)

	switch {
	case !note.SendMessage.IsNil():
		con := mc.con
		con.peer = origPeer
		_, err = con.SendMessage(c.Context(), note.SendMessage.Get())
	}
	if err != nil {
		return true, nil
	}

	return true, nil
}

func (c *tgBot) tryToHandleInputForEditing(mc *messageContext, replyTo rt.MessageID) (bool, error) {
	userID := mc.src.From.ID()

	mc.logger.V("try handle input for editing")
	// check if it's a reply for conversion started by links to /edit, /delete, /list
	req, isPendingEditing := c.sessions.GetPendingEditing(userID)
	if !isPendingEditing {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isExpectingInput || replyTo != msgIDshouldReplyTo {
		return false, nil
	}

	pub, userConfig, err := req.Workflow().CreatePublisher()
	defer func() {
		if err != nil {
			c.sessions.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Plain("The edit was canceled due to error, please retry later: "),
				styling.Bold(err.Error()),
			)
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return true, nil
	}

	userConfig.SetToken(strings.TrimSpace(mc.msg.GetMessage()))
	_, err = pub.Login(&mc.con, userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Bold(req.Workflow().PublisherName()),
			styling.Plain(" auth error: "),
			styling.Bold(err.Error()),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	authURL, err := pub.RequestExternalAccess(&mc.con)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Bold(req.Workflow().PublisherName()),
			styling.Plain(" unable to get auth url: "),
			styling.Bold(err.Error()),
		)
		return true, err
	}

	c.sessions.ResolvePendingRequest(userID)

	var msgIDs []rt.MessageID
	switch {
	case !authURL.SendMessage.IsNil():
		msgIDs, _ = mc.con.SendMessage(c.Context(), authURL.SendMessage.Get())
	}

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&mc.src.Chat, 10*time.Second,
		rt.MessageID(mc.msg.GetID()),
		msgIDshouldReplyTo,
	)

	// delete auth url later
	c.scheduleMessageDelete(&mc.src.Chat, 5*time.Minute, msgIDs...)

	return true, nil
}

func (c *tgBot) tryToHandleInputForListing(mc *messageContext, replyTo rt.MessageID) (handled bool, err error) {
	userID := mc.src.From.ID()

	mc.logger.V("try handle input for listing")
	// check if it's a reply for conversion started by links to /list
	req, isPendingListing := c.sessions.GetPendingListing(userID)
	if !isPendingListing {
		return
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isExpectingInput || replyTo != msgIDshouldReplyTo {
		return
	}

	handled = true
	defer func() {
		if err != nil {
			c.sessions.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Plain("The list request was canceled due to error, please retry later: "),
				styling.Bold(err.Error()),
			)
		}
	}()

	pub, userConfig, err := req.Workflow().CreatePublisher()
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return
	}

	userConfig.SetToken(strings.TrimSpace(mc.msg.GetMessage()))
	_, err = pub.Login(&mc.con, userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Bold(req.Workflow().PublisherName()),
			styling.Plain(" auth error: "),
			styling.Bold(err.Error()),
		)
		// usually not our fault, let user try again
		err = nil
		return
	}

	posts, err := pub.List(&mc.con)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return
	}

	c.sessions.ResolvePendingRequest(userID)

	switch {
	case !posts.SendMessage.IsNil():
		_, _ = mc.con.SendMessage(c.Context(), posts.SendMessage.Get())
	}

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&mc.src.Chat,
		10*time.Second,
		rt.MessageID(mc.msg.GetID()),
		msgIDshouldReplyTo,
	)

	return
}

// tryToHandleInputForDeleting treate msg as user input for deleting published post
func (c *tgBot) tryToHandleInputForDeleting(mc *messageContext, replyTo rt.MessageID) (handled bool, err error) {
	userID := mc.src.From.ID()

	mc.logger.V("try handle input for deleting")
	// check if it's a reply for conversion started by links to /delete
	req, isPendingDeleting := c.sessions.GetPendingDeleting(userID)
	if !isPendingDeleting {
		return
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isPendingDeleting || !isExpectingInput || replyTo != msgIDshouldReplyTo {
		return
	}

	handled = true
	defer func() {
		if err != nil {
			c.sessions.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Plain("The delete request was canceled due to error, please retry later: "),
				styling.Bold(err.Error()),
			)
		}
	}()

	pub, userConfig, err := req.Workflow().CreatePublisher()
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return
	}

	userConfig.SetToken(strings.TrimSpace(mc.msg.GetMessage()))
	_, err = pub.Login(&mc.con, userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Bold(req.Workflow().PublisherName()),
			styling.Plain(" auth error: "),
			styling.Bold(err.Error()),
		)
		// usually not our fault, let user try again
		err = nil
		return
	}

	note, err := pub.Delete(
		&mc.con,
		req.Workflow().BotCommands.TextOf(rt.BotCmd_Delete),
		req.Params,
	)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Plain(err.Error()),
		)
		return
	}

	c.sessions.ResolvePendingRequest(userID)

	switch {
	case !note.SendMessage.IsNil():
		_, _ = mc.con.SendMessage(c.Context(), note.SendMessage.Get())
	}

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&mc.src.Chat,
		10*time.Second,
		rt.MessageID(mc.msg.GetID()),
		rt.MessageID(msgIDshouldReplyTo),
	)

	return
}
