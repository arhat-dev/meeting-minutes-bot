package telegram

import (
	"encoding/base64"
	"strings"
	"time"

	"arhat.dev/pkg/log"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/manager"
)

func (c *tgBot) handleStartCommand(
	logger log.Interface,
	wf *bot.Workflow,
	src *messageSource,
	params string,
	msg *tg.Message,
) error {
	if len(params) == 0 {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Welcome, need some help? send command "),
			styling.Code(wf.BotCommands.TextOf(bot.BotCmd_Help)),
			styling.Plain(" to show all commands."),
		)
		return nil
	}

	chatID := uint64(src.Chat.ID())
	userID := uint64(src.From.GetPtr().ID())

	createOrEnter, err := base64.URLEncoding.DecodeString(params)
	if err != nil {
		// this is just a notice for bad /start params
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Spoiler("I am alive."),
		)
		return nil
	}

	parts := strings.SplitN(string(createOrEnter), ":", 3)
	if len(parts) != 3 {
		// this is just a notice for bad /start params
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Spoiler("Told you, I'm alive."),
		)
		return nil
	}

	action := parts[0]

	originalUserID, err := decodeUint64Hex(parts[1])
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Bad start params: "),
			styling.Bold(err.Error()),
		)
		return nil
	}

	// ensure same user
	if originalUserID != userID {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Bold("The link is not for you :("),
		)
		return nil
	}

	originalChatID, err := decodeUint64Hex(parts[2])
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
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
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("No session requested"),
			)
			return nil
		}

		expectedOriginalChatID = standbySession.ChatID
	case "edit", "delete", "list":
		expectedOriginalChatID = chatID
	default:
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Unknown action"),
		)
		return nil
	}

	// delete `/start` message
	c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msg.GetID()))

	if expectedOriginalChatID != originalChatID {
		// should not happen, defensive check
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Unexpected chat id not match"),
		)
		return nil
	}

	origPeer := src.Chat.InputPeer()
	if chatID != expectedOriginalChatID {
		cs, err := c.getChatByID(int64(expectedOriginalChatID))
		if err != nil {
			return err
		}

		origPeer = cs.InputPeer()
	}

	switch action {
	case "create":
		pub, userConfig, err2 := wf.CreatePublisher()
		defer func() {
			if err2 != nil {
				_, _ = c.ResolvePendingRequest(userID)

				// best effort
				_, _ = c.sendTextMessage(
					c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent(),
					styling.Plain("The session was canceled due to error, please retry later: "),
					styling.Bold(err2.Error()),
				)

				if standbySession.ChatID != chatID {
					_, _ = c.sendTextMessage(
						c.sender.To(origPeer).NoForwards().NoWebpage().Silent(),
						styling.Plain("The session was canceled due to error, please retry later"),
					)
				}
			}
		}()

		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold(err2.Error()),
			)
			return err2
		}

		token, err2 := pub.Login(userConfig)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Bold(pub.Name()),
				styling.Plain(" login failed: "),
				styling.Bold(err2.Error()),
			)
			return err2
		}

		_, err2 = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage(),
			styling.Plain("Here is your "),
			styling.Bold(pub.Name()),
			styling.Plain(" token, keep it on your own for later use:\n\n"),
			styling.Code(token),
		)
		if err2 != nil {
			// best effort
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold("unable to send auth token"),
			)
			return err2
		}

		content, err2 := wf.Generator.RenderPageHeader()
		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Reply(msg.GetID()),
				styling.Plain("Internal bot error: failed to render page header"),
				styling.Bold(err2.Error()),
			)

			return err2
		}

		note, err2 := pub.Publish(standbySession.Topic, content)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Bold(pub.Name()),
				styling.Plain(" publishing not working: "),
				styling.Bold(err2.Error()),
			)

			return err2
		}

		_, err2 = c.ActivateSession(wf, standbySession.ChatID, userID, pub)
		if err2 != nil {
			logger.I("session activation error", log.Error(err2))
			_, _ = c.sendTextMessage(
				c.sender.To(origPeer).NoForwards().NoWebpage().Silent(),
				styling.Plain("Session not activated: "),
				styling.Bold(err2.Error()),
			)

			return err2
		}

		defer func() {
			if err2 != nil {
				// bset effort
				_, _ = c.DeactivateSession(standbySession.ChatID)
			}
		}()

		// error checked by `defer` section
		_, err2 = c.sendTextMessage(
			c.sender.To(origPeer).NoForwards().NoWebpage().Silent(),
			translateEntities(note)...,
		)

		return nil
	case "enter":
		msgID, err2 := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Markup(&tg.ReplyKeyboardForceReply{
				SingleUse:   true,
				Selective:   true,
				Placeholder: wf.PublisherName() + " token",
			}),
			styling.Plain("Enter your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token as a reply to this message"),
		)
		if err2 != nil {
			// this message must be sent to user, this error will trigger message redelivery
			return err2
		}

		if !c.MarkRequestExpectingInput(userID, uint64(msgID)) {
			msgID2, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Reply(msg.GetID()),
				styling.Plain("The session is not expecting any input"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msgID2))

			return nil
		}

		return nil
	case "edit", "delete", "list":
		msgID, err2 := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().
				NoWebpage().Markup(&tg.ReplyKeyboardForceReply{
				SingleUse:   true,
				Selective:   true,
				Placeholder: wf.PublisherName() + " token",
			}),
			styling.Plain("Enter your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token as a reply to this message"),
		)

		if !c.MarkRequestExpectingInput(userID, uint64(msgID)) {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent(),
				styling.Plain("Internal bot error: "),
				styling.Bold("could not find your pending request"),
			)

			c.scheduleMessageDelete(&src.Chat, 100*time.Millisecond, uint64(msgID))

			return nil
		}

		// this message must be sent to user, when the error is not nil
		// telegram will redeliver message
		return err2
	default:
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Spoiler("Nice try!"),
		)

		return nil
	}
}
