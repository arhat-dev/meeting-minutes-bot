package telegram

import (
	"encoding/base64"
	"fmt"
	"time"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (c *tgBot) handleBotCmd_Edit(
	mc *messageContext,
	wf *bot.Workflow,
	cmd, params string,
) error {
	if !mc.src.Chat.IsPrivateChat() {
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("You cannot use "),
			styling.Code(cmd),
			styling.Plain(" command in group chat."),
		)

		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))
		return nil
	}

	chatID, userID := mc.src.Chat.ID(), mc.src.From.ID()

	prevCmd, ok := c.sessions.MarkPendingEditing(wf, userID, 5*time.Minute)
	if !ok {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("You have pending "),
			styling.Code(wf.BotCommands.TextOf(prevCmd)),
			styling.Plain(" request not finished."),
		)

		return nil
	}

	// base64-url(edit:hex(userID):hex(chatID))
	userIDPart := encodeUint64Hex(userID)
	chatIDPart := encodeUint64Hex(chatID)

	urlForEdit := fmt.Sprintf(
		"https://t.me/%s?start=%s",
		c.username,
		base64.URLEncoding.EncodeToString(
			[]byte(fmt.Sprintf("edit:%s:%s", userIDPart, chatIDPart)),
		),
	)

	_, err := c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()).
			Row(&tg.KeyboardButtonURL{
				Text: "Enter",
				URL:  urlForEdit,
			}),
		styling.Plain("Enter your "),
		styling.Bold(wf.PublisherName()),
		styling.Plain(" token to edit."),
	)

	return err
}
