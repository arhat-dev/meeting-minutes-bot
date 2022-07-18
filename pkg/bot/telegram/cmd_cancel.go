package telegram

import (
	"time"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/session"
	"github.com/gotd/td/telegram/message/styling"
)

func (c *tgBot) handleBotCmd_Cancel(
	mc *messageContext,
	wf *bot.Workflow,
	cmd, params string,
) error {
	userID := mc.src.From.ID()

	prevReq, ok := c.sessions.ResolvePendingRequest(userID)
	if ok {
		// a pending request, no generator involved
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("You have canceled the pending "),
			styling.Code(wf.BotCommands.TextOf(session.GetCommandFromRequest[chatIDWrapper](prevReq))),
			styling.Plain(" request"),
		)

		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))

		if sr, isSR := prevReq.(*session.SessionRequest[chatIDWrapper]); isSR {
			if sr.Data.ID() != mc.src.Chat.ID() {
				_, _ = c.sendTextMessage(
					c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent(),
					styling.Plain("Session canceled by the initiator."),
				)
			}
		}
	} else {
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("There is no pending request."),
		)

		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))
	}

	return nil
}
