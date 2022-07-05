package telegram

import (
	"context"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
	"unicode/utf16"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/queue"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/conf"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/manager"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

type msgDeleteKey struct {
	chatID    uint64
	messageID uint64
}

var _ bot.Interface = (*telegramBot)(nil)

type telegramBot struct {
	ctx    context.Context
	logger log.Interface
	client *telegram.Client

	botUsername string

	*manager.SessionManager

	oldToNew map[string]conf.BotCommandMappingConfig
	newToOld map[string]string

	storage     storage.Interface
	webArchiver webarchiver.Interface
	generator   generator.Interface

	publisherName         string
	publisherRequireLogin bool
	createPublisher       bot.PublisherFactoryFunc

	msgTpl *template.Template

	opts *conf.TelegramConfig

	msgDelQ *queue.TimeoutQueue[msgDeleteKey, struct{}]
}

func Create(
	ctx context.Context,
	logger log.Interface,
	storage storage.Interface,
	webArchiver webarchiver.Interface,
	generator generator.Interface,
	createPublisher bot.PublisherFactoryFunc,
	oldToNew map[string]conf.BotCommandMappingConfig,
	newToOld map[string]string,
	opts *conf.TelegramConfig,
) (bot.Interface, error) {
	tgClient, err := telegram.NewClient(
		fmt.Sprintf("https://%s/bot%s/", opts.Endpoint, opts.BotToken),
		telegram.WithHTTPClient(&http.Client{}),
		telegram.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bot client: %w", err)
	}

	pub, _, err := createPublisher()
	if err != nil {
		return nil, fmt.Errorf("failed to test publisher creation: %w", err)
	}

	msgTpl, err := template.New("").Parse(messageTemplate)
	if err != nil {
		return nil, fmt.Errorf("invalid message template: %w", err)
	}

	client := &telegramBot{
		ctx:    ctx,
		logger: logger,
		client: tgClient,

		opts: opts,

		botUsername: "", // set in Configure()

		SessionManager: manager.NewSessionManager(ctx),

		oldToNew: oldToNew,
		newToOld: newToOld,

		webArchiver: webArchiver,
		storage:     storage,
		generator:   generator,

		publisherName:         pub.Name(),
		publisherRequireLogin: pub.RequireLogin(),
		createPublisher:       createPublisher,

		msgTpl: msgTpl,

		msgDelQ: queue.NewTimeoutQueue[msgDeleteKey, struct{}](),
	}

	return client, nil
}

// nolint:unparam
func (c *telegramBot) onTelegramUpdate(updates ...telegram.Update) (maxID int, _ error) {
	for _, update := range updates {
		switch {
		case update.Message != nil:
			err := c.handleNewMessage(update.Message)
			if err != nil {
				c.logger.D("failed to handle new message", log.Error(err))
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
func (c *telegramBot) handleNewMessage(msg *telegram.Message) error {
	from := "<unknown>"
	if msg.From != nil && msg.From.Username != nil {
		from = *msg.From.Username
	}
	chat := "<unknown>"
	if msg.Chat.Username != nil {
		from = *msg.Chat.Username
	}

	logger := c.logger.WithFields(log.String("from", from), log.String("chat", chat))

	switch {
	case msg.Text != nil:
		// check command
		content := utf16.Encode([]rune(*msg.Text))

		isCmd := false
		cmd := ""
		if msg.Entities != nil {
			for _, e := range *msg.Entities {
				// only check first command
				if e.Type == telegram.MessageEntityTypeBotCommand {
					isCmd = true
					cmd = string(utf16.Decode(content[e.Offset : e.Offset+e.Length]))
					content = content[e.Offset+e.Length:]

					parts := strings.SplitN(cmd, "@", 2)
					cmd = parts[0]

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
		if msg.Chat.Type == telegram.ChatTypePrivate && msg.From != nil && msg.ReplyToMessage != nil {
			userID := uint64(msg.From.Id)
			chatID := uint64(msg.Chat.Id)

			for _, handle := range []tokenInputHandleFunc{
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
	msg *telegram.Message,
) error {
	currentSession, ok := c.GetActiveSession(chatID)
	if !ok {
		return nil
	}

	// ignore bot messages
	if msg.From != nil && msg.From.IsBot {
		return nil
	}

	logger.V("appending session meesage")
	m := newTelegramMessage(msg, currentSession.RefMessages())

	errCh, err := c.preProcess(m, c.webArchiver, c.storage)
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

	currentSession.AppendMessage(m)

	return nil
}

// nolint:gocyclo
func (c *telegramBot) handleCmd(
	logger log.Interface,
	chatID uint64,
	cmd, params string,
	msg *telegram.Message,
) error {
	logger = logger.WithFields(log.String("cmd", cmd), log.String("params", params))

	if msg.From == nil {
		// ignore
		msgID, err := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"Sorry, anonymous users are not allowed",
		)
		c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
		return err
	}

	userID := uint64(msg.From.Id)
	isPrivateMessage := false

	// ensure only group admin can start session
	switch msg.Chat.Type {
	case telegram.ChatTypeChannel:
		_, err := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"This bot doesn't work in channels",
		)
		return err
	case telegram.ChatTypePrivate:
		// direct message to bot, no permission check
		isPrivateMessage = true
	case telegram.ChatTypeGroup, telegram.ChatTypeSupergroup:
		// ensure only admin can use this bot
		resp, err := c.client.PostGetChatAdministrators(c.ctx, telegram.PostGetChatAdministratorsJSONRequestBody{
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

		admins, err := telegram.ParsePostGetChatAdministratorsResponse(resp)
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
			if uint64(admin.(telegram.ChatMemberAdministrator).User.Id) == userID {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Sorry, only administrators can use this bot in group chat",
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return err
		}
	}

	switch c.newToOld[cmd] {
	case constant.CommandDiscuss, constant.CommandContinue:
		// mark this session as standby, wait for reply from bot private message
		var topic, url, onInvalidCmdMsg string
		switch c.newToOld[cmd] {
		case constant.CommandDiscuss:
			topic = params
			onInvalidCmdMsg = fmt.Sprintf("Please specify a session topic, e.g. <code>%s foo</code>", cmd)
		case constant.CommandContinue:
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

		if !c.MarkSessionStandby(userID, chatID, topic, url, 5*time.Minute) {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"You have already started a session with no token replied, please end that first",
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		if !c.publisherRequireLogin {
			pub, _, _ := c.createPublisher()

			_, err := c.ActivateSession(chatID, userID, pub)
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

			pageHeader, err := c.generator.RenderPageHeader()
			if err != nil {
				return fmt.Errorf("failed to render page header: %w", err)
			}

			note, err := pub.Publish(topic, pageHeader)
			if err != nil {
				return fmt.Errorf("failed to pre-publish page: %w", err)
			}

			if len(note) != 0 {
				_, _ = c.sendTextMessage(chatID, true, true, 0, c.renderEntities(note))
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
				inlineKeyboard [1][]telegram.InlineKeyboardButton
				textPrompt     = ""
			)

			switch c.newToOld[cmd] {
			case constant.CommandDiscuss:
				textPrompt = "Create or enter your %s token for this session"
				inlineKeyboard[0] = append(inlineKeyboard[0], telegram.InlineKeyboardButton{
					Text: "Create",
					Url:  &urlForCreate,
				})
			case constant.CommandContinue:
				textPrompt = "Enter your %s token to continue this session"
			}

			inlineKeyboard[0] = append(inlineKeyboard[0], telegram.InlineKeyboardButton{
				Text: "Enter",
				Url:  &urlForEnter,
			})

			_, err = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf(textPrompt, c.publisherName),
				telegram.InlineKeyboardMarkup{
					InlineKeyboard: inlineKeyboard[:],
				},
			)
			return err
		}()
	case constant.CommandStart:
		return c.handleStartCommand(logger, chatID, userID, isPrivateMessage, params, msg)
	case constant.CommandCancel:
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
						"Session was canceled by the initiator.",
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
	case constant.CommandEnd:
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

		n, content, err := currentSession.GenerateContent(c.generator)
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
		note, err := pub.Append(c.ctx, content)
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

		_, _ = c.sendTextMessage(chatID, false, false, 0, c.renderEntities(note))

		return nil
	case constant.CommandInclude:
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
	case constant.CommandIgnore:
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
	case constant.CommandEdit:
		if !isPrivateMessage {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You cannot use <code>%s</code> command in groups", cmd),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return nil
		}

		prevCmd, ok := c.MarkPendingEditing(userID, 5*time.Minute)
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
			fmt.Sprintf("Enter your %s token to edit", c.publisherName),
			telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{{{
					Text: "Enter",
					Url:  &urlForEdit,
				}}},
			},
		)

		return err
	case constant.CommandDelete:
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
				fmt.Sprintf("Please specify the url(s) of the %s post(s) to be deleted", c.publisherName),
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		prevCmd, ok := c.MarkPendingDeleting(userID, strings.Split(params, " "), 5*time.Minute)
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
			fmt.Sprintf("Enter your %s token to delete the post", c.publisherName),
			telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{{{
					Text: "Enter",
					Url:  &urlForDelete,
				}}},
			},
		)

		return err
	case constant.CommandList:
		if !isPrivateMessage {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You cannot use <code>%s</code> command in groups", cmd),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return nil
		}

		prevCmd, ok := c.MarkPendingListing(userID, 5*time.Minute)
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
			fmt.Sprintf("Enter your %s token to list your posts", c.publisherName),
			telegram.InlineKeyboardMarkup{
				InlineKeyboard: [][]telegram.InlineKeyboardButton{{{
					Text: "Enter",
					Url:  &urlForList,
				}}},
			},
		)

		return err
	case constant.CommandHelp:
		body := ""

		for _, cmd := range constant.VisibleBotCommands {
			spec, ok := c.oldToNew[cmd]
			if !ok {
				continue
			}

			body += "<pre>" + spec.As + "</pre> - " + spec.Description + "\n"
		}

		_, _ = c.sendTextMessage(chatID, true, true, 0, fmt.Sprintf("Usage:\n\n%s\n", body))
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
