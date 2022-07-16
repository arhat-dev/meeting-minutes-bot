package telegram

import (
	"bytes"
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

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/session"
)

var _ bot.Interface = (*tgBot)(nil)

type tgBot struct {
	bot.BaseBot

	isBot    bool
	botToken string
	username string // set when Configure() called

	client     *telegram.Client
	dispatcher tg.UpdateDispatcher
	sender     message.Sender
	downloader downloader.Downloader
	uploader   uploader.Uploader

	sessions session.Manager[chatIDWrapper]

	wfSet   bot.WorkflowSet
	msgDelQ queue.TimeoutQueue[msgDeleteKey, tg.InputPeerClass]
}

// nolint:gocyclo
func (c *tgBot) dispatchNewMessage(src *messageSource, msg *tg.Message) error {
	logger := c.Logger().WithFields(
		rt.LogChatID(src.Chat.ID()),
		rt.LogSenderID(src.From.ID()),
	)

	logger.V("dispatch message")
	if len(msg.GetMessage()) == 0 {
		return c.appendSessionMessage(logger, src, msg)
	}

	var (
		isCmd   bool
		cmd     string
		params  string
		mention string
	)

	entities, _ := msg.GetEntities()
	for _, v := range entities {
		e, ok := v.(*tg.MessageEntityBotCommand)
		if !ok {
			continue
		}

		isCmd = true
		cmd = msg.Message[e.GetOffset() : e.GetOffset()+e.GetLength()]
		cmd, mention, ok = strings.Cut(cmd, "@")
		if ok /* found '@' */ && mention != c.username {
			continue
		}

		params = strings.TrimSpace(msg.Message[e.GetOffset()+e.GetLength():])
		break
	}

	if isCmd {
		return c.handleBotCmd(logger, src, cmd, params, msg)
	}

	// not containing bot command
	// filter private message for special replies to this bot
	if src.Chat.IsPrivateChat() {
		replyTo, ok := msg.GetReplyTo()
		logger.V("check potential input", log.Bool("do_check", ok))
		if ok {
			replyToMsgID := rt.MessageID(replyTo.ReplyToMsgID)
			for _, handle := range [...]tokenInputHandleFunc{
				c.tryToHandleInputForDiscussOrContinue,
				c.tryToHandleInputForEditing,
				c.tryToHandleInputForListing,
				c.tryToHandleInputForDeleting,
			} {
				done, err := handle(logger, src, msg, replyToMsgID)
				if done {
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
	logger.V("check active session")
	s, ok := c.sessions.GetActiveSession(src.Chat.ID())
	if !ok {
		return nil
	}

	logger.V("append session message")
	m := newTelegramMessage(src, msg, s.RefMessages())

	errCh, err := c.preprocessMessage(s.Workflow(), &m, msg)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Internal bot error: message unhandled: "),
			styling.Bold(err.Error()),
		)
		return
	}

	if errCh != nil {
		go func(msgID int) {
			for errProcessing := range errCh {
				// best effort, no error check
				_, _ = c.sendTextMessage(
					c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msgID),
					styling.Plain("Internal bot error: failed to process that message: "),
					styling.Bold(errProcessing.Error()),
				)
			}
		}(msg.GetID())
	}

	s.AppendMessage(&m)

	return nil
}

// handleBotCmd handles single command with all params as a single string
// nolint:gocyclo
func (c *tgBot) handleBotCmd(
	logger log.Interface,
	src *messageSource,
	cmd, params string,
	msg *tg.Message,
) error {
	logger.V("handle bot command", log.String("cmd", cmd))

	if !src.From.IsUser() {
		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Only non-anonymous users can use this bot"),
		)

		return nil
	}

	chatID := src.Chat.ID()
	userID := src.From.ID()

	// ensure only group admin can start session
	if !src.Chat.IsPrivateChat() {
		// ensure only admin can use this bot in group

		if src.Chat.IsLegacyGroupChat() {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Unsupported chat type: please consider upgrading this chat to supergroup."),
			)
			return nil
		}

		pt, err := c.client.API().ChannelsGetParticipant(c.Context(), &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  src.Chat.InputPeer().(*tg.InputPeerChannel).GetChannelID(),
				AccessHash: src.Chat.InputPeer().(*tg.InputPeerChannel).GetAccessHash(),
			},
			Participant: &tg.InputPeerUser{
				UserID:     src.From.InputUser().GetUserID(),
				AccessHash: src.From.InputUser().GetAccessHash(),
			},
		})
		if err != nil || len(pt.Users) != 1 {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold("unable to check status of the initiator."),
			)
			return err
		}

		switch pt.GetParticipant().(type) {
		case *tg.ChannelParticipantCreator:
		case *tg.ChannelParticipantAdmin:
		default:
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Only administrators can use this bot in group chat"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
			return nil
		}
	}

	wf, ok := c.wfSet.WorkflowFor(cmd)
	if !ok {
		msgID, _ := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Unknown command: "),
			styling.Code(cmd),
		)

		c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
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
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))

			return nil
		}

		_, ok := c.sessions.GetActiveSession(chatID)
		if ok {
			logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "already in a session"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Please end existing session before starting a new one."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))

			return nil
		}

		if !c.sessions.MarkSessionStandby(wf, userID, chatIDWrapper{chat: src.Chat.InputPeer()}, topic, url, 5*time.Minute) {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You have already started a session with no token replied, please end that first"),
			)
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))

			return nil
		}

		// skip login when not required
		if !wf.PublisherRequireLogin() {
			pub, _, _ := wf.CreatePublisher()

			_, err := c.sessions.ActivateSession(wf, userID, chatID, pub)
			if err != nil {
				msgID, _ := c.sendTextMessage(
					c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
					styling.Plain("You have already started a session before, please end that first"),
				)
				c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))

				return nil
			}

			defer func() {
				if err != nil {
					c.sessions.DeactivateSession(chatID)
				}
			}()

			pageHeader, err := wf.Generator.RenderPageHeader()
			if err != nil {
				return fmt.Errorf("failed to render page header: %w", err)
			}

			var (
				rd bytes.Reader
				in rt.Input
			)
			rd.Reset(pageHeader)
			in = rt.NewInput(rd.Size(), &rd)

			note, err := pub.Publish(topic, &in)
			if err != nil {
				return fmt.Errorf("failed to pre-publish page: %w", err)
			}

			if len(note) != 0 {
				_, _ = c.sendTextMessage(
					c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent(),
					translateSpans(note)...,
				)
			}

			return nil
		}

		// login is required, redirect user to private message to enter credentials

		// base64-url({create | enter}:hex(userID):hex(chatID))
		userIDPart := encodeUint64Hex(userID)
		chatIDPart := encodeUint64Hex(chatID)

		urlForCreate := fmt.Sprintf(
			"https://t.me/%s?start=%s",
			c.username,
			base64.URLEncoding.EncodeToString(
				[]byte(fmt.Sprintf("create:%s:%s", userIDPart, chatIDPart)),
			),
		)

		urlForEnter := fmt.Sprintf(
			"https://t.me/%s?start=%s",
			c.username,
			base64.URLEncoding.EncodeToString(
				[]byte(fmt.Sprintf("enter:%s:%s", userIDPart, chatIDPart)),
			),
		)

		var (
			buttons []tg.KeyboardButtonClass
			prompt  []styling.StyledTextOption
		)

		switch bc {
		case bot.BotCmd_Discuss:
			prompt = []styling.StyledTextOption{
				styling.Plain("Create or enter your "),
				styling.Bold(wf.PublisherName()),
				styling.Plain(" token for this session."),
			}
			buttons = append(buttons, &tg.KeyboardButtonURL{
				Text: "Create",
				URL:  urlForCreate,
			})
		case bot.BotCmd_Continue:
			prompt = []styling.StyledTextOption{
				styling.Plain("Enter your "),
				styling.Bold(wf.PublisherName()),
				styling.Plain(" token to continue this session."),
			}
		}

		buttons = append(buttons, &tg.KeyboardButtonURL{
			Text: "Enter",
			URL:  urlForEnter,
		})

		_, err := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()).
				Row(buttons...),
			prompt...,
		)
		if err != nil {
			c.sessions.ResolvePendingRequest(userID)
		}
		return err
	case bot.BotCmd_Start:
		if !src.Chat.IsPrivateChat() {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You cannot "),
				styling.Code("/start"),
				styling.Plain(" this bot in group chat."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
			return nil
		}

		return c.handleStartCommand(logger, wf, src, params, msg)
	case bot.BotCmd_Cancel:
		prevReq, ok := c.sessions.ResolvePendingRequest(userID)
		if ok {
			// a pending request, no generator involved
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You have canceled the pending "),
				styling.Code(wf.BotCommands.TextOf(session.GetCommandFromRequest[chatIDWrapper](prevReq))),
				styling.Plain(" request"),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))

			if sr, isSR := prevReq.(*session.SessionRequest[chatIDWrapper]); isSR {
				if sr.Data.ID() != src.Chat.ID() {
					_, _ = c.sendTextMessage(
						c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent(),
						styling.Plain("Session canceled by the initiator."),
					)
				}
			}
		} else {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("There is no pending request."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
		}

		return nil
	case bot.BotCmd_End:
		currentSession, ok := c.sessions.GetActiveSession(chatID)
		if !ok {
			// TODO
			logger.D("invalid usage of end", log.String("reason", "no active session"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("There is no active session."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))

			return nil
		}

		msgs := currentSession.GetMessages()
		content, err := bot.GenerateContent(wf.Generator, msgs)
		if err != nil {
			logger.I("failed to generate post content", log.Error(err))
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: failed to generate post content: "),
				styling.Bold(err.Error()),
			)

			// do not execute again on telegram redelivery
			return nil
		}

		var (
			rd bytes.Reader
			in rt.Input
		)
		rd.Reset(content)
		in = rt.NewInput(rd.Size(), &rd)

		pub := currentSession.GetPublisher()
		note, err := pub.Append(c.Context(), &in)
		if err != nil {
			logger.I("failed to append content to post", log.Error(err))
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
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
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: active session already been ended out of no reason."),
			)

			return nil
		}

		_, _ = c.sendTextMessage(
			&c.sender.To(src.Chat.InputPeer()).Builder,
			translateSpans(note)...,
		)

		return nil
	case bot.BotCmd_Include:
		replyTo, ok := msg.GetReplyTo()
		if !ok {
			logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "not a reply"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Code(cmd),
				styling.Plain(" can only be used as a reply."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID)
			return nil
		}

		_, ok = c.sessions.GetActiveSession(chatID)
		if !ok {
			logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "not in a session"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Plain("There is no active session, "),
				styling.Code(cmd),
				styling.Plain(" will do nothing in this case."),
			)
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
			return nil
		}

		var (
			history tg.MessagesMessagesClass
			err     error
		)

		if c.isBot {
			if src.Chat.IsPrivateChat() || src.Chat.IsLegacyGroupChat() {
				history, err = c.client.API().MessagesGetMessages(c.Context(), []tg.InputMessageClass{
					&tg.InputMessageID{
						ID: replyTo.GetReplyToMsgID(),
					},
				})
			} else {
				peer := src.Chat.InputPeer().(*tg.InputPeerChannel)
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
				Peer:     src.Chat.InputPeer(),
				OffsetID: replyTo.GetReplyToMsgID(),
				Limit:    1,
			})
		}

		if err != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
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
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold(err.Error()),
			)

			return nil
		}

		err = c.appendSessionMessage(logger, src, toAppend)
		if err != nil {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Failed to include that message."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
			return err
		}

		msgID, _ := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Included."),
		)

		c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID)
		return nil
	case bot.BotCmd_Ignore:
		replyTo, ok := msg.GetReplyTo()
		if !ok {
			logger.D("invalid command usage", log.String("cmd", cmd), log.String("reason", "not a reply"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Code(cmd),
				styling.Plain(" can only be used as a reply."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
			return nil
		}

		currentSession, ok := c.sessions.GetActiveSession(chatID)
		if !ok {
			logger.D("invalid command usage", log.String("reason", "not in a session"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
				styling.Plain("There is not active session, "),
				styling.Code(cmd),
				styling.Plain(" will do nothing in this case."),
			)
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))

			return nil
		}

		_ = currentSession.DeleteMessage(rt.MessageID(replyTo.GetReplyToMsgID()))

		logger.V("ignored message")
		msgID, _ := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).Silent().Reply(msg.GetID()),
			styling.Plain("Ignored."),
		)

		c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID)

		return nil
	case bot.BotCmd_Edit:
		if !src.Chat.IsPrivateChat() {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You cannot use "),
				styling.Code(cmd),
				styling.Plain(" command in group chat."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
			return nil
		}

		prevCmd, ok := c.sessions.MarkPendingEditing(wf, userID, 5*time.Minute)
		if !ok {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
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
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()).
				Row(&tg.KeyboardButtonURL{
					Text: "Enter",
					URL:  urlForEdit,
				}),
			styling.Plain("Enter your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token to edit."),
		)

		return err
	case bot.BotCmd_Delete:
		if !src.Chat.IsPrivateChat() {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You cannot use "),
				styling.Code(cmd),
				styling.Plain(" command in group chat."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
			return nil
		}

		if len(params) == 0 {
			logger.D("invalid command usage", log.String("reason", "missing param"))
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).Reply(msg.GetID()),
				styling.Plain("Please specify the url(s) of the "),
				styling.Bold(wf.PublisherName()),
				styling.Plain(" post(s) to be deleted."),
			)
			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))

			return nil
		}

		prevCmd, ok := c.sessions.MarkPendingDeleting(wf, userID, strings.Split(params, " "), 5*time.Minute)
		if !ok {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
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
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()).
				Row(&tg.KeyboardButtonURL{
					Text: "Enter",
					URL:  urlForDelete,
				}),
			styling.Plain("Enter your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token to delete the post."),
		)

		return err
	case bot.BotCmd_List:
		if !src.Chat.IsPrivateChat() {
			msgID, _ := c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You cannot use "),
				styling.Code(cmd),
				styling.Plain(" command in group chat."),
			)

			c.scheduleMessageDelete(&src.Chat, 5*time.Second, msgID, rt.MessageID(msg.GetID()))
			return nil
		}

		prevCmd, ok := c.sessions.MarkPendingListing(wf, userID, 5*time.Minute)
		if !ok {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("You have pending "),
				styling.Code(wf.BotCommands.TextOf(prevCmd)),
				styling.Plain(" request not finished."),
			)
			return nil
		}

		// base64-url(list:hex(userID):hex(chatID))
		userIDPart := encodeUint64Hex(userID)
		chatIDPart := encodeUint64Hex(chatID)

		urlForList := fmt.Sprintf(
			"https://t.me/%s?start=%s",
			c.username,
			base64.URLEncoding.EncodeToString(
				[]byte(fmt.Sprintf("list:%s:%s", userIDPart, chatIDPart)),
			),
		)

		_, err := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()).
				Row(&tg.KeyboardButtonURL{
					Text: "Enter",
					URL:  urlForList,
				}),
			styling.Plain("Enter your "),
			styling.Bold(wf.PublisherName()),
			styling.Plain(" token to list your posts."),
		)
		if err != nil {
			_, _ = c.sendTextMessage(
				c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Bold(err.Error()),
			)

			return err
		}

		return nil
	case bot.BotCmd_Help:
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

		body = append(body, styling.Plain("\n"))

		_, _ = c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent(),
			body...,
		)
		return nil
	default:
		logger.D("unknown cmd")

		_, err := c.sendTextMessage(
			c.sender.To(src.Chat.InputPeer()).NoWebpage().Silent().Reply(msg.GetID()),
			styling.Plain("Command "),
			styling.Code(cmd),
			styling.Plain(" is not supported."),
		)

		return err
	}
}

func expectExactOneMessage(msgs []tg.MessageClass) (*tg.Message, error) {
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
