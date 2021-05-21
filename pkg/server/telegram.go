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
	"arhat.dev/meeting-minutes-bot/pkg/generator"
)

// nolint:lll
const (
	htmlHeadingPoweredBy = `<p><em>powered by <a href="https://github.com/arhat-dev/meeting-minutes-bot">meeting-minutes-bot</a></em></p><br>`
)

type msgDeleteKey struct {
	ChatID    uint64
	MessageID uint64
}

type telegramBot struct {
	ctx    context.Context
	logger log.Interface
	client *telegram.Client

	botUsername string

	// chart_id -> session
	*SessionManager

	msgDelQ *queue.TimeoutQueue
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
			return c.handleCmd(logger, uint64(msg.Chat.Id), cmd, string(utf16.Decode(content)), msg)
		}

		// filter private message for special replies to this bot
		if msg.Chat.Type == telegram.ChatTypePrivate && msg.From != nil && msg.ReplyToMessage != nil {
			userID := uint64(msg.From.Id)
			msgIDshouldReplyTo, isExpectingInput := c.sessionIsExpectingInput(userID)
			standbySession, hasStandbySession := c.getStandbySession(userID)

			// check if this message is an auth token
			if isExpectingInput &&
				uint64(msg.ReplyToMessage.MessageId) == msgIDshouldReplyTo &&
				hasStandbySession {
				chatID := uint64(msg.Chat.Id)

				gen, err := generator.NewTelegraph()
				defer func() {
					if err != nil {
						_ = c.cancelSessionStandby(userID)

						// best effort
						_, _ = c.sendTextMessage(
							chatID, true, true, 0,
							fmt.Sprintf("The discussion was canceled due to error, please retry later: %v", err),
						)

						if standbySession.ChatID != chatID {
							_, _ = c.sendTextMessage(
								standbySession.ChatID, true, true, 0,
								"The discussion was canceled due to error, please retry later",
							)
						}
					}
				}()

				if err != nil {
					_, _ = c.sendTextMessage(
						chatID, true, true, msg.MessageId,
						fmt.Sprintf("Internal bot error: %v", err),
					)
					return nil
				}

				_, err = gen.Login(&generator.TelegraphLoginConfig{
					AuthToken: strings.TrimSpace(*msg.Text),
				})
				if err != nil {
					_, _ = c.sendTextMessage(
						chatID, true, true, msg.MessageId,
						fmt.Sprintf("telegraph: auth error: %v", err),
					)
					// usually not our fault, let user try again
					err = nil
					return nil
				}

				var title string
				switch {
				case len(standbySession.Topic) != 0:
					// is /discuss, create a new post
					title = standbySession.Topic

					postURL, err2 := gen.Publish(title, []byte(htmlHeadingPoweredBy))
					if err2 != nil {
						_, _ = c.sendTextMessage(
							chatID, true, true, 0,
							fmt.Sprintf("Telegraph pre-publish failed: %v", err2),
						)
						return err2
					}

					_, err2 = c.sendTextMessage(
						standbySession.ChatID, true, true, 0,
						fmt.Sprintf(
							"The post for your discussion around %q has been created: %s",
							title, postURL,
						),
					)
					if err2 != nil {
						return err2
					}
				case len(standbySession.URL) != 0:
					// is /continue, find existing post to edit
					title, err = gen.Retrieve(standbySession.URL)
					if err != nil {
						_, _ = c.sendTextMessage(
							chatID, true, true, msg.MessageId,
							fmt.Sprintf("Retrieve telegraph post error: %v", err),
						)
						return err
					}
				}

				_, err = c.activateSession(
					standbySession.ChatID, userID, title,
					standbySession.ChatUsername, gen,
				)
				if err != nil {
					_, _ = c.sendTextMessage(
						chatID, true, true, msg.MessageId,
						fmt.Sprintf("Internal bot error: %v", err),
					)
					return err
				}

				_ = c.markSessionInputResolved(userID)
				defer func() {
					if err != nil {
						// bset effort
						_, _ = c.deactivateSession(standbySession.ChatID)
					}
				}()

				_, _ = c.sendTextMessage(
					standbySession.ChatID, true, true, msg.MessageId,
					"You can start your discussion now, the post will be updated after the discussion",
				)

				return nil
			}
		}

		return c.appendMessageToSession(logger, uint64(msg.Chat.Id), msg)
	default:
		return c.appendMessageToSession(logger, uint64(msg.Chat.Id), msg)
	}
}

func (c *telegramBot) appendMessageToSession(
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

	logger.V("appending meesage")
	currentSession.appendMessage(msg)

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
		c.scheduleMessageDelete(chatID, uint64(msgID), 5*time.Second)
		c.scheduleMessageDelete(chatID, uint64(msg.MessageId), 5*time.Second)
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
			topic = strings.TrimSpace(params)
			onInvalidCmdMsg = "Please specify discussion topic, e.g. <code>/discuss foo</code>"
		case constant.CommandContinue:
			url = strings.TrimSpace(params)
			// nolint:lll
			onInvalidCmdMsg = "Please specify url of the discussion post, e.g. <code>/continue https://telegra.ph/foo-01-21-100</code>"
		}

		if len(params) == 0 {
			logger.D("invalid command usage", log.String("reason", "missing topic param"))
			_, err := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				onInvalidCmdMsg,
			)
			return err
		}

		if !c.markSessionStandby(userID, chatID, defaultChatUsername, topic, url) {
			_, err := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"You have already started a discussion with no telegraph auth token specified, please end that first",
			)
			return err
		}

		_, ok := c.getActiveSession(chatID)
		if ok {
			logger.D("invalid command usage", log.String("reason", "already in a session"))
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Please <code>/end</code> current discussion before starting a new one",
			)
			return nil
		}

		return func() (err error) {
			defer func() {
				if err != nil {
					c.cancelSessionStandby(userID)
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
				textPrompt = "Create or enter telegraph auth token for this discussion"
				inlineKeyboard[0] = append(inlineKeyboard[0], telegram.InlineKeyboardButton{
					Text: "Create",
					Url:  &urlForCreate,
				})
			case constant.CommandContinue:
				textPrompt = "Enter telegraph auth token to continue this discussion"
			}

			inlineKeyboard[0] = append(inlineKeyboard[0], telegram.InlineKeyboardButton{
				Text: "Enter",
				Url:  &urlForEnter,
			})

			_, err = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				textPrompt,
				telegram.InlineKeyboardMarkup{
					InlineKeyboard: inlineKeyboard[:],
				},
			)
			return err
		}()
	case constant.CommandStart:
		if !isPrivateMessage {
			_, err := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"You cannot <code>/start</code> this bot in groups",
			)
			return err
		}

		payload := strings.TrimSpace(params)
		if len(payload) == 0 {
			_, err2 := c.sendTextMessage(chatID, true, true, msg.MessageId, "I am alive.")
			return err2
		}

		createOrEnter, err := base64.URLEncoding.DecodeString(payload)
		if err != nil {
			_, err2 := c.sendTextMessage(chatID, true, true, msg.MessageId, "I am still alive.")
			return err2
		}

		parts := strings.SplitN(string(createOrEnter), ":", 3)
		if len(parts) != 3 {
			_, err2 := c.sendTextMessage(chatID, true, true, msg.MessageId, "Told you, I'm alive.")
			return err2
		}

		originalUserID, err := decodeUint64Hex(parts[1])
		if err != nil {
			_, err2 := c.sendTextMessage(chatID, true, true, msg.MessageId, fmt.Sprintf("Internal bot error: %s", err))
			return err2
		}

		// ensure same user
		if originalUserID != userID {
			_, err2 := c.sendTextMessage(chatID, true, true, msg.MessageId, "This is not for you :(")
			return err2
		}

		originalChatID, err := decodeUint64Hex(parts[2])
		if err != nil {
			_, err2 := c.sendTextMessage(chatID, true, true, msg.MessageId, fmt.Sprintf("Internal bot error: %s", err))
			return err2
		}

		standbySession, ok := c.getStandbySession(userID)
		if !ok {
			_, err2 := c.sendTextMessage(chatID, true, true, msg.MessageId, "No discussion requested")
			return err2
		}

		if standbySession.ChatID != originalChatID {
			// should not happen, defensive check
			_, err2 := c.sendTextMessage(chatID, true, true, msg.MessageId, "Unexpected chat id not match")
			return err2
		}

		switch parts[0] {
		case "create":
			gen, err2 := generator.NewTelegraph()
			defer func() {
				if err2 != nil {
					_ = c.cancelSessionStandby(userID)

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

			token, err2 := gen.Login(c.createTelegraphLoginConfigFromMessage(msg))
			if err2 != nil {
				_, _ = c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					fmt.Sprintf("Telegraph login failed: %v", err2),
				)
				return err2
			}

			_, err2 = c.sendTextMessage(
				chatID, false, true, 0,
				fmt.Sprintf(
					"Here is your telegraph token, keep it on your own for later use:\n\n<pre>%s</pre>",
					token,
				),
			)
			if err2 != nil {
				_, _ = c.sendTextMessage(
					chatID, false, true, msg.MessageId,
					fmt.Sprintf("Internal bot error: unable to send telegraph token: %v", err2),
				)
				return err2
			}

			postURL, err2 := gen.Publish(standbySession.Topic, []byte(htmlHeadingPoweredBy))
			if err2 != nil {
				_, _ = c.sendTextMessage(
					chatID, true, true, msg.MessageId,
					fmt.Sprintf("Telegraph pre-publish failed: %v", err2),
				)
				return err2
			}

			currentSession, err2 := c.activateSession(
				standbySession.ChatID,
				userID,
				standbySession.Topic,
				defaultChatUsername,
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

			return err2
		case "enter":
			msgID, err2 := c.sendTextMessage(chatID, false, true, msg.MessageId,
				"Enter your telegraph auth token as a reply to this message",
				telegram.ForceReply{
					ForceReply: true,
					Selective:  constant.True(),
				},
			)
			if err2 != nil {
				return err2
			}

			ok := c.markSessionExpectingInput(userID, uint64(msgID))
			if !ok {
				_, _ = c.sendTextMessage(
					chatID, false, true, msg.MessageId,
					"Internal bot error: the discussion is not expecting any input",
				)
				return nil
			}

			return nil
		default:
			_, err = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Nice try!",
			)
			return err
		}
	case constant.CommandEnd:
		if c.cancelSessionStandby(userID) {
			_ = c.markSessionInputResolved(userID)

			// a standby session, no more action

			msgID, _ := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"You have canceled the unstarted discussion",
			)

			c.scheduleMessageDelete(chatID, uint64(msgID), 5*time.Second)
			return nil
		}

		currentSession, ok := c.getActiveSession(chatID)
		if !ok {
			// TODO
			logger.D("invalid usage of end", log.String("reason", "no active session"))
			_, err2 := c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"There is no active discussion",
			)
			return err2
		}

		n, content := currentSession.generateHTMLContent()
		postURL, err := currentSession.generator.Append(currentSession.Topic, content)
		if err != nil {
			_, _ = c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				fmt.Sprintf("Telegraph post update error: %v", err),
			)
			return err
		}

		currentSession.deleteFirstNMessage(n)

		currentSession, ok = c.deactivateSession(chatID)
		if !ok {
			_, _ = c.sendTextMessage(
				chatID, true, true, msg.MessageId,
				"Internal bot error: discussion already been ended",
			)
			return nil
		}

		_, err = c.sendTextMessage(
			chatID, false, false, msg.MessageId,
			fmt.Sprintf(
				"Your discussion around %q has been ended, please view and edit post here: %s",
				currentSession.Topic, postURL,
			),
		)

		return err
	case constant.CommandIgnore:
		replyTo := msg.ReplyToMessage
		if replyTo == nil {
			logger.D("invalid usage of ignore command", log.String("reason", "not a reply"))
			_, err2 := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				"<code>/ignore</code> can only be used as a reply",
			)
			return err2
		}

		currentSession, ok := c.getActiveSession(chatID)
		if !ok {
			logger.D("invalid usage of ignore command", log.String("reason", "not in a session"))
			_, err2 := c.sendTextMessage(
				chatID, false, false, msg.MessageId,
				"There is not active discussion, <code>/ignore</code> will do nothing in this case",
			)
			return err2
		}

		_ = currentSession.deleteMessage(replyTo.MessageId)

		logger.V("ignored message")
		msgID, err2 := c.sendTextMessage(
			chatID, false, false, msg.MessageId,
			"The message you replied just got ignored",
		)

		if err2 == nil {
			c.scheduleMessageDelete(chatID, uint64(msgID), 5*time.Second)
		}

		return err2
	default:
		logger.D("unknown command")

		_, err := c.sendTextMessage(
			chatID, true, true, msg.MessageId,
			fmt.Sprintf("Command `%s</code> is not supported", cmd),
		)

		return err
	}
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

// nolint:unparam
func (c telegramBot) scheduleMessageDelete(chatID, msgID uint64, after time.Duration) {
	_ = c.msgDelQ.OfferWithDelay(msgDeleteKey{
		ChatID:    chatID,
		MessageID: msgID,
	}, struct{}{}, after)
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
