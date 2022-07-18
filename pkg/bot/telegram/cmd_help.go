package telegram

import (
	"github.com/gotd/td/telegram/message/styling"
)

func (c *tgBot) handleBotCmd_Help(mc *messageContext) error {
	body := []styling.StyledTextOption{
		styling.Plain("Usage:\n\n"),
	}

	for i := range c.wfSet.Workflows {
		wf := &c.wfSet.Workflows[i]

		for i, cmd := range wf.BotCommands.Commands {
			if len(cmd) == 0 || len(wf.BotCommands.Descriptions[i]) == 0 {
				continue
			}

			body = append(body, styling.Code(cmd))
			body = append(body, styling.Plain(" - "))
			body = append(body, styling.Plain(wf.BotCommands.Descriptions[i]))
			body = append(body, styling.Plain("\n"))
		}
	}

	_, _ = c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent(),
		append(body, styling.Plain("\n"))...,
	)
	return nil
}
