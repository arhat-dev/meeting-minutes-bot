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

	"arhat.dev/pkg/log"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/conf"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
)

type telegramBot struct {
	ctx    context.Context
	logger log.Interface
	client *telegram.Client

	botUsername string

	// chart_id -> session
	*SessionManager
}

// nolint:gocyclo
func createTelegramBot(
	ctx context.Context,
	logger log.Interface,
	baseURL string,
	mux *http.ServeMux,
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

	allowedUpdates := []string{
		"message",
		"edited_message",
	}

	client := &telegramBot{
		ctx:    ctx,
		logger: logger,
		client: tgClient,

		botUsername: opts.BotUsername,

		SessionManager: newSessionManager(),
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

func (c *telegramBot) handleNewMessage(msg *telegram.Message) error {
	from := "unknown"
	if msg.From != nil && msg.From.Username != nil {
		from = *msg.From.Username
	}
	chat := "unknown"
	if msg.Chat.Username != nil {
		from = *msg.Chat.Username
	}

	logger := c.logger.WithFields(log.String("from", from), log.String("chat", chat))

	switch {
	case msg.Text != nil:
		// check command
		content := *msg.Text

		isCmd := false
		cmd := ""
		if msg.Entities != nil {
			for _, e := range *msg.Entities {
				// only check first command
				if e.Type == telegram.MessageEntityTypeBotCommand {
					isCmd = true
					cmd = content[e.Offset : e.Offset+e.Length]
					content = content[e.Offset+e.Length:]

					parts := strings.SplitN(cmd, "@", 2)
					cmd = parts[0]

					break
				}
			}
		}

		if isCmd {
			return c.handleCmd(logger, int64(msg.Chat.Id), cmd, content, msg)
		}

		return c.appendMessageToSession(logger, int64(msg.Chat.Id), msg)
	default:
		return c.appendMessageToSession(logger, int64(msg.Chat.Id), msg)
	}
}

func (c *telegramBot) appendMessageToSession(
	logger log.Interface, chatID int64, msg *telegram.Message,
) error {
	currentSession, ok := c.getActiveSession(chatID)
	if !ok {
		return nil
	}

	// ignore bot messages
	if msg.From != nil && msg.From.IsBot {
		return nil
	}

	logger.V("appending meesage")
	currentSession.appendMessage(msg)

	return nil
}

// nolint:gocyclo
func (c *telegramBot) handleCmd(logger log.Interface, chatID int64, cmd, params string, msg *telegram.Message) error {
	logger = logger.WithFields(log.String("cmd", cmd), log.String("params", params))

	if msg.From == nil {
		// ignore
		return c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"Sorry, anonymous users are not allowed",
		)
	}

	defaultChatUsername := ""
	// ensure only group admin can start discussion
	switch msg.Chat.Type {
	case telegram.ChatTypeChannel:
		return c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"This bot doesn't work in channels",
		)
	case telegram.ChatTypePrivate:
		// direct message to bot, no permission check
		defaultChatUsername = c.botUsername
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
				return c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					fmt.Sprintf("Unable to check group chat info: %v", err),
				)
			}

			chat, err := telegram.ParsePostGetChatResponse(resp)
			_ = resp.Body.Close()
			if err != nil {
				return c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					fmt.Sprintf("Unable to parse group chat info: %v", err),
				)
			}

			if chat.JSON200 == nil || !chat.JSON200.Ok {
				return c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					fmt.Sprintf("Telegram: unable to check group administrators: %s", chat.JSONDefault.Description),
				)
			}

			if usernamePtr := chat.JSON200.Result.Username; usernamePtr != nil {
				defaultChatUsername = *usernamePtr
			}
		}

		resp, err := c.client.PostGetChatAdministrators(c.ctx, telegram.PostGetChatAdministratorsJSONRequestBody{
			ChatId: chatID,
		})
		if err != nil {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Unable to check group administrators: %v", err),
			)
		}

		admins, err := telegram.ParsePostGetChatAdministratorsResponse(resp)
		_ = resp.Body.Close()
		if err != nil {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Unable to parse group administrators: %v", err),
			)
		}

		if admins.JSON200 == nil || !admins.JSON200.Ok {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Telegram: unable to check group administrators: %s", admins.JSONDefault.Description),
			)
		}

		isAdmin := false
		for _, admin := range admins.JSON200.Result {
			if admin.User.Id == msg.From.Id {
				isAdmin = true
				break
			}
		}

		if !isAdmin {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Sorry, only administrators can use this bot in group chat",
			)
		}
	}

	switch cmd {
	case "/discuss":
		if len(params) == 0 {
			logger.D("invalid usage of discuss", log.String("reason", "missing topic param"))
			return c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				"Please specify discussion topic, e.g. `/discuss foo`",
			)
		}

		// TODO: add inline keyboard for generator selection
		gen, err := generator.NewTelegraph()
		if err != nil {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Bot internal error: %v", err),
			)
		}

		// c.sendForceReply(chatID, msg.MessageId,
		// "Enter your telegraph auth token or create a new one with empty input")

		// we can discuss this session now, request a new telegraph user with user info,
		// and publish a telegraph post, send back the link to the post

		err = gen.Login(c.createTelegraphLoginConfigFromMessage(msg))
		if err != nil {
			_ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Telegraph login failed: %v", err),
			)
		}

		postURL, err := gen.Publish(params, []byte(
			// nolint:lll
			`<p><em>powered by <a href="https://github.com/arhat-dev/meeting-minutes-bot">meeting-minutes-bot</a></em></p><br>`,
		))
		if err != nil {
			_ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Telegraph pre-publish failed: %v", err),
			)
		}

		currentSession, ok := c.setSessionState(chatID, true, params, defaultChatUsername, gen)
		if !ok {
			logger.D("invalid usage of discuss", log.String("reason", "chat already in discussion"))
			_ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Please `/end` current discussion before starting a new one",
			)
		}

		_ = c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("The post for your discussion around %q has been created: %s", currentSession.Topic, postURL),
		)
	case "/end":
		currentSession, ok := c.getActiveSession(chatID)
		if !ok {
			// TODO
			logger.D("invalid usage of end", log.String("reason", "no active session"))
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"There is no active discussion",
			)
		}

		n, content := currentSession.generateHTMLContent()
		postURL, err := currentSession.generator.Append(currentSession.Topic, content)
		if err != nil {
			return c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf("Telegraph post update error: %v", err),
			)
		}

		currentSession.deleteFirstNMessage(n)

		currentSession, ok = c.setSessionState(chatID, false, "", "", nil)
		if !ok {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Bot internal error: discussion already been ended",
			)
		}

		return c.sendTextMessage(
			chatID, false, false, msg.MessageId,
			fmt.Sprintf(
				"Your discussion around %q has been ended, please view and edit post here: %s",
				currentSession.Topic, postURL,
			),
		)
	case "/ignore":
		replyTo := msg.ReplyToMessage
		if replyTo == nil {
			logger.D("invalid usage of ignore command", log.String("reason", "not a reply"))
			return c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				"`/ignore` can only be used as a reply",
			)
		}

		currentSession, ok := c.getActiveSession(chatID)
		if !ok {
			logger.D("invalid usage of ignore command", log.String("reason", "not in a session"))
			return c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				"There is not active discussion, `/ignore` will do nothing in this case",
			)
		}

		_ = currentSession.deleteMessage(replyTo.MessageId)

		logger.V("ignored message")
		return c.sendTextMessage(
			chatID, false, false, msg.MessageId,
			"The message you replied just got ignored",
		)
	case "/annotate":
		return c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"Sorry, command `/annotate` currently not implemented",
		)
	case "/continue":
		// pre-check
		_, ok := c.getActiveSession(chatID)
		if ok {
			logger.D("invalid usage of continue command", log.String("reason", "already in a session"))
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Please `/end` current discussion before starting a new one",
			)
		}

		gen, err := generator.NewTelegraph()
		if err != nil {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Bot internal error: %v", err),
			)
		}

		err = gen.Login(c.createTelegraphLoginConfigFromMessage(msg))
		if err != nil {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Bot internal error: %v", err),
			)
		}

		title, err := gen.Retrieve(params)
		if err != nil {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				fmt.Sprintf("Retrieve telegraph post error: %v", err),
			)
		}

		_, ok = c.setSessionState(chatID, true, title, defaultChatUsername, gen)
		if ok {
			return c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Please `/end` current discussion before starting a new one",
			)
		}

		return c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			"The post will be updated after the discussion",
		)
	default:
		logger.D("unknown command")

		return c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Command `%s` is not supported", cmd),
		)
	}

	return nil
}

// func (c *telegramBot) handleMessageEdit(me *telegram.Message) error {
// 	return nil
// }

func (c *telegramBot) createTelegraphLoginConfigFromMessage(msg *telegram.Message) *generator.TelegraphLoginConfig {
	loginConfig := &generator.TelegraphLoginConfig{}
	// select username
	username := ""
	switch {
	case msg.From != nil && msg.From.Username != nil:
		username = *msg.From.Username
	case msg.Chat.Username != nil:
		username = *msg.Chat.Username
	case msg.SenderChat != nil && msg.SenderChat.Username != nil:
		username = *msg.SenderChat.Username
	}

	loginConfig.AuthorName = username
	if len(username) != 0 {
		loginConfig.AuthorURL = fmt.Sprintf("https://t.me/%s", username)
	}

	return loginConfig
}

func (c *telegramBot) sendTextMessage(
	chatID int64,
	disableNotification,
	disableWebPreview bool,
	replyTo int,
	text string,
) error {
	var replyToMsgIDPtr *int
	if replyTo > 0 {
		replyToMsgIDPtr = &replyTo
	}

	resp, err := c.client.PostSendMessage(c.ctx, telegram.PostSendMessageJSONRequestBody{
		AllowSendingWithoutReply: constant.True(),
		ChatId:                   chatID,
		DisableNotification:      &disableNotification,
		DisableWebPagePreview:    &disableWebPreview,
		ReplyToMessageId:         replyToMsgIDPtr,
		Text:                     text,
	})
	if err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	result, err := telegram.ParsePostSendMessageResponse(resp)
	_ = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to parse response of message send: %w", err)
	}

	if result.JSON200 == nil && !result.JSON200.Ok {
		return fmt.Errorf("telegram: failed to send message: %s", result.JSONDefault.Description)
	}

	return nil
}

// func (c *telegramBot) sendForceReply(chatID int64, replyTo int, text string) error {
// 	forceReply := (interface{})(&telegram.ForceReply{
// 		ForceReply: true,
// 		Selective:  constant.True(),
// 	})

// 	resp, err := c.client.PostSendMessage(c.ctx, telegram.PostSendMessageJSONRequestBody{
// 		AllowSendingWithoutReply: constant.True(),
// 		ChatId:                   chatID,
// 		ReplyToMessageId:         &replyTo,
// 		ReplyMarkup:              &forceReply,
// 		Text:                     text,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to send force reply: %w", err)
// 	}

// 	result, err := telegram.ParsePostSendMessageResponse(resp)
// 	if err != nil {
// 		return fmt.Errorf("failed to parse response of force reply send: %w", err)
// 	}

// 	if result.JSON200 == nil && !result.JSON200.Ok {
// 		return fmt.Errorf("telegram: failed to send force reply: %s", result.JSONDefault.Description)
// 	}

// 	return nil
// }
