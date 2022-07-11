package telegram

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/queue"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/manager"
)

type msgDeleteKey struct{ chatID, msgID uint64 }

var _ bot.Interface = (*tgBot)(nil)

type tgBot struct {
	bot.BaseBot

	botToken    string
	botUsername string // set when Configure() called

	client     telegram.Client
	dispatcher tg.UpdateDispatcher
	sender     message.Sender
	downloader downloader.Downloader
	uploader   uploader.Uploader

	manager.SessionManager[*Message]

	wfSet         bot.WorkflowSet
	webhookConfig webhookConfig

	msgDelQ queue.TimeoutQueue[msgDeleteKey, tg.InputPeerClass]
}

// nolint:gocyclo
func (c *tgBot) handleNewMessage(src *messageSource, msg *tg.Message) error {
	if src.From.IsNil() {
		return fmt.Errorf("unsupport anonymous message")
	}

	logger := c.Logger().WithFields(
		log.Int64("chat_id", src.Chat.ID()),
		log.Int64("sender_id", src.From.GetPtr().ID()),
	)

	entities, ok := msg.GetEntities()
	if len(msg.GetMessage()) == 0 || !ok || len(entities) == 0 {
		return c.appendSessionMessage(logger, src, msg)
	}

	var (
		isCmd   bool
		cmd     string
		params  string
		mention string
	)

	for _, v := range entities {
		e, ok := v.(*tg.MessageEntityBotCommand)
		if !ok {
			continue
		}

		isCmd = true
		cmd = msg.Message[e.GetOffset() : e.GetOffset()+e.GetLength()]
		cmd, mention, ok = strings.Cut(cmd, "@")
		if ok /* found '@' */ && mention != c.botUsername {
			continue
		}

		params = strings.TrimSpace(msg.Message[e.GetOffset()+e.GetLength():])
		break
	}

	if isCmd {
		return c.handleCmd(logger, src, cmd, params, msg)
	}

	// not containing bot command
	// filter private message for special replies to this bot
	if src.Chat.IsPrivateChat() {
		if replyTo, ok := msg.GetReplyTo(); ok {
			for _, handle := range [...]tokenInputHandleFunc{
				c.tryToHandleInputForDiscussOrContinue,
				c.tryToHandleInputForEditing,
				c.tryToHandleInputForListing,
				c.tryToHandleInputForDeleting,
			} {
				handled, err := handle(src, msg, replyTo.GetReplyToMsgID())
				if handled {
					return err
				}
			}
		}
	}

	return c.appendSessionMessage(logger, src, msg)
}

func (c *tgBot) appendSessionMessage(
	logger log.Interface, src *messageSource, msg *tg.Message,
) (err error) {
	s, ok := c.GetActiveSession(uint64(src.Chat.ID()))
	if !ok {
		return nil
	}

	logger.V("append session message")
	m := newTelegramMessage(src, msg, s.RefMessages())

	errCh, err := c.preprocessMessage(s.Workflow(), &m, msg)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Internal bot error: message unhandled: "),
			styling.Bold(err.Error()),
		)
		return
	}

	if errCh != nil {
		go func(msgID int) {
			defer func() {

			}()
			for errProcessing := range errCh {
				// best effort, no error check
				_, _ = c.sendTextMessage(
					c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msgID),
					styling.Plain("Internal bot error: failed to process that message: "),
					styling.Bold(errProcessing.Error()),
				)
			}
		}(msg.GetID())
	}

	s.AppendMessage(&m)

	return nil
}

// handleCmd handle single command with all params as a single string
// nolint:gocyclo
func (c *tgBot) handleCmd(
	logger log.Interface,
	src *messageSource,
	cmd, params string,
	msg *tg.Message,
) error {
	if src.From.IsNil() {
		msgID, err := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Anonymous user not allowed"),
		)

		c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
		return err
	}

	chatID := uint64(src.Chat.ID())
	userID := uint64(src.From.GetPtr().ID())

	// ensure only group admin can start session
	if !src.Chat.IsPrivateChat() {
		// ensure only admin can use this bot in group

		user, err := c.client.API().UsersGetFullUser(c.Context(), &tg.InputUserFromMessage{
			Peer:   src.Chat.InputPeer(),
			MsgID:  msg.GetID(),
			UserID: src.From.GetPtr().ID(),
		})
		if err != nil {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold("unable to check status of the initiator."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
			return fmt.Errorf("check admin of user %d for chat %d", userID, chatID)
		}

		uChats := user.GetChats()
		var isAdmin bool
		for _, c := range uChats {
			if uint64(c.GetID()) != chatID {
				continue
			}

			switch ct := c.(type) {
			case *tg.Chat:
				_, isAdmin = ct.GetAdminRights()
			case *tg.Channel:
				_, isAdmin = ct.GetAdminRights()
			case *tg.ChatEmpty:
			case *tg.ChatForbidden:
			case *tg.ChannelForbidden:
			}

			break
		}

		if !isAdmin {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Only administrators can use this bot in group chat"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
			return nil
		}
	}

	wf, ok := c.wfSet.WorkflowFor(cmd)
	if !ok {
		msgID, _ := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Unknown command: "),
			styling.Code(cmd),
		)

		c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
		return nil
	}

	switch bc := wf.BotCommands.Parse(cmd); bc {
	case bot.BotCmd_Discuss, bot.BotCmd_Continue:
		// mark this session as standby, wait for reply from bot private message
		var (
			topic, url       string
			onInvalidCmdResp []styling.StyledTextOption
		)

		switch bc {
		case bot.BotCmd_Discuss:
			topic = params
			onInvalidCmdResp = []styling.StyledTextOption{
				styling.Plain("Please specify a session topic, e.g. "),
				styling.Code(cmd + " foo"),
			}
		case bot.BotCmd_Continue:
			url = params
			// nolint:lll
			onInvalidCmdResp = []styling.StyledTextOption{
				styling.Plain("Please specify the key of the session, e.g. "),
				styling.Code(cmd + " your-key"),
			}
		}

		if len(params) == 0 {
			logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "missing param"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				onInvalidCmdResp...,
			)
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))

			return nil
		}

		_, ok := c.GetActiveSession(uint64(chatID))
		if ok {
			logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "already in a session"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Please end existing session before starting a new one."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))

			return nil
		}

		if !c.MarkSessionStandby(wf, uint64(userID), uint64(chatID), topic, url, 5*time.Minute) {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You have already started a session with no token replied, please end that first"),
			)
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))

			return nil
		}

		if !wf.PublisherRequireLogin() {
			pub, _, _ := wf.CreatePublisher()

			_, err := c.ActivateSession(wf, uint64(chatID), uint64(userID), pub)
			if err != nil {
				msgID, _ := c.sendTextMessage(
					c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
					styling.Plain("You have already started a session before, please end that first"),
				)
				c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))

				return nil
			}

			defer func() {
				if err != nil {
					c.DeactivateSession(uint64(chatID))
				}
			}()

			pageHeader, err := wf.Generator.RenderPageHeader()
			if err != nil {
				return fmt.Errorf("failed to render page header: %w", err)
			}

			note, err := pub.Publish(topic, pageHeader)
			if err != nil {
				return fmt.Errorf("failed to pre-publish page: %w", err)
			}

			if len(note) != 0 {
				_, _ = c.sendTextMessage(
					c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent(),
					translateEntities(note)...,
				)
			}

			return nil
		}

		return func() (err error) {
			defer func() {
				if err != nil {
					c.ResolvePendingRequest(uint64(userID))
				}
			}()

			// base64-url({create | enter}:hex(userID):hex(chatID))
			userIDPart := encodeUint64Hex(uint64(userID))
			chatIDPart := encodeUint64Hex(uint64(chatID))

			urlForCreate := fmt.Sprintf(
				"https://t.me/%s?start=%s",
				c.botUsername,
				base64.URLEncoding.EncodeToString(
					[]byte(fmt.Sprintf("create:%s:%s", userIDPart, chatIDPart)),
				),
			)

			urlForEnter := fmt.Sprintf(
				"https://t.me/%s?start=%s",
				c.botUsername,
				base64.URLEncoding.EncodeToString(
					[]byte(fmt.Sprintf("enter:%s:%s", userIDPart, chatIDPart)),
				),
			)

			var (
				buttons    []tg.KeyboardButtonClass
				textPrompt []styling.StyledTextOption
			)

			switch bc {
			case bot.BotCmd_Discuss:
				textPrompt = []styling.StyledTextOption{
					styling.Plain("Create or enter your "),
					styling.Bold(wf.PublisherName()),
					styling.Plain(" token for this session."),
				}
				buttons = append(buttons, &tg.KeyboardButtonURL{
					Text: "Create",
					URL:  urlForCreate,
				})
			case bot.BotCmd_Continue:
				textPrompt = []styling.StyledTextOption{
					styling.Plain("Enter your "),
					styling.Bold(wf.PublisherName()),
					styling.Plain(" token to continue this session."),
				}
			}

			buttons = append(buttons, &tg.KeyboardButtonURL{
				Text: "Enter",
				URL:  urlForEnter,
			})

			_, err = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().
					Reply(msg.GetID()).Markup(&tg.ReplyKeyboardMarkup{
					SingleUse: true,
					Selective: true,
					Rows: []tg.KeyboardButtonRow{
						{
							Buttons: buttons,
						},
					},
				}),
				textPrompt...,
			)
			return err
		}()
	case bot.BotCmd_Start:
		if !src.Chat.IsPrivateChat() {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You cannot "),
				styling.Code("/start"),
				styling.Plain(" this bot in group chat"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
			return nil
		}

		return c.handleStartCommand(logger, wf, src, params, msg)
	case bot.BotCmd_Cancel:
		prevReq, ok := c.ResolvePendingRequest(uint64(userID))
		if ok {
			// a pending request, no generator involved
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You have canceled the pending "),
				styling.Code(wf.BotCommands.TextOf(manager.GetCommandFromRequest(prevReq))),
				styling.Plain(" request"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))

			if sr, isSR := prevReq.(*manager.SessionRequest); isSR {
				if sr.ChatID != uint64(src.Chat.ID()) {
					_, _ = c.sendTextMessage(
						c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent(),
						styling.Plain("Session canceled by the initiator."),
					)
				}
			}
		} else {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("There is no pending request"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
		}

		return nil
	case bot.BotCmd_End:
		currentSession, ok := c.GetActiveSession(uint64(src.Chat.ID()))
		if !ok {
			// TODO
			logger.D("invalid usage of end", log.String("reason", "no active session"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("There is no active session"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))

			return nil
		}

		msgs := currentSession.GetMessages()
		content, err := bot.GenerateContent(wf.Generator, msgs)
		if err != nil {
			logger.I("failed to generate post content", log.Error(err))
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: failed to generate post content: "),
				styling.Bold(err.Error()),
			)

			// do not execute again on telegram redelivery
			return nil
		}

		pub := currentSession.GetPublisher()
		note, err := pub.Append(c.Context(), content)
		if err != nil {
			logger.I("failed to append content to post", log.Error(err))
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Bold(pub.Name()),
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

		_, ok = c.DeactivateSession(uint64(chatID))
		if !ok {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: active session already been ended out of no reason"),
			)

			return nil
		}

		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards(),
			translateEntities(note)...,
		)

		return nil
	case bot.BotCmd_Include:
		replyTo, ok := msg.GetReplyTo()
		if !ok {
			logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "not a reply"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Code(cmd),
				styling.Plain(" can only be used as a reply"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID))
			return nil
		}

		_, ok = c.GetActiveSession(uint64(src.Chat.ID()))
		if !ok {
			logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "not in a session"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Plain("There is no active session, "),
				styling.Code(cmd),
				styling.Plain(" will do nothing in this case."),
			)
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
			return nil
		}

		history, err := c.client.API().MessagesGetHistory(c.Context(), &tg.MessagesGetHistoryRequest{
			Peer:     src.Chat.InputPeer(),
			OffsetID: replyTo.GetReplyToMsgID(),
			Limit:    1,
		})
		if err != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold("unable to get that message"),
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

		toAppend, err := expectExactlySingleMessage(hmsgs)
		if err != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold(err.Error()),
			)

			return nil
		}

		err = c.appendSessionMessage(logger, src, toAppend)
		if err != nil {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Failed to include that message"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
			return err
		}

		msgID, _ := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Done."),
		)

		c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID))
		return nil
	case bot.BotCmd_Ignore:
		replyTo, ok := msg.GetReplyTo()
		if !ok {
			logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "not a reply"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Code(cmd),
				styling.Plain(" can only be used as a reply"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
			return nil
		}

		currentSession, ok := c.GetActiveSession(uint64(src.Chat.ID()))
		if !ok {
			logger.D("invalid command usage", log.String("reason", "not in a session"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Plain("There is not active session, "),
				styling.Code(cmd),
				styling.Plain(" will do nothing in this case."),
			)
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))

			return nil
		}

		_ = currentSession.DeleteMessage(uint64(replyTo.GetReplyToMsgID()))

		logger.V("ignored message")
		msgID, _ := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
			styling.Plain("Done."),
		)

		c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID))

		return nil
	case bot.BotCmd_Edit:
		if !src.Chat.IsPrivateChat() {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You cannot use "),
				styling.Code(cmd),
				styling.Plain(" command in group chat"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
			return nil
		}

		prevCmd, ok := c.MarkPendingEditing(wf, userID, 5*time.Minute)
		if !ok {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You have pending "),
				styling.Code(wf.BotCommands.TextOf(prevCmd)),
				styling.Plain(" request not finished"),
			)
			return nil
		}

		// base64-url(edit:hex(userID):hex(chatID))
		userIDPart := encodeUint64Hex(userID)
		chatIDPart := encodeUint64Hex(chatID)

		urlForEdit := fmt.Sprintf(
			"https://t.me/%s?start=%s",
			c.botUsername,
			base64.URLEncoding.EncodeToString(
				[]byte(fmt.Sprintf("edit:%s:%s", userIDPart, chatIDPart)),
			),
		)

		_, err := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().
				Reply(msg.GetID()).Markup(&tg.ReplyKeyboardMarkup{
				SingleUse: true,
				Selective: true,
				Rows: []tg.KeyboardButtonRow{
					{
						Buttons: []tg.KeyboardButtonClass{
							&tg.KeyboardButtonURL{
								Text: "Enter",
								URL:  urlForEdit,
							},
						},
					},
				},
			}),
			styling.Plain("Enter your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token to edit"),
		)

		return err
	case bot.BotCmd_Delete:
		if !src.Chat.IsPrivateChat() {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You cannot use "),
				styling.Code(cmd),
				styling.Plain(" command in group chat"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
			return nil
		}

		if len(params) == 0 {
			logger.D("invalid command usage", log.String("reason", "missing param"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().Reply(msg.GetID()),
				styling.Plain("Please specify the url(s) of the "),
				styling.Bold(wf.PublisherName()),
				styling.Plain(" post(s) to be deleted"),
			)
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))

			return nil
		}

		prevCmd, ok := c.MarkPendingDeleting(wf, userID, strings.Split(params, " "), 5*time.Minute)
		if !ok {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You have pending "),
				styling.Code(wf.BotCommands.TextOf(prevCmd)),
				styling.Plain(" request not finished"),
			)
			return nil
		}

		// base64-url(delete:hex(userID):hex(chatID))
		userIDPart := encodeUint64Hex(userID)
		chatIDPart := encodeUint64Hex(chatID)

		urlForDelete := fmt.Sprintf(
			"https://t.me/%s?start=%s",
			c.botUsername,
			base64.URLEncoding.EncodeToString(
				[]byte(fmt.Sprintf("delete:%s:%s", userIDPart, chatIDPart)),
			),
		)

		_, err := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().
				Reply(msg.GetID()).Markup(&tg.ReplyKeyboardMarkup{
				SingleUse: true,
				Selective: true,
				Rows: []tg.KeyboardButtonRow{
					{
						Buttons: []tg.KeyboardButtonClass{
							&tg.KeyboardButtonURL{
								Text: "Enter",
								URL:  urlForDelete,
							},
						},
					},
				},
			}),
			styling.Plain("Enter your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token to delete the post"),
		)

		return err
	case bot.BotCmd_List:
		if !src.Chat.IsPrivateChat() {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You cannot use "),
				styling.Code(cmd),
				styling.Plain(" command in groups"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, uint64(msgID), uint64(msg.GetID()))
			return nil
		}

		prevCmd, ok := c.MarkPendingListing(wf, userID, 5*time.Minute)
		if !ok {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You have pending "),
				styling.Code(wf.BotCommands.TextOf(prevCmd)),
				styling.Plain(" request not finished"),
			)
			return nil
		}

		// base64-url(list:hex(userID):hex(chatID))
		userIDPart := encodeUint64Hex(userID)
		chatIDPart := encodeUint64Hex(chatID)

		urlForList := fmt.Sprintf(
			"https://t.me/%s?start=%s",
			c.botUsername,
			base64.URLEncoding.EncodeToString(
				[]byte(fmt.Sprintf("list:%s:%s", userIDPart, chatIDPart)),
			),
		)

		_, err := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().
				Reply(msg.GetID()).Markup(&tg.ReplyKeyboardMarkup{
				SingleUse: true,
				Selective: true,
				Rows: []tg.KeyboardButtonRow{
					{
						Buttons: []tg.KeyboardButtonClass{
							&tg.KeyboardButtonURL{
								Text: "Enter",
								URL:  urlForList,
							},
						},
					},
				},
			}),
			styling.Plain("Enter your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token to list your posts"),
		)

		return err
	case bot.BotCmd_Help:
		var body []styling.StyledTextOption
		body = append(body, styling.Plain("Usage:\n\n"))

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

		body = append(body, styling.Plain("\n"))

		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent(),
			body...,
		)
		return nil
	default:
		logger.D("unknown cmd")

		_, err := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoForwards().NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Command "),
			styling.Code(cmd),
			styling.Plain(" is not supported."),
		)

		return err
	}
}

func expectExactlySingleMessage(msgs []tg.MessageClass) (*tg.Message, error) {
	if len(msgs) != 1 {
		return nil, fmt.Errorf("not single message (%d)", len(msgs))
	}

	switch t := msgs[0].(type) {
	case *tg.Message:
		return t, nil
	default:
		return nil, fmt.Errorf("unexpected message type %T", t)
	}
}
