package telegram

import (
	"fmt"
	"strings"
	"time"

	api "arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

type tokenInputHandleFunc func(chatID uint64, userID uint64, msg *api.Message) (bool, error)

func (c *telegramBot) tryToHandleInputForDiscussOrContinue(
	chatID uint64,
	userID uint64,
	msg *api.Message,
) (bool, error) {
	standbySession, hasStandbySession := c.GetStandbySession(userID)
	if !hasStandbySession {
		return false, nil
	}

	msgIDShouldReplyTo, isExpectingInput := standbySession.GetMessageIDShouldReplyTo()
	if !isExpectingInput ||
		uint64(msg.ReplyToMessage.MessageId) != msgIDShouldReplyTo {
		return false, nil
	}

	pub, userConfig, err := c.createPublisher()
	defer func() {
		if err != nil {
			_, _ = c.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				chatID, true, true, 0,
				fmt.Sprintf("The session was canceled due to error, please retry later: %v", err),
			)

			if standbySession.ChatID != chatID {
				_, _ = c.sendTextMessage(
					standbySession.ChatID, true, true, 0,
					"The session was canceled due to error, please retry later",
				)
			}
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Internal bot error: %v", err),
		)
		return true, nil
	}

	userConfig.SetAuthToken(strings.TrimSpace(*msg.Text))
	_, err = pub.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: auth error: %v", pub.Name(), err),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	var (
		note []message.Entity
	)
	switch {
	case len(standbySession.Topic) != 0:
		// is /discuss, create a new post
		content, err2 := c.generator.RenderPageHeader()
		if err2 != nil {
			return true, fmt.Errorf("failed to generate initial page: %w", err2)
		}

		note, err2 = pub.Publish(standbySession.Topic, content)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, 0,
				fmt.Sprintf("%s pre-publish failed: %v", pub.Name(), err2),
			)
			return true, err2
		}
	case len(standbySession.URL) != 0:
		// is /continue, find existing post to edit

		note, err = pub.Retrieve(standbySession.URL)
		if err != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Retrieve %s post error: %v", pub.Name(), err),
			)

			// we may not find the post if user provided a wrong url, don't count this error
			// as session error
			return true, nil
		}
	}

	_, err = c.ActivateSession(standbySession.ChatID, userID, pub)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Internal bot error: %v", err),
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

	msgID, _ := c.sendTextMessage(chatID, true, true, 0, "Success!")

	// delete user provided token related messages
	c.scheduleMessageDelete(
		chatID, 10*time.Second,
		uint64(msgID),
		uint64(msg.MessageId),
		msgIDShouldReplyTo,
	)

	_, err = c.sendTextMessage(standbySession.ChatID, true, true, 0, c.renderEntities(note))
	if err != nil {
		return true, nil
	}

	return true, nil
}

func (c *telegramBot) tryToHandleInputForEditing(
	chatID uint64,
	userID uint64,
	msg *api.Message,
) (bool, error) {
	// check if it's a reply for conversion started by links to /edit, /delete, /list
	req, isPendingEditing := c.GetPendingEditing(userID)
	if !isPendingEditing {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isExpectingInput ||
		uint64(msg.ReplyToMessage.MessageId) != msgIDshouldReplyTo {
		return false, nil
	}

	pub, userConfig, err := c.createPublisher()
	defer func() {
		if err != nil {
			c.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				chatID, true, true, 0,
				fmt.Sprintf("The edit was canceled due to error, please retry later: %v", err),
			)
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Internal bot error: %v", err),
		)
		return true, nil
	}

	userConfig.SetAuthToken(strings.TrimSpace(*msg.Text))
	_, err = pub.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: auth error: %v", pub.Name(), err),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	authURL, err := pub.AuthURL()
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: unable to get auth url: %v", pub.Name(), err),
		)
		return true, err
	}

	c.ResolvePendingRequest(userID)

	msgID, _ := c.sendTextMessage(
		chatID, true, true, 0, "Login Success!",
		api.InlineKeyboardMarkup{
			InlineKeyboard: [][]api.InlineKeyboardButton{{{
				Text: "Edit on this device",
				Url:  &authURL,
			}}},
		},
	)

	// delete user provided token related messages
	c.scheduleMessageDelete(
		chatID, 10*time.Second,
		uint64(msg.MessageId),
		msgIDshouldReplyTo,
	)

	// delete auth url later
	c.scheduleMessageDelete(chatID, 5*time.Minute, uint64(msgID))

	return true, nil
}

func (c *telegramBot) tryToHandleInputForListing(
	chatID uint64,
	userID uint64,
	msg *api.Message,
) (bool, error) {
	// check if it's a reply for conversion started by links to /list
	req, isPendingListing := c.GetPendingListing(userID)
	if !isPendingListing {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isExpectingInput ||
		uint64(msg.ReplyToMessage.MessageId) != msgIDshouldReplyTo {
		return false, nil
	}

	pub, userConfig, err := c.createPublisher()
	defer func() {
		if err != nil {
			c.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				chatID, true, true, 0,
				fmt.Sprintf("The list request was canceled due to error, please retry later: %v", err),
			)
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Internal bot error: %v", err),
		)
		return true, nil
	}

	userConfig.SetAuthToken(strings.TrimSpace(*msg.Text))
	_, err = pub.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: auth error: %v", pub.Name(), err),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	posts, err := pub.List()
	if err != nil && len(posts) == 0 {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: unable to list posts: %v", pub.Name(), err),
		)
		return true, err
	}

	c.ResolvePendingRequest(userID)

	messages := make([]string, 1)
	i := 0
	for _, p := range posts {
		line := fmt.Sprintf(`- <a href="%s">%s</a>`+"\n", p.URL, p.Title)
		if len(messages[i])+len(line) > 4096 {
			messages = append(messages, line)
			i++
		} else {
			messages[i] += line
		}
	}

	for _, msg := range messages {
		_, _ = c.sendTextMessage(
			chatID, true, true, 0, msg,
		)
	}

	// delete user provided token related messages
	c.scheduleMessageDelete(
		chatID, 10*time.Second,
		uint64(msg.MessageId),
		msgIDshouldReplyTo,
	)

	return true, nil
}

func (c *telegramBot) tryToHandleInputForDeleting(
	chatID uint64,
	userID uint64,
	msg *api.Message,
) (bool, error) {
	// check if it's a reply for conversion started by links to /delete
	req, isPendingDeleting := c.GetPendingDeleting(userID)
	if !isPendingDeleting {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isPendingDeleting || !isExpectingInput ||
		uint64(msg.ReplyToMessage.MessageId) != msgIDshouldReplyTo {
		return false, nil
	}

	pub, userConfig, err := c.createPublisher()
	defer func() {
		if err != nil {
			c.ResolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				chatID, true, true, 0,
				fmt.Sprintf("The delete request was canceled due to error, please retry later: %v", err),
			)
		}
	}()

	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Internal bot error: %v", err),
		)
		return true, nil
	}

	userConfig.SetAuthToken(strings.TrimSpace(*msg.Text))
	_, err = pub.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: auth error: %v", pub.Name(), err),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	err = pub.Delete(req.URLs...)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: unable to delete: %v", pub.Name(), err),
		)
		return true, err
	}

	c.ResolvePendingRequest(userID)

	_, _ = c.sendTextMessage(
		chatID, true, true, 0, "The post(s) has been deleted",
	)

	// delete user provided token related messages
	c.scheduleMessageDelete(
		chatID, 10*time.Second,
		uint64(msg.MessageId),
		msgIDshouldReplyTo,
	)

	return true, nil
}
