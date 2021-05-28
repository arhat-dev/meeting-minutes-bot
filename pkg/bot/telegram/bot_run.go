package telegram

import (
	"bytes"
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

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
)

func (c *telegramBot) Configure() error {
	// get bot username
	resp, err2 := c.client.PostGetMe(c.ctx)
	if err2 != nil {
		return fmt.Errorf("failed to get my own info: %w", err2)
	}

	mr, err2 := telegram.ParsePostGetMeResponse(resp)
	_ = resp.Body.Close()
	if err2 != nil {
		return fmt.Errorf("failed to parse bot get info response: %w", err2)
	}

	if mr.JSON200 == nil || !mr.JSON200.Ok {
		return fmt.Errorf("failed to get bot info: %s", mr.JSONDefault.Description)
	}

	if mr.JSON200.Result.Username == nil {
		return fmt.Errorf("bot username not returned by server")
	}

	c.botUsername = *mr.JSON200.Result.Username
	c.logger.D("got bot username", log.String("username", c.botUsername))

	// everything working, set commands to keep them up to date (best effort)

	var commands []telegram.BotCommand

	for _, cmd := range constant.VisibleBotCommands {
		commands = append(commands, telegram.BotCommand{
			Command:     strings.TrimPrefix(cmd, "/"),
			Description: constant.BotCommandShortDescriptions[cmd],
		})
	}

	// set bot commands
	resp, err := c.client.PostSetMyCommands(c.ctx, telegram.PostSetMyCommandsJSONRequestBody{
		Commands: commands,
	})
	if err != nil {
		c.logger.E("failed to request set telegram bot commands", log.Error(err))
	} else {
		sr, err := telegram.ParsePostSetMyCommandsResponse(resp)
		_ = resp.Body.Close()
		if err != nil {
			c.logger.E("failed to parse bot command set response", log.Error(err))
		} else {
			if sr.JSON200 == nil || !sr.JSON200.Ok {
				c.logger.E("failed to set telegram bot commands", log.String("reason", sr.JSONDefault.Description))
			} else {
				c.logger.D("telegram bot command set", log.Any("commands", commands))
			}
		}
	}

	return nil
}

// nolint:gocyclo
func (c *telegramBot) Start(baseURL string, mux bot.Mux) error {
	allowedUpdates := []string{
		"message",
		// TODO: support message edit, how could we present edited messages?
		// "edited_message",
	}

	if c.opts.Webhook.Enabled && len(baseURL) != 0 {
		// set webhook
		base, err2 := url.Parse(baseURL)
		if err2 != nil {
			return fmt.Errorf("failed to parse base url for telegram bot webhook: %w", err2)
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
				return fmt.Errorf("invalid port for telegram bot: port %s not allowed", port)
			}
		case len(p) == 0 && base.Scheme != "https":
			// not using https and no port set, lets guess!
			switch base.Scheme {
			case "http":
				port = ":80" // https over 80 is allowed by telegram
			default:
				return fmt.Errorf("invalid base url for telegram bot: unable to find port")
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
				path.Join(base.RawPath, c.opts.Webhook.Path),
			),
		)
		if err2 != nil {
			return fmt.Errorf("failed to set url: %w", err2)
		}

		var certBytes []byte
		switch {
		case len(c.opts.Webhook.TLSPublicKeyData) != 0:
			// use embedded data
			certBytes, err2 = base64.StdEncoding.DecodeString(c.opts.Webhook.TLSPublicKeyData)
			if err2 != nil {
				return fmt.Errorf("failed to decode base64 encoded cert: %w", err2)
			}
		case len(c.opts.Webhook.TLSPublicKey) != 0 && len(c.opts.Webhook.TLSPublicKeyData) == 0:
			// read from file
			certBytes, err2 = ioutil.ReadFile(c.opts.Webhook.TLSPublicKey)
			if err2 != nil {
				return fmt.Errorf("failed to load public key: %w", err2)
			}
		}

		if len(certBytes) != 0 {
			filePart, err3 := mw.CreateFormFile("certificate", "cert.pem")
			if err3 != nil {
				return fmt.Errorf("failed to create form file: %w", err3)
			}

			_, err3 = filePart.Write(certBytes)
			if err3 != nil {
				return fmt.Errorf("failed to write cert bytes: %w", err3)
			}
		}

		_ = mw.WriteField("max_connections", strconv.FormatInt(int64(c.opts.Webhook.MaxConnections), 10))
		allowedUpdatesBytes, err2 := json.Marshal(allowedUpdates)
		if err2 != nil {
			return fmt.Errorf("failed to marshal allowed update types to json array: %w", err2)
		}

		_ = mw.WriteField("allowed_updates", string(allowedUpdatesBytes))
		_ = mw.WriteField("drop_pending_updates", "False")

		mw.Close()

		resp, err2 := c.client.PostSetWebhookWithBody(c.ctx, mw.FormDataContentType(), body)
		if err2 != nil {
			return fmt.Errorf("failed to request set webhook: %w", err2)
		}

		swr, err2 := telegram.ParsePostSetWebhookResponse(resp)
		_ = resp.Body.Close()
		if err2 != nil {
			return fmt.Errorf("failed to parse response of set webhook: %w", err2)
		}

		if swr.JSON200 == nil || !swr.JSON200.Ok {
			return fmt.Errorf("failed to set telegram webhook: %s", swr.JSONDefault.Description)
		}

		mux.HandleFunc(c.opts.Webhook.Path, func(w http.ResponseWriter, r *http.Request) {
			dec := json.NewDecoder(r.Body)
			defer func() { _ = r.Body.Close() }()

			var dest struct {
				Ok     bool              `json:"ok"`
				Result []telegram.Update `json:"result"`
			}

			if err3 := dec.Decode(&dest); err3 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				c.logger.E("failed to decode webhook payload", log.Error(err3))
				return
			}

			_, err3 := c.onTelegramUpdate(dest.Result...)
			if err3 != nil {
				w.WriteHeader(http.StatusInternalServerError)
				c.logger.E("failed to process telegram update", log.Error(err3))
				return
			}
		})
	} else {
		// delete webhook if exists
		resp, err2 := c.client.PostGetWebhookInfo(c.ctx)
		if err2 != nil {
			return fmt.Errorf("failed to check bot webhook status: %w", err2)
		}

		info, err2 := telegram.ParsePostGetWebhookInfoResponse(resp)
		_ = resp.Body.Close()
		if err2 != nil {
			return fmt.Errorf("failed to parse bot webhook info: %w", err2)
		}

		// telegram will return a result with non-empty url if webhook is set
		// https://core.telegram.org/bots/api/#getwebhookinfo
		if info.JSON200 == nil || !info.JSON200.Ok {
			return fmt.Errorf("telegram: get webhook info failed: %s", info.JSONDefault.Description)
		}

		if len(info.JSON200.Result.Url) != 0 {
			// TODO: handle error
			resp, err2 = c.client.PostDeleteWebhook(c.ctx, telegram.PostDeleteWebhookJSONRequestBody{
				DropPendingUpdates: constant.False(),
			})
			if err2 != nil {
				return fmt.Errorf("failed to delete webhook: %w", err2)
			}

			wd, err3 := telegram.ParsePostDeleteWebhookResponse(resp)
			_ = resp.Body.Close()
			if err3 != nil {
				return fmt.Errorf("failed to parse webhook deletion response: %w", err3)
			}

			if wd.JSON200 == nil || !wd.JSON200.Ok {
				return fmt.Errorf("telegram: delete webhook failed: %s", wd.JSONDefault.Description)
			}
		}

		// long polling
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
					resp, err3 := c.client.PostGetUpdates(
						c.ctx,
						telegram.PostGetUpdatesJSONRequestBody{
							AllowedUpdates: &allowedUpdates,
							Offset:         offsetPtr,
						},
					)
					if err3 != nil {
						c.logger.I("failed to poll updates", log.Error(err3))
						continue
					}

					updates, err3 := telegram.ParsePostGetUpdatesResponse(resp)
					_ = resp.Body.Close()

					if err3 != nil {
						c.logger.I("failed to parse updates", log.Error(err3))
						continue
					}

					if updates.JSON200 == nil || !updates.JSON200.Ok {
						c.logger.I(
							"telegram: get updates failed",
							log.String("reason", updates.JSONDefault.Description),
						)
						continue
					}

					if len(updates.JSON200.Result) == 0 {
						c.logger.V("no message update got")
						continue
					}

					// see https://core.telegram.org/bots/api/#getupdates
					// 		An update is considered confirmed as soon as getUpdates is called
					// 		with an offset higher than its update_id.
					maxID, err3 := c.onTelegramUpdate(updates.JSON200.Result...)
					if err3 != nil {
						c.logger.I("failed to process telegram update", log.Error(err3))
						continue
					}
					offset = maxID + 1
				case <-c.ctx.Done():
					return
				}
			}
		}()
	}

	c.msgDelQ.Start(c.ctx.Done())
	msgDelCh := c.msgDelQ.TakeCh()
	go func() {
		for td := range msgDelCh {
			// delete message with best effort
			k := td.Key.(msgDeleteKey)
			for i := 0; i < 5; i++ {
				resp, err2 := c.client.PostDeleteMessage(
					c.ctx,
					telegram.PostDeleteMessageJSONRequestBody{
						ChatId:    k.chatID,
						MessageId: int(k.messageID),
					},
				)
				if err2 != nil {
					continue
				}
				_ = resp.Body.Close()
			}
		}
	}()

	return nil
}
