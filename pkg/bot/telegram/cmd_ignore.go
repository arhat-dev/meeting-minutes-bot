package telegram

import (
	"time"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/pkg/log"
	"github.com/gotd/td/telegram/message/styling"
)

func (c *tgBot) handleBotCmd_Ignore(
	mc *messageContext,
	wf *bot.Workflow,
	cmd, params string,
) error {
	replyTo, ok := mc.msg.GetReplyTo()
	if !ok {
		mc.logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "not a reply"))
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).Silent().Reply(mc.msg.GetID()),
			styling.Code(cmd),
			styling.Plain(" can only be used as a reply."),
		)

		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))
		return nil
	}

	chatID := mc.src.Chat.ID()

	currentSession, ok := c.sessions.GetActiveSession(chatID)
	if !ok {
		mc.logger.D("invalid command usage", log.String("reason", "not in a session"))
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).Silent().Reply(mc.msg.GetID()),
			styling.Plain("There is not active session, "),
			styling.Code(cmd),
			styling.Plain(" will do nothing in this case."),
		)
		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))

		return nil
	}

	_ = currentSession.DeleteMessage(rt.MessageID(replyTo.GetReplyToMsgID()))

	mc.logger.V("ignored message")
	msgID, _ := c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).Silent().Reply(mc.msg.GetID()),
		styling.Plain("Ignored."),
	)

	c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID)

	return nil
}
