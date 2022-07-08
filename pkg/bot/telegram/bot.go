package telegram

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"
	"unicode/utf16"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/queue"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	api "arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/manager"
)

type msgDeleteKey struct{ chatID, msgID uint64 }

var _ bot.Interface = (*telegramBot)(nil)

type telegramBot struct {
	bot.BaseBot

	botToken    string
	botUsername string // set when Configure() called

	client api.Client

	manager.SessionManager

	wfSet         bot.WorkflowSet
	webhookConfig webhookConfig

	msgDelQ queue.TimeoutQueue[msgDeleteKey, struct{}]
}

// nolint:unparam
func (c *telegramBot) onTelegramUpdate(updates ...api.Update) (maxID int, _ error) {
	for _, update := range updates {
		switch {
		case update.Message != nil:
			err := c.handleNewMessage(update.Message)
			if err != nil {
				c.Logger().D("failed to handle new message", log.Error(err))
				continue
			}
		case update.EditedMessage != nil:
			// 	c.handleMessageEdit(update.EditedMessage)
		}

		if maxID < update.UpdateId {
			maxID = update.UpdateId
		}
	}

	return maxID, nil
}

// nolint:gocyclo
func (c *telegramBot) handleNewMessage(msg *api.Message) error {
	from := "<unknown>"
	if msg.From != nil && msg.From.Username != nil {
		from = *msg.From.Username
	}
	chat := "<unknown>"
	if msg.Chat.Username != nil {
		from = *msg.Chat.Username
	}

	logger := c.Logger().WithFields(log.String("from", from), log.String("chat", chat))

	switch {
	case msg.Text != nil:
		// check command
		content := utf16.Encode([]rune(*msg.Text))

		isCmd := false
		cmd := ""
		if msg.Entities != nil {
			for _, e := range *msg.Entities {
				// only check first command
				if e.Type == api.MessageEntityTypeBotCommand {
					isCmd = true
					cmd = string(utf16.Decode(content[e.Offset : e.Offset+e.Length]))
					content = content[e.Offset+e.Length:]

					cmd, _, _ = strings.Cut(cmd, "@")

					break
				}
			}
		}

		if isCmd {
			return c.handleCmd(
				logger,
				uint64(msg.Chat.Id),
				cmd,
				strings.TrimSpace(string(utf16.Decode(content))),
				msg,
			)
		}

		// filter private message for special replies to this bot
		if msg.Chat.Type == api.ChatTypePrivate && msg.From != nil && msg.ReplyToMessage != nil {
			var (
				chatID = uint64(msg.Chat.Id)
				userID = uint64(msg.From.Id)
			)

			for _, handle := range [...]tokenInputHandleFunc{
				c.tryToHandleInputForDiscussOrContinue,
				c.tryToHandleInputForEditing,
				c.tryToHandleInputForListing,
				c.tryToHandleInputForDeleting,
			} {
				handled, err := handle(chatID, userID, msg)
				if handled {
					return err
				}
			}
		}

		return c.appendSessionMessage(logger, uint64(msg.Chat.Id), msg)
	default:
		return c.appendSessionMessage(logger, uint64(msg.Chat.Id), msg)
	}
}

func (c *telegramBot) appendSessionMessage(
	logger log.Interface,
	chatID uint64,
	msg *api.Message,
) error {
	// ignore bot messages
	if msg.From != nil && msg.From.IsBot {
		return nil
	}

	s, ok := c.GetActiveSession(chatID)
	if !ok {
		return nil
	}

	logger.V("append session message")
	m := newTelegramMessage(msg, s.RefMessages())

	errCh, err := c.preProcess(s.Workflow(), &m)
	if err != nil {
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Internal bot error: failed to pre-process this message: %v", err),
		)
		return fmt.Errorf("failed to pre-process this message: %w", err)
	}

	if errCh != nil {
		logger.V("doing pre-processing")
		go func() {
			defer logger.V("pre-processing completed")

			for err := range errCh {
				// best effort, no error check
				_, _ = c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					fmt.Sprintf("Internal bot error: failed to pre-process this message: %v", err),
				)
			}
		}()
	}

	s.AppendMessage(&m)

	return nil
}

// handleCmd handle single command with all params as a single string
// nolint:gocyclo
func (c *telegramBot) handleCmd(
	logger log.Interface,
	chatID uint64,
	cmd, params string,
	msg *api.Message,
) error {
	logger = logger.WithFields(log.String("cmd", cmd), log.String("params", params))

	if msg.From == nil {
		// ignore
		msgID, err := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"Anonymous user not allowed",
		)

		c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
		return err
	}

	userID := uint64(msg.From.Id)
	isPrivateMessage := false

	// ensure only group admin can start session
	switch msg.Chat.Type {
	case api.ChatTypeChannel:
		_, err := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"This bot doesn't work in channels",
		)
		return err
	case api.ChatTypePrivate:
		// direct message to bot, no permission check
		isPrivateMessage = true
	case api.ChatTypeGroup, api.ChatTypeSupergroup:
		// ensure only admin can use this bot
		resp, err := c.client.PostGetChatAdministrators(c.Context(), api.PostGetChatAdministratorsJSONRequestBody{
			ChatId: chatID,
		})
		if err != nil {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Failed to check group administrators: %v", err),
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return err
		}

		admins, err := api.ParsePostGetChatAdministratorsResponse(resp)
		_ = resp.Body.Close()
		if err != nil {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Failed to parse group administrators: %v", err),
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return err
		}

		if admins.JSON200 == nil || !admins.JSON200.Ok {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("telegram: unable to check group administrators: %s", admins.JSONDefault.Description),
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return err
		}

		isAdmin := false
		for _, admin := range admins.JSON200.Result {
			if uint64(admin.(api.ChatMemberAdministrator).User.Id) == userID {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Only administrators can use this bot in group chat",
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return err
		}
	}

	wf, ok := c.wfSet.WorkflowFor(cmd)
	if !ok {
		msgID, _ := c.sendTextMessage(chatID, true, true, msg.MessageId, "unknown command")
		c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
		err := fmt.Errorf("unkonwn command %q", cmd)
		return err
	}

	switch bc := wf.BotCommands.Parse(cmd); bc {
	case bot.BotCmd_Discuss, bot.BotCmd_Continue:
		// mark this session as standby, wait for reply from bot private message
		var topic, url, onInvalidCmdMsg string

		switch bc {
		case bot.BotCmd_Discuss:
			topic = params
			onInvalidCmdMsg = fmt.Sprintf("Please specify a session topic, e.g. <code>%s foo</code>", cmd)
		case bot.BotCmd_Continue:
			url = params
			// nolint:lll
			onInvalidCmdMsg = fmt.Sprintf("Please specify the key of the session, e.g. <code>%s your-key</code>", cmd)
		}

		if len(params) == 0 {
			logger.D("invalid command usage", log.String("reason", "missing param"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				onInvalidCmdMsg,
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		_, ok := c.GetActiveSession(chatID)
		if ok {
			logger.D("invalid command usage", log.String("reason", "already in a session"))
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Please end current session before starting a new one",
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		if !c.MarkSessionStandby(wf, userID, chatID, topic, url, 5*time.Minute) {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"You have already started a session with no token replied, please end that first",
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		if !wf.PublisherRequireLogin() {
			pub, _, _ := wf.CreatePublisher()

			_, err := c.ActivateSession(wf, chatID, userID, pub)
			if err != nil {
				msgID, _ := c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					"You have already started a session before, please end that first",
				)
				c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

				return nil
			}

			defer func() {
				if err != nil {
					c.DeactivateSession(chatID)
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
				_, _ = c.sendTextMessage(chatID, true, true, 0, renderEntities(note))
			}

			return nil
		}

		return func() (err error) {
			defer func() {
				if err != nil {
					c.ResolvePendingRequest(userID)
				}
			}()

			// base64-url({create | enter}:hex(userID):hex(chatID))
			userIDPart := encodeUint64Hex(userID)
			chatIDPart := encodeUint64Hex(chatID)

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
				inlineKeyboard [1][]api.InlineKeyboardButton
				textPrompt     = ""
			)

			switch bc {
			case bot.BotCmd_Discuss:
				textPrompt = "Create or enter your %s token for this session"
				inlineKeyboard[0] = append(inlineKeyboard[0], api.InlineKeyboardButton{
					Text: "Create",
					Url:  &urlForCreate,
				})
			case bot.BotCmd_Continue:
				textPrompt = "Enter your %s token to continue this session"
			}

			inlineKeyboard[0] = append(inlineKeyboard[0], api.InlineKeyboardButton{
				Text: "Enter",
				Url:  &urlForEnter,
			})

			_, err = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf(textPrompt, wf.PublisherName()),
				api.InlineKeyboardMarkup{
					InlineKeyboard: inlineKeyboard[:],
				},
			)
			return err
		}()
	case bot.BotCmd_Start:
		return c.handleStartCommand(wf, logger, chatID, userID, isPrivateMessage, params, msg)
	case bot.BotCmd_Cancel:
		prevReq, ok := c.ResolvePendingRequest(userID)
		if ok {
			// a pending request, no generator involved
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf(
					"You have canceled the pending <code>%s</code> request",
					manager.GetCommandFromRequest(prevReq),
				),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			if sr, isSR := prevReq.(*manager.SessionRequest); isSR {
				if sr.ChatID != chatID {
					_, _ = c.sendTextMessage(
						sr.ChatID, true, true, 0,
						"Session canceled by the initiator.",
					)
				}
			}
		} else {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"There is no pending request",
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
		}

		return nil
	case bot.BotCmd_End:
		currentSession, ok := c.GetActiveSession(chatID)
		if !ok {
			// TODO
			logger.D("invalid usage of end", log.String("reason", "no active session"))
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"There is no active session",
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		n, content, err := currentSession.GenerateContent(wf.Generator)
		if err != nil {
			logger.I("failed to generate post content", log.Error(err))
			_, _ = c.sendTextMessage(
				chatID, false, true, msg.MessageId,
				fmt.Sprintf("Internal bot error: failed to generate post content: %v", err),
			)

			// do not execute again on telegram redelivery
			return nil
		}

		pub := currentSession.GetPublisher()
		note, err := pub.Append(c.Context(), content)
		if err != nil {
			logger.I("failed to append content to post", log.Error(err))
			_, _ = c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf("%s post update error: %v", pub.Name(), err),
			)

			// do not execute again on telegram redelivery
			return nil
		}

		currentSession.DeleteFirstNMessage(n)

		_, ok = c.DeactivateSession(chatID)
		if !ok {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Internal bot error: active session already been ended out of no reason",
			)

			return nil
		}

		_, _ = c.sendTextMessage(chatID, false, false, 0, renderEntities(note))

		return nil
	case bot.BotCmd_Include:
		replyTo := msg.ReplyToMessage
		if replyTo == nil {
			logger.D("invalid command usage", log.String("reason", "not a reply"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf("<code>%s</code> can only be used as a reply", cmd),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID))
			return nil
		}

		_, ok := c.GetActiveSession(chatID)
		if !ok {
			logger.D("invalid command usage", log.String("reason", "not in a session"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf("There is no active session, <code>%s</code> will do nothing in this case", cmd),
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return nil
		}

		err := c.appendSessionMessage(logger, chatID, replyTo)
		if err != nil {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Failed to include that message",
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return err
		}

		msgID, _ := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"The message you replied just got included",
		)

		c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID))

		return nil
	case bot.BotCmd_Ignore:
		replyTo := msg.ReplyToMessage
		if replyTo == nil {
			logger.D("invalid command usage", log.String("reason", "not a reply"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf("<code>%s</code> can only be used as a reply", cmd),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return nil
		}

		currentSession, ok := c.GetActiveSession(chatID)
		if !ok {
			logger.D("invalid command usage", log.String("reason", "not in a session"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf("There is not active session, <code>%s</code> will do nothing in this case", cmd),
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		_ = currentSession.DeleteMessage(formatMessageID(replyTo.MessageId))

		logger.V("ignored message")
		msgID, _ := c.sendTextMessage(
			chatID, false, false, msg.MessageId,
			"The message you replied just got ignored",
		)

		c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID))

		return nil
	case bot.BotCmd_Edit:
		if !isPrivateMessage {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You cannot use <code>%s</code> command in groups", cmd),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return nil
		}

		prevCmd, ok := c.MarkPendingEditing(wf, userID, 5*time.Minute)
		if !ok {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You have pending <code>%s</code> request not finished", prevCmd),
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
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Enter your %s token to edit", wf.PublisherName()),
			api.InlineKeyboardMarkup{
				InlineKeyboard: [][]api.InlineKeyboardButton{{{
					Text: "Enter",
					Url:  &urlForEdit,
				}}},
			},
		)

		return err
	case bot.BotCmd_Delete:
		if !isPrivateMessage {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You cannot use <code>%s</code> command in groups", cmd),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return nil
		}

		if len(params) == 0 {
			logger.D("invalid command usage", log.String("reason", "missing param"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf("Please specify the url(s) of the %s post(s) to be deleted", wf.PublisherName()),
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		prevCmd, ok := c.MarkPendingDeleting(wf, userID, strings.Split(params, " "), 5*time.Minute)
		if !ok {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You have pending <code>%s</code> request not finished", prevCmd),
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
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Enter your %s token to delete the post", wf.PublisherName()),
			api.InlineKeyboardMarkup{
				InlineKeyboard: [][]api.InlineKeyboardButton{{{
					Text: "Enter",
					Url:  &urlForDelete,
				}}},
			},
		)

		return err
	case bot.BotCmd_List:
		if !isPrivateMessage {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You cannot use <code>%s</code> command in groups", cmd),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return nil
		}

		prevCmd, ok := c.MarkPendingListing(wf, userID, 5*time.Minute)
		if !ok {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You have pending <code>%s</code> request not finished", prevCmd),
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
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Enter your %s token to list your posts", wf.PublisherName()),
			api.InlineKeyboardMarkup{
				InlineKeyboard: [][]api.InlineKeyboardButton{{{
					Text: "Enter",
					Url:  &urlForList,
				}}},
			},
		)

		return err
	case bot.BotCmd_Help:
		var body strings.Builder

		body.WriteString("Usage:\n\n")
		for i, cmd := range wf.BotCommands.Commands {
			if len(cmd) == 0 || len(wf.BotCommands.Descriptions[i]) == 0 {
				continue
			}

			body.WriteString("<pre>")
			body.WriteString(cmd)
			body.WriteString("</pre> - ")
			body.WriteString(wf.BotCommands.Descriptions[i])
			body.WriteString("\n")
		}

		body.WriteString("\n")

		_, _ = c.sendTextMessage(chatID, true, true, 0, body.String())
		return nil
	default:
		logger.D("unknown command")

		_, err := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Command <code>%s</code> is not supported", cmd),
		)

		return err
	}
}
