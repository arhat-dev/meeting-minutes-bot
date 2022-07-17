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

	userConfig.SetAuthToken(strings.TrimSpace(mc.msg.GetMessage()))
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
		note []rt.Span
	)

	if standbySession.IsDiscuss {
		// is /discuss, create a new post

		content, err2 := standbySession.Workflow().Generator.New(
			&mc.con,
			standbySession.Workflow().BotCommands.TextOf(rt.BotCmd_Discuss),
			"",
		)
		if err2 != nil {
			return true, fmt.Errorf("failed to generate initial page: %w", err2)
		}

		var (
			rd strings.Reader
			in rt.Input
		)
		rd.Reset(content)
		in = rt.NewInput(rd.Size(), &rd)

		note, err2 = pub.CreateNew(
			&mc.con,
			standbySession.Workflow().BotCommands.TextOf(rt.BotCmd_Discuss),
			standbySession.Params,
			&in,
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

	_, err = c.sendTextMessage(
		c.sender.To(origPeer).NoWebpage().Silent().Reply(mc.msg.GetID()),
		translateTextSpans(note)...,
	)
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

	userConfig.SetAuthToken(strings.TrimSpace(mc.msg.GetMessage()))
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

	msgID, _ := c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
		translateTextSpans(authURL)...,
	)

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&mc.src.Chat, 10*time.Second,
		rt.MessageID(mc.msg.GetID()),
		msgIDshouldReplyTo,
	)

	// delete auth url later
	c.scheduleMessageDelete(&mc.src.Chat, 5*time.Minute, msgID)

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

	userConfig.SetAuthToken(strings.TrimSpace(mc.msg.GetMessage()))
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
	if err != nil && len(posts) == 0 {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Bold(req.Workflow().PublisherName()),
			styling.Plain(" unable to list posts: "),
			styling.Bold(err.Error()),
		)
		return
	}

	c.sessions.ResolvePendingRequest(userID)

	entities := make([]styling.StyledTextOption, 0, len(posts)*3)
	for _, p := range posts {
		entities = append(entities,
			styling.Plain("- "),
			styling.TextURL(p.Title, p.URL),
			styling.Plain("\n"),
		)
	}

	_, _ = c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent(),
		entities...,
	)

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

	userConfig.SetAuthToken(strings.TrimSpace(mc.msg.GetMessage()))
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

	err = pub.Delete(
		&mc.con,
		req.Workflow().BotCommands.TextOf(rt.BotCmd_Delete),
		req.Params,
	)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Bold(req.Workflow().PublisherName()),
			styling.Plain(": "),
			styling.Plain(err.Error()),
		)
		return
	}

	c.sessions.ResolvePendingRequest(userID)

	_, _ = c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent(),
		styling.Plain("The post(s) has been deleted"),
	)

	// delete user provided token related messages
	c.scheduleMessageDelete(
		&mc.src.Chat,
		10*time.Second,
		rt.MessageID(mc.msg.GetID()),
		rt.MessageID(msgIDshouldReplyTo),
	)

	return
}
