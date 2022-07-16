package telegram

import (
	"bytes"
	"encoding/base64"
	"strings"
	"time"

	"arhat.dev/pkg/log"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/session"
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
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Welcome, need some help? send command "),
			styling.Code(wf.BotCommands.TextOf(bot.BotCmd_Help)),
			styling.Plain(" to show all commands."),
		)
		return nil
	}

	chatID := src.Chat.ID()
	userID := src.From.ID()

	createOrEnter, err := base64.URLEncoding.DecodeString(params)
	if err != nil {
		// this is just a notice for bad /start params
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Spoiler("I am alive."),
		)
		return nil
	}

	parts := strings.SplitN(string(createOrEnter), ":", 3)
	if len(parts) != 3 {
		// this is just a notice for bad /start params
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Spoiler("Told you, I'm alive."),
		)
		return nil
	}

	action := parts[0]

	originalUserID, err := decodeUint64Hex[rt.UserID](parts[1])
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Bad start params: "),
			styling.Bold(err.Error()),
		)
		return nil
	}

	// ensure same user
	if originalUserID != userID {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Bold("The link is not for you :("),
		)
		return nil
	}

	originalChatID, err := decodeUint64Hex[rt.ChatID](parts[2])
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)
		return nil
	}

	var (
		standbySession         *session.SessionRequest[chatIDWrapper]
		expectedOriginalChatID chatIDWrapper
	)

	switch action {
	case "create", "enter":
		var ok bool
		standbySession, ok = c.sessions.GetStandbySession(userID)
		if !ok {
			logger.I("bad start attempt",
				log.String("reason", "no session requested"),
				rt.LogOrigChatID(originalChatID),
				rt.LogOrigSenderID(originalUserID),
			)
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("No session requested."),
			)

			return nil
		}

		expectedOriginalChatID = standbySession.Data
	case "edit", "delete", "list":
		expectedOriginalChatID = chatIDWrapper{chat: src.Chat.InputPeer()}
	default:
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Unknown action."),
		)
		return nil
	}

	// delete `/start` message
	c.scheduleMessageDelete(&src.Chat, 5*time.Second, rt.MessageID(msg.GetID()))

	// defensive check, should not happen
	if expectedOriginalChatID.ID() != originalChatID {
		logger.E("unexpected chat id not match",
			log.Uint64("expected_orig_chat_id", uint64(expectedOriginalChatID.ID())),
			log.Uint64("actual_orig_chat_id", uint64(originalChatID)),
		)

		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Unexpected chat id not match."),
		)
		return nil
	}

	origPeer := src.Chat.InputPeer()
	// check if user redirected to this chat
	if chatID != expectedOriginalChatID.ID() { // redirected
		origPeer = expectedOriginalChatID.chat
	}

	switch action {
	case "create":
		pub, userConfig, err2 := wf.CreatePublisher()
		defer func() {
			if err2 != nil {
				_, _ = c.sessions.ResolvePendingRequest(userID)

				// best effort
				_, _ = c.sendTextMessage(
					c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent(),
					styling.Plain("The session was canceled due to error, please retry later: "),
					styling.Bold(err2.Error()),
				)

				if standbySession.Data.ID() != chatID {
					_, _ = c.sendTextMessage(
						c.sender.To(origPeer).NoWebpage().Silent(),
						styling.Plain("The session was canceled due to error, please retry later."),
					)
				}
			}
		}()

		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold(err2.Error()),
			)
			return err2
		}

		token, err2 := pub.Login(userConfig)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Bold(pub.Name()),
				styling.Plain(" login failed: "),
				styling.Bold(err2.Error()),
			)
			return err2
		}

		_, err2 = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage(),
			styling.Plain("Here is your "),
			styling.Bold(pub.Name()),
			styling.Plain(" token, keep it for future use:\n\n"),
			styling.Code(token),
		)
		if err2 != nil {
			// best effort
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold("unable to send auth token."),
			)
			return err2
		}

		content, err2 := wf.Generator.RenderPageHeader()
		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Reply(msg.GetID()),
				styling.Plain("Internal bot error: failed to render page header: "),
				styling.Bold(err2.Error()),
			)

			return err2
		}

		var (
			rd bytes.Reader
			in rt.Input
		)
		rd.Reset(content)
		in = rt.NewInput(rd.Size(), &rd)

		note, err2 := pub.Publish(standbySession.Topic, &in)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Bold(pub.Name()),
				styling.Plain(" publishing not working: "),
				styling.Bold(err2.Error()),
			)

			return err2
		}

		_, err2 = c.sessions.ActivateSession(wf, userID, standbySession.Data.ID(), pub)
		if err2 != nil {
			logger.I("session activation error", log.Error(err2))
			_, _ = c.sendTextMessage(
				c.sender.To(origPeer).NoWebpage().Silent(),
				styling.Plain("Session not activated: "),
				styling.Bold(err2.Error()),
			)

			return err2
		}

		defer func(chatID rt.ChatID) {
			if err2 != nil {
				// bset effort
				_, _ = c.sessions.DeactivateSession(chatID)
			}
		}(standbySession.Data.ID())

		// error checked by `defer` section
		_, err2 = c.sendTextMessage(
			c.sender.To(origPeer).NoWebpage().Silent(),
			translateSpans(note)...,
		)

		return nil
	case "enter":
		msgID, err2 := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().
				Markup(&tg.ReplyKeyboardForceReply{
					SingleUse:   true,
					Selective:   true,
					Placeholder: wf.PublisherName() + " token",
				}),
			styling.Plain("Reply this message with your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token."),
		)
		if err2 != nil {
			// this message must be sent to user, this error will trigger message redelivery
			return err2
		}

		if !c.sessions.MarkRequestExpectingInput(userID, msgID) {
			msgID2, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Reply(msg.GetID()),
				styling.Plain("The session is not expecting any input."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, msgID2)

			return nil
		}

		return nil
	case "edit", "delete", "list":
		msgID, err2 := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().
				Markup(&tg.ReplyKeyboardForceReply{
					SingleUse:   true,
					Selective:   true,
					Placeholder: wf.PublisherName() + " token",
				}),
			styling.Plain("Reply this message with your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token."),
		)

		if !c.sessions.MarkRequestExpectingInput(userID, msgID) {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent(),
				styling.Plain("Internal bot error: "),
				styling.Bold("no pending request."),
			)

			c.scheduleMessageDelete(&src.Chat, 100*time.Millisecond, msgID)

			return nil
		}

		// this message must be sent to user, when the error is not nil
		// telegram will redeliver message
		return err2
	default:
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Spoiler("Nice try!"),
		)

		return nil
	}
}
