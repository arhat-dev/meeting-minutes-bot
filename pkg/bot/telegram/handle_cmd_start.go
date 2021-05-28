package telegram

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"arhat.dev/pkg/log"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/manager"
)

func (c *telegramBot) handleStartCommand(
	logger log.Interface,
	chatID uint64,
	userID uint64,
	isPrivateMessage bool,
	params string,
	msg *telegram.Message,
) error {
	if !isPrivateMessage {
		msgID, _ := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"You cannot <code>/start</code> this bot in groups",
		)

		c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
		return nil
	}

	if len(params) == 0 {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "Welcome, need some <code>/help</code> ?")
		return nil
	}

	createOrEnter, err := base64.URLEncoding.DecodeString(params)
	if err != nil {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "I am alive.")
		return nil
	}

	parts := strings.SplitN(string(createOrEnter), ":", 3)
	if len(parts) != 3 {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "Told you, I'm alive.")
		return nil
	}

	action := parts[0]

	originalUserID, err := decodeUint64Hex(parts[1])
	if err != nil {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, fmt.Sprintf("Internal bot error: %s", err))
		return nil
	}

	// ensure same user
	if originalUserID != userID {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "The link is not for you :(")
		return nil
	}

	originalChatID, err := decodeUint64Hex(parts[2])
	if err != nil {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, fmt.Sprintf("Internal bot error: %s", err))
		return nil
	}

	var (
		standbySession         *manager.SessionRequest
		expectedOriginalChatID uint64
	)

	switch action {
	case "create", "enter":
		var ok bool
		standbySession, ok = c.GetStandbySession(userID)
		if !ok {
			_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "No discussion requested")
			return nil
		}

		expectedOriginalChatID = standbySession.ChatID
	case "edit", "delete", "list":
		expectedOriginalChatID = chatID
	default:
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "Unknown action")
		return nil
	}

	// delete `/start` message
	c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msg.MessageId))

	if expectedOriginalChatID != originalChatID {
		// should not happen, defensive check
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "Unexpected chat id not match")
		return nil
	}

	switch action {
	case "create":
		pub, userConfig, err2 := c.createPublisher()
		defer func() {
			if err2 != nil {
				_, _ = c.ResolvePendingRequest(userID)

				// best effort
				_, _ = c.sendTextMessage(
					chatID, true, true, 0,
					fmt.Sprintf("The discussion was canceled due to error, please retry later: %v", err2),
				)

				if standbySession.ChatID != chatID {
					_, _ = c.sendTextMessage(
						standbySession.ChatID, true, true, 0,
						"The discussion was canceled due to error, please retry later",
					)
				}
			}
		}()

		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Internal bot error: %v", err2),
			)
			return err2
		}

		token, err2 := pub.Login(userConfig)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("%s login failed: %v", pub.Name(), err2),
			)
			return err2
		}

		_, err2 = c.sendTextMessage(
			chatID, false, true, 0,
			fmt.Sprintf(
				"Here is your %s token, keep it on your own for later use:\n\n<pre>%s</pre>",
				pub.Name(), token,
			),
		)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, false, true, msg.MessageId,
				fmt.Sprintf("Internal bot error: unable to send %s token: %v", pub.Name(), err2),
			)
			return err2
		}

		content, err2 := c.generator.FormatPageHeader()
		if err2 != nil {
			return fmt.Errorf("failed to generate initial page: %w", err2)
		}

		postURL, err2 := pub.Publish(standbySession.Topic, content)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("%s pre-publish failed: %v", pub.Name(), err2),
			)
			return err2
		}

		currentSession, err2 := c.ActivateSession(
			standbySession.ChatID,
			userID,
			standbySession.Topic,
			pub,
		)
		if err2 != nil {
			logger.D("invalid usage of discuss", log.String("reason", err2.Error()))
			_, _ = c.sendTextMessage(
				standbySession.ChatID, true, true, 0,
				fmt.Sprintf("Could not activate discussion: %v", err2),
			)
			return err2
		}

		defer func() {
			if err2 != nil {
				// bset effort
				_, _ = c.DeactivateSession(standbySession.ChatID)
			}
		}()

		_, err2 = c.sendTextMessage(
			standbySession.ChatID, true, true, 0,
			fmt.Sprintf(
				"The post for your discussion around %q has been created: %s",
				currentSession.GetTopic(), postURL,
			),
		)

		return nil
	case "enter":
		msgID, err2 := c.sendTextMessage(chatID, false, true, 0,
			fmt.Sprintf("Enter your %s token as a reply to this message", c.publisherName),
			telegram.ForceReply{
				ForceReply: true,
				Selective:  constant.True(),
			},
		)
		if err2 != nil {
			// this message must be sent to user, this error will trigger message redelivery
			return err2
		}

		if !c.MarkRequestExpectingInput(userID, uint64(msgID)) {
			msgID2, _ := c.sendTextMessage(
				chatID, false, true, msg.MessageId,
				"The discussion is not expecting any input",
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msgID2))

			return nil
		}

		return nil
	case "edit", "delete", "list":
		msgID, err2 := c.sendTextMessage(
			chatID, true, true, 0,
			fmt.Sprintf("Enter your %s token as a reply to this message", c.publisherName),
			telegram.ForceReply{
				ForceReply: true,
				Selective:  constant.True(),
			},
		)

		if !c.MarkRequestExpectingInput(userID, uint64(msgID)) {
			_, _ = c.sendTextMessage(
				chatID, true, true, 0,
				"Internal bot error: could not find your pending request",
			)

			c.scheduleMessageDelete(chatID, 100*time.Millisecond, uint64(msgID))

			return nil
		}

		// this message must be sent to user, when the error is not nil
		// telegram will redeliver message
		return err2
	default:
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"Nice try!",
		)

		return nil
	}
}
