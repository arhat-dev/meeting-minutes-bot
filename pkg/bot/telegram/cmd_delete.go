package telegram

import (
	"encoding/base64"
	"fmt"
	"time"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/pkg/log"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (c *tgBot) handleBotCmd_Delete(
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

	if len(params) == 0 {
		mc.logger.D("invalid command usage", log.String("reason", "missing param"))
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).Reply(mc.msg.GetID()),
			styling.Plain("Please specify the url(s) of the "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" post(s) to be deleted."),
		)
		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))

		return nil
	}

	chatID, userID := mc.src.Chat.ID(), mc.src.From.ID()

	prevCmd, ok := c.sessions.MarkPendingDeleting(wf, userID, params, 5*time.Minute)
	if !ok {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("You have pending "),
			styling.Code(wf.BotCommands.TextOf(prevCmd)),
			styling.Plain(" request not finished."),
		)
		return nil
	}

	// base64-url(delete:hex(userID):hex(chatID))
	userIDPart := encodeUint64Hex(userID)
	chatIDPart := encodeUint64Hex(chatID)

	urlForDelete := fmt.Sprintf(
		"https://t.me/%s?start=%s",
		c.username,
		base64.URLEncoding.EncodeToString(
			[]byte(fmt.Sprintf("delete:%s:%s", userIDPart, chatIDPart)),
		),
	)

	_, err := c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()).
			Row(&tg.KeyboardButtonURL{
				Text: "Enter",
				URL:  urlForDelete,
			}),
		styling.Plain("Enter your "),
		styling.Bold(wf.PublisherName()),
		styling.Plain(" token to delete the post."),
	)

	return err
}
