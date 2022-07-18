package telegram

import (
	"time"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/pkg/log"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
)

func (c *tgBot) handleBotCmd_Include(
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

		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID)
		return nil
	}

	chatID := mc.src.Chat.ID()
	_, ok = c.sessions.GetActiveSession(chatID)
	if !ok {
		mc.logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "not in a session"))
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).Silent().Reply(mc.msg.GetID()),
			styling.Plain("There is no active session, "),
			styling.Code(cmd),
			styling.Plain(" will do nothing in this case."),
		)
		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))
		return nil
	}

	var (
		history tg.MessagesMessagesClass
		err     error
	)

	if c.isBot {
		if mc.src.Chat.IsPrivateChat() || mc.src.Chat.IsLegacyGroupChat() {
			history, err = c.client.API().MessagesGetMessages(c.Context(), []tg.InputMessageClass{
				&tg.InputMessageID{
					ID: replyTo.GetReplyToMsgID(),
				},
			})
		} else {
			peer := mc.src.Chat.InputPeer().(*tg.InputPeerChannel)
			history, err = c.client.API().ChannelsGetMessages(c.Context(), &tg.ChannelsGetMessagesRequest{
				Channel: &tg.InputChannel{
					ChannelID:  peer.ChannelID,
					AccessHash: peer.AccessHash,
				},
				ID: []tg.InputMessageClass{
					&tg.InputMessageID{
						ID: replyTo.GetReplyToMsgID(),
					},
				},
			})
		}
	} else {
		// message.getHistory is not allowed for bot user
		history, err = c.client.API().MessagesGetHistory(c.Context(), &tg.MessagesGetHistoryRequest{
			Peer:     mc.src.Chat.InputPeer(),
			OffsetID: replyTo.GetReplyToMsgID(),
			Limit:    1,
		})
	}

	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold("unable to get that message."),
		)

		return err
	}

	var hmsgs []tg.MessageClass
	switch h := history.(type) {
	case *tg.MessagesMessages:
		hmsgs = h.GetMessages()
	case *tg.MessagesMessagesSlice:
		hmsgs = h.GetMessages()
	case *tg.MessagesChannelMessages:
		hmsgs = h.GetMessages()
	case *tg.MessagesMessagesNotModified:
	}

	toAppend, err := expectExactOneMessage(hmsgs)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)

		return nil
	}

	nMC := *mc
	nMC.msg = toAppend
	err = c.appendSessionMessage(&nMC)
	if err != nil {
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Failed to include that message."),
		)

		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))
		return err
	}

	msgID, _ := c.sendTextMessage(
		c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
		styling.Plain("Included."),
	)

	c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID)
	return nil
}
