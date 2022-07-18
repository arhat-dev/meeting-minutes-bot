package telegram

import (
	"time"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/pkg/log"
	"github.com/gotd/td/telegram/message/styling"
)

func (c *tgBot) handleBotCmd_End(
	mc *messageContext,
	wf *bot.Workflow,
	cmd, params string,
) error {
	chatID := mc.src.Chat.ID()
	currentSession, ok := c.sessions.GetActiveSession(chatID)
	if !ok {
		// TODO
		mc.logger.D("invalid usage of end", log.String("reason", "no active session"))
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("There is no active session."),
		)

		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))

		return nil
	}

	msgs := currentSession.GetMessages()
	content, err := bot.GenerateContent(
		wf.Generator,
		&mc.con,
		wf.BotCommands.TextOf(rt.BotCmd_End),
		params,
		msgs,
	)
	if err != nil {
		mc.logger.I("failed to generate post content", log.Error(err))
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: failed to generate post content: "),
			styling.Bold(err.Error()),
		)

		// do not execute again on telegram redelivery
		return nil
	}

	pub := currentSession.GetPublisher()
	note, err := pub.AppendToExisting(
		&mc.con,
		wf.BotCommands.TextOf(rt.BotCmd_End),
		params,
		&content,
	)
	if err != nil {
		mc.logger.I("failed to append content to post", log.Error(err))
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).Silent().Reply(mc.msg.GetID()),
			styling.Bold(currentSession.Workflow().PublisherName()),
			styling.Plain(" post update error: "),
			styling.Bold(err.Error()),
		)

		// do not execute again on telegram redelivery
		return nil
	}

	for _, m := range msgs {
		m.Dispose()
	}

	currentSession.TruncMessages(len(msgs))

	_, ok = c.sessions.DeactivateSession(chatID)
	if !ok {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: active session already been ended out of no reason."),
		)

		return nil
	}

	switch {
	case !note.SendMessage.IsNil():
		mc.con.SendMessage(c.Context(), note.SendMessage.Get())
	}

	return nil
}
