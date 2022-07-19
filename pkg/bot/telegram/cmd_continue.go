package telegram

import (
	"encoding/base64"
	"fmt"
	"time"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/stringhelper"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (c *tgBot) handleBotCmd_Continue(
	mc *messageContext,
	wf *bot.Workflow,
	cmd, params string,
) error {
	// mark this session as standby, wait for reply from bot private message
	chatID, userID := mc.src.Chat.ID(), mc.src.From.ID()

	if len(params) == 0 {
		mc.logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "missing param"))
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).Silent().Reply(mc.msg.GetID()),
			styling.Plain("Please specify the key of the session, e.g. "),
			styling.Code(cmd+" your-key"),
		)
		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))

		return nil
	}

	_, ok := c.sessions.GetActiveSession(chatID)
	if ok {
		mc.logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "already in a session"))
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Please end existing session before starting a new one."),
		)

		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))

		return nil
	}

	if !c.sessions.MarkSessionStandby(
		wf,
		userID,
		chatIDWrapper{chat: mc.src.Chat.InputPeer()},
		params,
		false,
		5*time.Minute,
	) {
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("You have already started a session with no token replied, please end that first"),
		)
		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))

		return nil
	}

	// skip login when not required
	pub, user, _ := wf.CreatePublisher()
	if user.NextCredential() == rt.LoginFlow_None {
		// no login required

		_, err := c.sessions.ActivateSession(wf, userID, chatID, pub)
		if err != nil {
			msgID, _ := c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Plain("You have already started a session before, please end that first"),
			)
			c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))

			return nil
		}

		defer func() {
			if err != nil {
				c.sessions.DeactivateSession(chatID)
			}
		}()

		pageHeader, err := wf.Generator.New(&mc.con, cmd, params)
		if err != nil {
			return fmt.Errorf("failed to render page header: %w", err)
		}

		note, err := pub.CreateNew(&mc.con, cmd, params, &pageHeader)
		if err != nil {
			return fmt.Errorf("failed to pre-publish page: %w", err)
		}

		switch {
		case !note.SendMessage.IsNil():
			mc.con.SendMessage(c.Context(), note.SendMessage.Get())
		}

		return nil
	}

	// login is required, redirect user to private dialog

	_, err := pub.CheckLogin(&mc.con, cmd, params, user)
	if err != nil {
		return nil
	}

	// base64-url({create | enter}:hex(userID):hex(chatID))
	userIDPart, chatIDPart := encodeUint64Hex(userID), encodeUint64Hex(chatID)

	_, err = c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()).
			Row(
				&tg.KeyboardButtonURL{
					Text: "Enter",
					URL: fmt.Sprintf(
						"https://t.me/%s?start=%s",
						c.username,
						base64.URLEncoding.EncodeToString(
							stringhelper.ToBytes[byte, byte](fmt.Sprintf("enter:%s:%s", userIDPart, chatIDPart)),
						),
					),
				},
			),
		styling.Plain("Enter your "),
		styling.Bold(wf.PublisherName()),
		styling.Plain(" token to continue this session."),
	)
	if err != nil {
		c.sessions.ResolvePendingRequest(userID)
	}
	return err
}
