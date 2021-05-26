package server

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/queue"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/conf"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

type msgDeleteKey struct {
	ChatID    uint64
	MessageID uint64
}

var _ Client = (*telegramBot)(nil)

type telegramBot struct {
	ctx    context.Context
	logger log.Interface
	client *telegram.Client

	botToken    string
	botUsername string

	// chart_id -> session
	*SessionManager

	storage         storage.Interface
	webArchiver     webarchiver.Interface
	generatorName   string
	createGenerator generatorFactoryFunc

	msgDelQ *queue.TimeoutQueue
}

// nolint:gocyclo
func createTelegramBot(
	ctx context.Context,
	logger log.Interface,
	baseURL string,
	mux *http.ServeMux,
	st storage.Interface,
	wa webarchiver.Interface,
	generatorName string,
	createGenerator generatorFactoryFunc,
	opts *conf.TelegramConfig,
) (*telegramBot, error) {
	tgClient, err := telegram.NewClient(
		fmt.Sprintf("https://%s/bot%s/", opts.Endpoint, opts.BotToken),
		telegram.WithHTTPClient(&http.Client{}),
		telegram.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			return nil
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot client: %w", err)
	}

	// get bot username
	resp, err2 := tgClient.PostGetMe(ctx)
	if err2 != nil {
		return nil, fmt.Errorf("failed to get my own info: %w", err2)
	}

	mr, err2 := telegram.ParsePostGetMeResponse(resp)
	_ = resp.Body.Close()
	if err2 != nil {
		return nil, fmt.Errorf("failed to parse bot get info response: %w", err2)
	}

	if mr.JSON200 == nil || !mr.JSON200.Ok {
		return nil, fmt.Errorf("failed to get telegram bot info: %s", mr.JSONDefault.Description)
	}

	if mr.JSON200.Result.Username == nil {
		return nil, fmt.Errorf("bot username not returned by telegram server")
	}

	botUsername := *mr.JSON200.Result.Username
	logger.D("got bot username", log.String("username", botUsername))

	allowedUpdates := []string{
		"message",
		// TODO: support message edit, how will we show edited messages?
		// "edited_message",
	}

	client := &telegramBot{
		ctx:    ctx,
		logger: logger,
		client: tgClient,

		botToken:    opts.BotToken,
		botUsername: botUsername,

		SessionManager: newSessionManager(ctx),

		webArchiver:     wa,
		storage:         st,
		generatorName:   generatorName,
		createGenerator: createGenerator,

		msgDelQ: queue.NewTimeoutQueue(),
	}

	if opts.Webhook.Enabled && len(baseURL) != 0 {
		// set webhook
		base, err2 := url.Parse(baseURL)
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse base url for telegram bot webhook: %w", err2)
		}

		port := ""
		p := base.Port()

		switch {
		case len(p) != 0:
			switch p {
			case "443":
				// omit 443 since we are uisng https
			case "80", "88", "8443":
				port = ":" + p
			default:
				return nil, fmt.Errorf("invalid port for telegram bot: port %s not allowed", port)
			}
		case len(p) == 0 && base.Scheme != "https":
			// not using https and no port set, lets guess!
			switch base.Scheme {
			case "http":
				port = ":80" // https over 80 is allowed by telegram
			default:
				return nil, fmt.Errorf("invalid base url for telegram bot: unable to find port")
			}
		}

		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		err2 = mw.WriteField(
			"url",
			fmt.Sprintf(
				"https://%s%s%s",
				base.Host,
				port,
				path.Join(base.RawPath, opts.Webhook.Path),
			),
		)
		if err2 != nil {
			return nil, fmt.Errorf("failed to set url: %w", err2)
		}

		var certBytes []byte
		switch {
		case len(opts.Webhook.TLSPublicKeyData) != 0:
			// use embedded data
			certBytes, err2 = base64.StdEncoding.DecodeString(opts.Webhook.TLSPublicKeyData)
			if err2 != nil {
				return nil, fmt.Errorf("failed to decode base64 encoded cert: %w", err2)
			}
		case len(opts.Webhook.TLSPublicKey) != 0 && len(opts.Webhook.TLSPublicKeyData) == 0:
			// read from file
			certBytes, err2 = ioutil.ReadFile(opts.Webhook.TLSPublicKey)
			if err2 != nil {
				return nil, fmt.Errorf("failed to load public key: %w", err2)
			}
		}

		if len(certBytes) != 0 {
			filePart, err3 := mw.CreateFormFile("certificate", "cert.pem")
			if err3 != nil {
				return nil, fmt.Errorf("failed to create form file: %w", err3)
			}

			_, err3 = filePart.Write(certBytes)
			if err3 != nil {
				return nil, fmt.Errorf("failed to write cert bytes: %w", err3)
			}
		}

		_ = mw.WriteField("max_connections", strconv.FormatInt(int64(opts.Webhook.MaxConnections), 10))
		allowedUpdatesBytes, err2 := json.Marshal(allowedUpdates)
		if err2 != nil {
			return nil, fmt.Errorf("failed to marshal allowed update types to json array: %w", err2)
		}

		_ = mw.WriteField("allowed_updates", string(allowedUpdatesBytes))
		_ = mw.WriteField("drop_pending_updates", "False")

		mw.Close()

		resp, err2 := tgClient.PostSetWebhookWithBody(ctx, mw.FormDataContentType(), body)
		if err2 != nil {
			return nil, fmt.Errorf("failed to request set webhook: %w", err2)
		}

		swr, err2 := telegram.ParsePostSetWebhookResponse(resp)
		_ = resp.Body.Close()
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse response of set webhook: %w", err2)
		}

		if swr.JSON200 == nil || !swr.JSON200.Ok {
			return nil, fmt.Errorf("failed to set telegram webhook: %s", swr.JSONDefault.Description)
		}

		mux.HandleFunc(opts.Webhook.Path, func(w http.ResponseWriter, r *http.Request) {
			dec := json.NewDecoder(r.Body)
			defer func() { _ = r.Body.Close() }()

			var dest struct {
				Ok     bool              `json:"ok"`
				Result []telegram.Update `json:"result"`
			}

			if err3 := dec.Decode(&dest); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logger.E("failed to decode webhook payload", log.Error(err3))
				return
			}

			_, err3 := client.onTelegramUpdate(dest.Result...)
			if err3 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				logger.E("failed to process telegram update", log.Error(err3))
				return
			}
		})
	} else {
		// delete webhook if exists
		resp, err2 := tgClient.PostGetWebhookInfo(ctx)
		if err2 != nil {
			return nil, fmt.Errorf("failed to check bot webhook status: %w", err2)
		}

		info, err2 := telegram.ParsePostGetWebhookInfoResponse(resp)
		_ = resp.Body.Close()
		if err2 != nil {
			return nil, fmt.Errorf("failed to parse bot webhook info: %w", err2)
		}

		// telegram will return a result with non-empty url if webhook is set
		// https://core.telegram.org/bots/api/#getwebhookinfo
		if info.JSON200 == nil || !info.JSON200.Ok {
			return nil, fmt.Errorf("telegram: get webhook info failed: %s", info.JSONDefault.Description)
		}

		if len(info.JSON200.Result.Url) != 0 {
			// TODO: handle error
			resp, err2 = tgClient.PostDeleteWebhook(ctx, telegram.PostDeleteWebhookJSONRequestBody{
				DropPendingUpdates: constant.False(),
			})
			if err2 != nil {
				return nil, fmt.Errorf("failed to delete webhook: %w", err2)
			}

			wd, err3 := telegram.ParsePostDeleteWebhookResponse(resp)
			_ = resp.Body.Close()
			if err3 != nil {
				return nil, fmt.Errorf("failed to parse webhook deletion response: %w", err3)
			}

			if wd.JSON200 == nil || !wd.JSON200.Ok {
				return nil, fmt.Errorf("telegram: delete webhook failed: %s", wd.JSONDefault.Description)
			}
		}

		// discuss long polling
		go func() {
			tk := time.NewTicker(2 * time.Second)
			defer tk.Stop()

			offset := 0
			for {
				select {
				case <-tk.C:
					// poll and ignore error
					offsetPtr := &offset
					if offset == 0 {
						offsetPtr = nil
					}

					func() {

					}()
					resp, err3 := tgClient.PostGetUpdates(
						ctx,
						telegram.PostGetUpdatesJSONRequestBody{
							AllowedUpdates: &allowedUpdates,
							Offset:         offsetPtr,
						},
					)
					if err3 != nil {
						logger.I("failed to poll updates", log.Error(err3))
						continue
					}

					updates, err3 := telegram.ParsePostGetUpdatesResponse(resp)
					_ = resp.Body.Close()

					if err3 != nil {
						logger.I("failed to parse updates", log.Error(err3))
						continue
					}

					if updates.JSON200 == nil || !updates.JSON200.Ok {
						logger.I("telegram: get updates failed", log.String("reason", updates.JSONDefault.Description))
						continue
					}

					if len(updates.JSON200.Result) == 0 {
						logger.V("no message update got")
						continue
					}

					// see https://core.telegram.org/bots/api/#getupdates
					// 		An update is considered confirmed as soon as getUpdates is called
					// 		with an offset higher than its update_id.
					maxID, err3 := client.onTelegramUpdate(updates.JSON200.Result...)
					if err3 != nil {
						logger.I("failed to process telegram update", log.Error(err3))
						continue
					}
					offset = maxID + 1
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	client.msgDelQ.Start(ctx.Done())
	msgDelCh := client.msgDelQ.TakeCh()
	go func() {
		for td := range msgDelCh {
			// delete message with best effort
			k := td.Key.(msgDeleteKey)
			for i := 0; i < 5; i++ {
				resp, err2 := tgClient.PostDeleteMessage(
					ctx,
					telegram.PostDeleteMessageJSONRequestBody{
						ChatId:    k.ChatID,
						MessageId: int(k.MessageID),
					},
				)
				if err2 != nil {
					continue
				}
				_ = resp.Body.Close()
			}
		}
	}()

	// everything working, set commands to keep them up to date (best effort)

	var commands []telegram.BotCommand

	for _, cmd := range constant.VisibleBotCommands {
		commands = append(commands, telegram.BotCommand{
			Command:     strings.TrimPrefix(cmd, "/"),
			Description: constant.BotCommandShortDescriptions[cmd],
		})
	}

	// set bot commands
	{
		resp, err := tgClient.PostSetMyCommands(ctx, telegram.PostSetMyCommandsJSONRequestBody{
			Commands: commands,
		})
		if err != nil {
			logger.E("failed to request set telegram bot commands", log.Error(err))
		} else {
			sr, err := telegram.ParsePostSetMyCommandsResponse(resp)
			_ = resp.Body.Close()
			if err != nil {
				logger.E("failed to parse bot command set response", log.Error(err))
			} else {
				if sr.JSON200 == nil || !sr.JSON200.Ok {
					logger.E("failed to set telegram bot commands", log.String("reason", sr.JSONDefault.Description))
				} else {
					logger.D("telegram bot command set", log.Any("commands", commands))
				}
			}
		}
	}

	return &telegramBot{client: tgClient}, nil
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
			return c.handleCmd(logger, uint64(msg.Chat.Id), cmd, strings.TrimSpace(string(utf16.Decode(content))), msg)
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

type tokenInputHandleFunc func(chatID uint64, userID uint64, msg *telegram.Message) (bool, error)

func (c *telegramBot) appendSessionMessage(
	logger log.Interface,
	chatID uint64,
	msg *telegram.Message,
) error {
	currentSession, ok := c.getActiveSession(chatID)
	if !ok {
		return nil
	}

	// ignore bot messages
	if msg.From != nil && msg.From.IsBot {
		return nil
	}

	logger.V("appending session meesage")
	m := newTelegramMessage(msg, c.botUsername)

	errCh, err := m.PreProcess(c, c.webArchiver, c.storage, currentSession.peekLastMessage())
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

	currentSession.appendMessage(m)

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

		_, ok := c.getActiveSession(chatID)
		if ok {
			logger.D("invalid command usage", log.String("reason", "already in a session"))
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Please <code>/end</code> current discussion before starting a new one",
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		if !c.markSessionStandby(userID, chatID, defaultChatUsername, topic, url, 5*time.Minute) {
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
					c.resolvePendingRequest(userID)
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
				fmt.Sprintf(textPrompt, c.generatorName),
				telegram.InlineKeyboardMarkup{
					InlineKeyboard: inlineKeyboard[:],
				},
			)
			return err
		}()
	case constant.CommandStart:
		return c.handleStartCommand(logger, chatID, userID, isPrivateMessage, params, msg)
	case constant.CommandEnd:
		prevReq, ok := c.resolvePendingRequest(userID)
		if ok {
			// a pending request, no generator involved
			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("You have canceled the pending <code>%s</code>", getCommandFromRequest(prevReq)),
			)

			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			if sr, isSR := prevReq.(*sessionRequest); isSR {
				if sr.ChatID != chatID {
					_, _ = c.sendTextMessage(
						sr.ChatID, true, true, 0,
						"Discussion was canceled by the initiator.",
					)
				}
			}

			return nil
		}

		currentSession, ok := c.getActiveSession(chatID)
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

		n, content, err := currentSession.generateContent()
		if err != nil {
			println("sending error message")
			_, _ = c.sendTextMessage(
				chatID, false, true, msg.MessageId,
				fmt.Sprintf("Internal bot error: failed to generate post content: %v", err),
			)

			// do not execute again on telegram redelivery
			return nil
		}

		postURL, err := currentSession.generator.Append(currentSession.Topic, content)
		if err != nil {
			_, _ = c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf("%s post update error: %v", currentSession.generator.Name(), err),
			)

			// do not execute again on telegram redelivery
			return nil
		}

		currentSession.deleteFirstNMessage(n)

		currentSession, ok = c.deactivateSession(chatID)
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
				currentSession.Topic, postURL,
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

		_, ok := c.getActiveSession(chatID)
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

		currentSession, ok := c.getActiveSession(chatID)
		if !ok {
			logger.D("invalid command usage", log.String("reason", "not in a session"))
			msgID, _ := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				"There is not active discussion, <code>/ignore</code> will do nothing in this case",
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		_ = currentSession.deleteMessage(formatTelegramMessageID(replyTo.MessageId))

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

		prevCmd, ok := c.markPendingEditing(userID, 5*time.Minute)
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
			fmt.Sprintf("Enter your %s token to edit", c.generatorName),
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
				fmt.Sprintf("Please specify the url(s) of the %s post(s) to be deleted", c.generatorName),
			)
			c.scheduleMessageDelete(chatID, 5*time.Second, uint64(msgID), uint64(msg.MessageId))

			return nil
		}

		prevCmd, ok := c.markPendingDeleting(userID, strings.Split(params, " "), 5*time.Minute)
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
			fmt.Sprintf("Enter your %s token to delete the post", c.generatorName),
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

		prevCmd, ok := c.markPendingListing(userID, 5*time.Minute)
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
			fmt.Sprintf("Enter your %s token to list your posts", c.generatorName),
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
			ChatID:    chatID,
			MessageID: msgID,
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
		standbySession         *sessionRequest
		expectedOriginalChatID uint64
	)

	switch action {
	case "create", "enter":
		var ok bool
		standbySession, ok = c.getStandbySession(userID)
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
		gen, userConfig, err2 := c.createGenerator()
		defer func() {
			if err2 != nil {
				_, _ = c.resolvePendingRequest(userID)

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

		token, err2 := gen.Login(userConfig)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("%s login failed: %v", gen.Name(), err2),
			)
			return err2
		}

		_, err2 = c.sendTextMessage(
			chatID, false, true, 0,
			fmt.Sprintf(
				"Here is your %s token, keep it on your own for later use:\n\n<pre>%s</pre>",
				gen.Name(), token,
			),
		)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, false, true, msg.MessageId,
				fmt.Sprintf("Internal bot error: unable to send %s token: %v", gen.Name(), err2),
			)
			return err2
		}

		content, err2 := gen.FormatPagePrefix()
		if err2 != nil {
			return fmt.Errorf("failed to generate initial page: %w", err2)
		}

		postURL, err2 := gen.Publish(standbySession.Topic, content)
		if err2 != nil {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("%s pre-publish failed: %v", gen.Name(), err2),
			)
			return err2
		}

		currentSession, err2 := c.activateSession(
			standbySession.ChatID,
			userID,
			standbySession.Topic,
			standbySession.ChatUsername,
			gen,
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
				_, _ = c.deactivateSession(standbySession.ChatID)
			}
		}()

		_, err2 = c.sendTextMessage(
			standbySession.ChatID, true, true, 0,
			fmt.Sprintf(
				"The post for your discussion around %q has been created: %s",
				currentSession.Topic, postURL,
			),
		)

		return nil
	case "enter":
		msgID, err2 := c.sendTextMessage(chatID, false, true, 0,
			fmt.Sprintf("Enter your %s token as a reply to this message", c.generatorName),
			telegram.ForceReply{
				ForceReply: true,
				Selective:  constant.True(),
			},
		)
		if err2 != nil {
			// this message must be sent to user, this error will trigger message redelivery
			return err2
		}

		if !c.markRequestExpectingInput(userID, uint64(msgID)) {
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
			fmt.Sprintf("Enter your %s token as a reply to this message", c.generatorName),
			telegram.ForceReply{
				ForceReply: true,
				Selective:  constant.True(),
			},
		)

		if !c.markRequestExpectingInput(userID, uint64(msgID)) {
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
