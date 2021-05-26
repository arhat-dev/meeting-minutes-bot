package server

import (
	"fmt"
	"strings"
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
)

func (c *telegramBot) tryToHandleInputForDiscussOrContinue(
	chatID uint64,
	userID uint64,
	msg *telegram.Message,
) (bool, error) {
	standbySession, hasStandbySession := c.getStandbySession(userID)
	if !hasStandbySession {
		return false, nil
	}

	msgIDShouldReplyTo, isExpectingInput := standbySession.GetMessageIDShouldReplyTo()
	if !isExpectingInput ||
		uint64(msg.ReplyToMessage.MessageId) != msgIDShouldReplyTo {
		return false, nil
	}

	gen, userConfig, err := c.createGenerator()
	defer func() {
		if err != nil {
			_, _ = c.resolvePendingRequest(userID)

			// best effort
			_, _ = c.sendTextMessage(
				chatID, true, true, 0,
				fmt.Sprintf("The discussion was canceled due to error, please retry later: %v", err),
			)

			if standbySession.ChatID != chatID {
				_, _ = c.sendTextMessage(
					standbySession.ChatID, true, true, 0,
					"The discussion was canceled due to error, please retry later",
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
	_, err = gen.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: auth error: %v", gen.Name(), err),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	var title string
	switch {
	case len(standbySession.Topic) != 0:
		// is /discuss, create a new post
		title = standbySession.Topic

		content, err2 := gen.FormatPagePrefix()
		if err2 != nil {
			return true, fmt.Errorf("failed to generate initial page: %w", err2)
		}

		postURL, err2 := gen.Publish(title, content)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, 0,
				fmt.Sprintf("%s pre-publish failed: %v", gen.Name(), err2),
			)
			return true, err2
		}

		_, err2 = c.sendTextMessage(
			standbySession.ChatID, true, true, 0,
			fmt.Sprintf(
				"The post for your discussion around %q has been created: %s",
				title, postURL,
			),
		)
		if err2 != nil {
			return true, err2
		}
	case len(standbySession.URL) != 0:
		// is /continue, find existing post to edit
		var err2 error

		// we may not find the post if user provided a wrong url, don't count this error
		// as session error
		title, err2 = gen.Retrieve(standbySession.URL)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Retrieve %s post error: %v", gen.Name(), err2),
			)
			return true, nil
		}
	}

	_, err = c.activateSession(
		standbySession.ChatID, userID, title,
		standbySession.ChatUsername, gen,
	)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Internal bot error: %v", err),
		)
		return true, err
	}

	c.resolvePendingRequest(userID)
	defer func() {
		if err != nil {
			// bset effort
			_, _ = c.deactivateSession(standbySession.ChatID)
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

	_, _ = c.sendTextMessage(
		standbySession.ChatID, true, true, 0,
		"You can start your discussion now, the post will be updated after the discussion",
	)

	return true, nil
}

func (c *telegramBot) tryToHandleInputForEditing(
	chatID uint64,
	userID uint64,
	msg *telegram.Message,
) (bool, error) {
	// check if it's a reply for conversion started by links to /edit, /delete, /list
	req, isPendingEditing := c.getPendingEditing(userID)
	if !isPendingEditing {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isExpectingInput ||
		uint64(msg.ReplyToMessage.MessageId) != msgIDshouldReplyTo {
		return false, nil
	}

	gen, userConfig, err := c.createGenerator()
	defer func() {
		if err != nil {
			c.resolvePendingRequest(userID)

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
	_, err = gen.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: auth error: %v", gen.Name(), err),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	authURL, err := gen.AuthURL()
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: unable to get auth url: %v", gen.Name(), err),
		)
		return true, err
	}

	c.resolvePendingRequest(userID)

	msgID, _ := c.sendTextMessage(
		chatID, true, true, 0, "Login Success!",
		telegram.InlineKeyboardMarkup{
			InlineKeyboard: [][]telegram.InlineKeyboardButton{{{
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
	msg *telegram.Message,
) (bool, error) {
	// check if it's a reply for conversion started by links to /list
	req, isPendingListing := c.getPendingListing(userID)
	if !isPendingListing {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isExpectingInput ||
		uint64(msg.ReplyToMessage.MessageId) != msgIDshouldReplyTo {
		return false, nil
	}

	gen, userConfig, err := c.createGenerator()
	defer func() {
		if err != nil {
			c.resolvePendingRequest(userID)

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
	_, err = gen.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: auth error: %v", gen.Name(), err),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	posts, err := gen.List()
	if err != nil && len(posts) == 0 {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: unable to list posts: %v", gen.Name(), err),
		)
		return true, err
	}

	c.resolvePendingRequest(userID)

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
	msg *telegram.Message,
) (bool, error) {
	// check if it's a reply for conversion started by links to /delete
	req, isPendingDeleting := c.getPendingDeleting(userID)
	if !isPendingDeleting {
		return false, nil
	}

	msgIDshouldReplyTo, isExpectingInput := req.GetMessageIDShouldReplyTo()
	if !isPendingDeleting || !isExpectingInput ||
		uint64(msg.ReplyToMessage.MessageId) != msgIDshouldReplyTo {
		return false, nil
	}

	gen, userConfig, err := c.createGenerator()
	defer func() {
		if err != nil {
			c.resolvePendingRequest(userID)

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
	_, err = gen.Login(userConfig)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: auth error: %v", gen.Name(), err),
		)
		// usually not our fault, let user try again
		err = nil
		return true, nil
	}

	err = gen.Delete(req.urls...)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("%s: unable to delete: %v", gen.Name(), err),
		)
		return true, err
	}

	c.resolvePendingRequest(userID)

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
