package telegram

import (
	"context"
	"encoding/base64"
	"fmt"
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

	storage     storage.Interface
	webArchiver webarchiver.Interface
	generator   generator.Interface

	publisherName         string
	publisherRequireLogin bool
	createPublisher       bot.PublisherFactoryFunc

	opts *conf.TelegramConfig

	msgDelQ *queue.TimeoutQueue
}

func Create(
	ctx context.Context,
	logger log.Interface,
	storage storage.Interface,
	webArchiver webarchiver.Interface,
	generator generator.Interface,
	createPublisher bot.PublisherFactoryFunc,
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

	client := &telegramBot{
		ctx:    ctx,
		logger: logger,
		client: tgClient,

		opts: opts,

		botUsername: "", // set in Configure()

		SessionManager: manager.NewSessionManager(ctx),

		webArchiver: webArchiver,
		storage:     storage,
		generator:   generator,

		publisherName:         pub.Name(),
		publisherRequireLogin: pub.RequireLogin(),
		createPublisher:       createPublisher,

		msgDelQ: queue.NewTimeoutQueue(),
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

	defaultChatUsername := ""
	// ensure only group admin can start discussion
	switch msg.Chat.Type {
	case telegram.ChatTypeChannel:
		_, err := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"This bot doesn't work in channels",
		)
		return err
	case telegram.ChatTypePrivate:
		// direct message to bot, no permission check
		defaultChatUsername = c.botUsername
		isPrivateMessage = true
	case telegram.ChatTypeGroup, telegram.ChatTypeSupergroup:
		if msg.Chat.Username != nil {
			defaultChatUsername = *msg.Chat.Username
		} else {
			resp, err := c.client.PostGetChat(
				c.ctx, telegram.PostGetChatJSONRequestBody{
					ChatId: chatID,
				},
			)
			if err != nil {
				_, _ = c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					fmt.Sprintf("Unable to check group chat info: %v", err),
				)
				return err
			}

			chat, err := telegram.ParsePostGetChatResponse(resp)
			_ = resp.Body.Close()
			if err != nil {
				_, _ = c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					fmt.Sprintf("Unable to parse group chat info: %v", err),
				)
				return err
			}

			if chat.JSON200 == nil || !chat.JSON200.Ok {
				_, _ = c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					fmt.Sprintf("Telegram: unable to check group administrators: %s", chat.JSONDefault.Description),
				)
				return err
			}

			if usernamePtr := chat.JSON200.Result.Username; usernamePtr != nil {
				defaultChatUsername = *usernamePtr
			}
		}

		resp, err := c.client.PostGetChatAdministrators(c.ctx, telegram.PostGetChatAdministratorsJSONRequestBody{
			ChatId: chatID,
		})
		if err != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Unable to check group administrators: %v", err),
			)
			return err
		}

		admins, err := telegram.ParsePostGetChatAdministratorsResponse(resp)
		_ = resp.Body.Close()
		if err != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Unable to parse group administrators: %v", err),
			)
			return err
		}

		if admins.JSON200 == nil || !admins.JSON200.Ok {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Telegram: unable to check group administrators: %s", admins.JSONDefault.Description),
			)
			return err
		}

		isAdmin := false
		for _, admin := range admins.JSON200.Result {
			if uint64(admin.User.Id) == userID {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Sorry, only administrators can use this bot in group chat",
			)
			return err
		}
	}

	switch cmd {
	case constant.CommandDiscuss, constant.CommandContinue:
		// mark this session as standby, wait for reply from bot private message
		var topic, url, onInvalidCmdMsg string
		switch cmd {
		case constant.CommandDiscuss:
			topic = params
			onInvalidCmdMsg = "Please specify a discussion topic, e.g. <code>%s foo</code>"
		case constant.CommandContinue:
			url = params
			// nolint:lll
			onInvalidCmdMsg = "Please specify the url of the discussion post, e.g. <code>%s https://telegra.ph/foo-01-21-100</code>"
		}

		if len(params) == 0 {
			logger.D("invalid command usage", log.String("reason", "missing param"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf(onInvalidCmdMsg, cmd),
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		_, ok := c.GetActiveSession(chatID)
		if ok {
			logger.D("invalid command usage", log.String("reason", "already in a session"))
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Please <code>/end</code> current discussion before starting a new one",
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		if !c.publisherRequireLogin {
			// TODO
		}

		if !c.MarkSessionStandby(userID, chatID, defaultChatUsername, topic, url, 5*time.Minute) {
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"You have already started a discussion with no auth token specified, please end that first",
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

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

			switch cmd {
			case constant.CommandDiscuss:
				textPrompt = "Create or enter your %s token for this discussion"
				inlineKeyboard[0] = append(inlineKeyboard[0], telegram.InlineKeyboardButton{
					Text: "Create",
					Url:  &urlForCreate,
				})
			case constant.CommandContinue:
				textPrompt = "Enter your %s token to continue this discussion"
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
	case constant.CommandEnd:
		prevReq, ok := c.ResolvePendingRequest(userID)
		if ok {
			// a pending request, no generator involved
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You have canceled the pending <code>%s</code>", manager.GetCommandFromRequest(prevReq)),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			if sr, isSR := prevReq.(*manager.SessionRequest); isSR {
				if sr.ChatID != chatID {
					_, _ = c.sendTextMessage(
						sr.ChatID, true, true, 0,
						"Discussion was canceled by the initiator.",
					)
				}
			}

			return nil
		}

		currentSession, ok := c.GetActiveSession(chatID)
		if !ok {
			// TODO
			logger.D("invalid usage of end", log.String("reason", "no active session"))
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"There is no active discussion",
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
		postURL, err := pub.Append(currentSession.GetTopic(), content)
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

		currentSession, ok = c.DeactivateSession(chatID)
		if !ok {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Internal bot error: active discussion already been ended out of no reason",
			)

			return nil
		}

		_, _ = c.sendTextMessage(
			chatID, false, false, 0,
			fmt.Sprintf(
				"Your discussion around %q has been ended, view and edit your post: %s",
				currentSession.GetTopic(), postURL,
			),
		)

		return nil
	case constant.CommandInclude:
		replyTo := msg.ReplyToMessage
		if replyTo == nil {
			logger.D("invalid command usage", log.String("reason", "not a reply"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				"<code>/include</code> can only be used as a reply",
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID))
			return nil
		}

		_, ok := c.GetActiveSession(chatID)
		if !ok {
			logger.D("invalid command usage", log.String("reason", "not in a session"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				"There is not active discussion, <code>/include</code> will do nothing in this case",
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
				"<code>/ignore</code> can only be used as a reply",
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
			return nil
		}

		currentSession, ok := c.GetActiveSession(chatID)
		if !ok {
			logger.D("invalid command usage", log.String("reason", "not in a session"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				"There is not active discussion, <code>/ignore</code> will do nothing in this case",
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
				"You cannot use <code>/edit</code> command in groups",
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
				"You cannot use <code>/delete</code> command in groups",
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
				"You cannot use <code>/list</code> command in groups",
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
		_, _ = c.sendTextMessage(chatID, true, true, 0, constant.CommandHelpText())
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

func (c telegramBot) scheduleMessageDelete(chatID uint64, after time.Duration, msgIDs ...uint64) {
	for _, msgID := range msgIDs {
		if msgID == 0 {
			// ignore invalid message id
			continue
		}

		_ = c.msgDelQ.OfferWithDelay(msgDeleteKey{
			chatID:    chatID,
			messageID: msgID,
		}, struct{}{}, after)
	}
}

func (c *telegramBot) sendTextMessage(
	chatID uint64,
	disableNotification,
	disableWebPreview bool,
	replyTo int,
	text string,
	replyMarkup ...interface{},
) (int, error) {
	var replyToMsgIDPtr *int
	if replyTo > 0 {
		replyToMsgIDPtr = &replyTo
	}

	var replyMarkupPtr *interface{}
	if len(replyMarkup) > 0 {
		replyMarkupPtr = &replyMarkup[0]
	}

	var htmlStyle = "HTML"
	resp, err := c.client.PostSendMessage(
		c.ctx,
		telegram.PostSendMessageJSONRequestBody{
			AllowSendingWithoutReply: constant.True(),
			ChatId:                   chatID,
			DisableNotification:      &disableNotification,
			DisableWebPagePreview:    &disableWebPreview,
			ReplyToMessageId:         replyToMsgIDPtr,
			ParseMode:                &htmlStyle,
			Text:                     text,
			ReplyMarkup:              replyMarkupPtr,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}

	result, err := telegram.ParsePostSendMessageResponse(resp)
	_ = resp.Body.Close()
	if err != nil {
		return 0, fmt.Errorf("failed to parse response of message send: %w", err)
	}

	if result.JSON200 == nil || !result.JSON200.Ok {
		return 0, fmt.Errorf("telegram: failed to send message: %s", result.JSONDefault.Description)
	}

	return result.JSON200.Result.MessageId, nil
}

func (c *telegramBot) handleStartCommand(
	logger log.Interface,
	chatID uint64,
	userID uint64,
	isPrivateMessage bool,
	params string,
	msg *telegram.Message,
) error {
	if !isPrivateMessage {
		msgID, _ := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"You cannot <code>/start</code> this bot in groups",
		)

		c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))
		return nil
	}

	if len(params) == 0 {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "Welcome, need some <code>/help</code> ?")
		return nil
	}

	createOrEnter, err := base64.URLEncoding.DecodeString(params)
	if err != nil {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "I am alive.")
		return nil
	}

	parts := strings.SplitN(string(createOrEnter), ":", 3)
	if len(parts) != 3 {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "Told you, I'm alive.")
		return nil
	}

	action := parts[0]

	originalUserID, err := decodeUint64Hex(parts[1])
	if err != nil {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, fmt.Sprintf("Internal bot error: %s", err))
		return nil
	}

	// ensure same user
	if originalUserID != userID {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "The link is not for you :(")
		return nil
	}

	originalChatID, err := decodeUint64Hex(parts[2])
	if err != nil {
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, fmt.Sprintf("Internal bot error: %s", err))
		return nil
	}

	var (
		standbySession         *manager.SessionRequest
		expectedOriginalChatID uint64
	)

	switch action {
	case "create", "enter":
		var ok bool
		standbySession, ok = c.GetStandbySession(userID)
		if !ok {
			_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "No discussion requested")
			return nil
		}

		expectedOriginalChatID = standbySession.ChatID
	case "edit", "delete", "list":
		expectedOriginalChatID = chatID
	default:
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "Unknown action")
		return nil
	}

	// delete `/start` message
	c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msg.MessageId))

	if expectedOriginalChatID != originalChatID {
		// should not happen, defensive check
		_, _ = c.sendTextMessage(chatID, true, true, msg.MessageId, "Unexpected chat id not match")
		return nil
	}

	switch action {
	case "create":
		pub, userConfig, err2 := c.createPublisher()
		defer func() {
			if err2 != nil {
				_, _ = c.ResolvePendingRequest(userID)

				// best effort
				_, _ = c.sendTextMessage(
					chatID, true, true, 0,
					fmt.Sprintf("The discussion was canceled due to error, please retry later: %v", err2),
				)

				if standbySession.ChatID != chatID {
					_, _ = c.sendTextMessage(
						standbySession.ChatID, true, true, 0,
						"The discussion was canceled due to error, please retry later",
					)
				}
			}
		}()

		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Internal bot error: %v", err2),
			)
			return err2
		}

		token, err2 := pub.Login(userConfig)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("%s login failed: %v", pub.Name(), err2),
			)
			return err2
		}

		_, err2 = c.sendTextMessage(
			chatID, false, true, 0,
			fmt.Sprintf(
				"Here is your %s token, keep it on your own for later use:\n\n<pre>%s</pre>",
				pub.Name(), token,
			),
		)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, false, true, msg.MessageId,
				fmt.Sprintf("Internal bot error: unable to send %s token: %v", pub.Name(), err2),
			)
			return err2
		}

		content, err2 := c.generator.FormatPageHeader()
		if err2 != nil {
			return fmt.Errorf("failed to generate initial page: %w", err2)
		}

		postURL, err2 := pub.Publish(standbySession.Topic, content)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("%s pre-publish failed: %v", pub.Name(), err2),
			)
			return err2
		}

		currentSession, err2 := c.ActivateSession(
			standbySession.ChatID,
			userID,
			standbySession.Topic,
			standbySession.ChatUsername,
			pub,
		)
		if err2 != nil {
			logger.D("invalid usage of discuss", log.String("reason", err2.Error()))
			_, _ = c.sendTextMessage(
				standbySession.ChatID, true, true, 0,
				fmt.Sprintf("Could not activate discussion: %v", err2),
			)
			return err2
		}

		defer func() {
			if err2 != nil {
				// bset effort
				_, _ = c.DeactivateSession(standbySession.ChatID)
			}
		}()

		_, err2 = c.sendTextMessage(
			standbySession.ChatID, true, true, 0,
			fmt.Sprintf(
				"The post for your discussion around %q has been created: %s",
				currentSession.GetTopic(), postURL,
			),
		)

		return nil
	case "enter":
		msgID, err2 := c.sendTextMessage(chatID, false, true, 0,
			fmt.Sprintf("Enter your %s token as a reply to this message", c.publisherName),
			telegram.ForceReply{
				ForceReply: true,
				Selective:  constant.True(),
			},
		)
		if err2 != nil {
			// this message must be sent to user, this error will trigger message redelivery
			return err2
		}

		if !c.MarkRequestExpectingInput(userID, uint64(msgID)) {
			msgID2, _ := c.sendTextMessage(
				chatID, false, true, msg.MessageId,
				"The discussion is not expecting any input",
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msgID2))

			return nil
		}

		return nil
	case "edit", "delete", "list":
		msgID, err2 := c.sendTextMessage(
			chatID, true, true, 0,
			fmt.Sprintf("Enter your %s token as a reply to this message", c.publisherName),
			telegram.ForceReply{
				ForceReply: true,
				Selective:  constant.True(),
			},
		)

		if !c.MarkRequestExpectingInput(userID, uint64(msgID)) {
			_, _ = c.sendTextMessage(
				chatID, true, true, 0,
				"Internal bot error: could not find your pending request",
			)

			c.scheduleMessageDelete(chatID, 100*time.Millisecond, uint64(msgID))

			return nil
		}

		// this message must be sent to user, when the error is not nil
		// telegram will redeliver message
		return err2
	default:
		_, _ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"Nice try!",
		)

		return nil
	}
}
