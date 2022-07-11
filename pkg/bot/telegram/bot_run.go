package telegram

import (
	"fmt"
	"strings"

	"arhat.dev/pkg/log"
	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/tg"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
)

func (c *tgBot) Configure() (err error) {
	stop, err := bg.Connect(&c.client, bg.WithContext(c.Context()))
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer func() {
		if err != nil {
			stop()
		}
	}()

	// TODO: support user account
	auth, err := c.client.Auth().Bot(c.Context(), c.botToken)
	if err != nil {
		return fmt.Errorf("auth: %w", err)
	}
	_ = auth

	// get bot username
	self, err := c.client.Self(c.Context())
	if err != nil {
		return fmt.Errorf("recognize self: %w", err)
	}

	c.botUsername, _ = self.GetUsername()
	c.Logger().D("got bot username", log.String("username", c.botUsername))

	if self.Bot {
		var (
			req tg.BotsSetBotCommandsRequest
		)

		for _, wf := range c.wfSet.Workflows {
			for i, cmd := range wf.BotCommands.Commands {
				if len(cmd) == 0 || len(wf.BotCommands.Descriptions[i]) == 0 {
					continue
				}

				req.Commands = append(req.Commands, tg.BotCommand{
					Command:     strings.TrimPrefix(cmd, "/"),
					Description: wf.BotCommands.Descriptions[i],
				})
			}
		}

		_, err = c.client.API().BotsSetBotCommands(c.Context(), &req)
		if err != nil {
			return fmt.Errorf("set bot commands: %w", err)
		}

		c.Logger().D("bot commands updated", log.Any("commands", req.Commands))
	}

	return nil
}

// nolint:gocyclo
func (c *tgBot) Start(baseURL string, mux bot.Mux) error {
	c.msgDelQ.Start(c.Context().Done())
	go func() {
		msgDelCh := c.msgDelQ.TakeCh()

		for td := range msgDelCh {
			// delete message with best effort
			for i := 0; i < 5; i++ {

				// TODO: which to use?
				// c.client.API().MessagesDeleteMessages(c.Context(), &tg.MessagesDeleteMessagesRequest{
				// 	ID:     []int{},
				// 	Revoke: true,
				// })

				c.sender.To(td.Data).Revoke().Messages(c.Context(), int(td.Key.msgID))
				deleted, err2 := c.sender.Delete().Messages(c.Context(), int(td.Key.msgID))
				if err2 != nil {
					continue
				}
				_ = deleted

				break
			}
		}
	}()

	return nil
}

// func (c *tgBot) startBotAPI(baseURL string, mux bot.Mux) error {
// 	allowedUpdates := []string{
// 		"message",
// 		// TODO: support message edit, how could we present edited messages?
// 		// "edited_message",
// 	}
//
// 	if c.webhookConfig.Enabled && len(baseURL) != 0 {
// 		// set webhook
// 		base, err2 := url.Parse(baseURL)
// 		if err2 != nil {
// 			return fmt.Errorf("failed to parse base url for telegram bot webhook: %w", err2)
// 		}
//
// 		port := ""
// 		p := base.Port()
//
// 		switch {
// 		case len(p) != 0:
// 			switch p {
// 			case "443":
// 				// omit 443 since we are uisng https
// 			case "80", "88", "8443":
// 				port = ":" + p
// 			default:
// 				return fmt.Errorf("invalid port for telegram bot: port %s not allowed", port)
// 			}
// 		case len(p) == 0 && base.Scheme != "https":
// 			// not using https and no port set, lets guess!
// 			switch base.Scheme {
// 			case "http":
// 				port = ":80" // https over 80 is allowed by telegram
// 			default:
// 				return fmt.Errorf("invalid base url for telegram bot: unable to find port")
// 			}
// 		}
//
// 		body := &bytes.Buffer{}
// 		mw := multipart.NewWriter(body)
// 		err2 = mw.WriteField(
// 			"url",
// 			fmt.Sprintf(
// 				"https://%s%s%s",
// 				base.Host,
// 				port,
// 				path.Join(base.RawPath, c.webhookConfig.Path),
// 			),
// 		)
// 		if err2 != nil {
// 			return fmt.Errorf("failed to set url: %w", err2)
// 		}
//
// 		if len(c.webhookConfig.TLSPublicKey) != 0 {
// 			filePart, err3 := mw.CreateFormFile("certificate", "cert.pem")
// 			if err3 != nil {
// 				return fmt.Errorf("failed to create form file: %w", err3)
// 			}
//
// 			_, err3 = filePart.Write([]byte(c.webhookConfig.TLSPublicKey))
// 			if err3 != nil {
// 				return fmt.Errorf("failed to write cert bytes: %w", err3)
// 			}
// 		}
//
// 		_ = mw.WriteField("max_connections", strconv.FormatInt(int64(c.webhookConfig.MaxConnections), 10))
// 		allowedUpdatesBytes, err2 := json.Marshal(allowedUpdates)
// 		if err2 != nil {
// 			return fmt.Errorf("failed to marshal allowed update types to json array: %w", err2)
// 		}
//
// 		_ = mw.WriteField("allowed_updates", string(allowedUpdatesBytes))
// 		_ = mw.WriteField("drop_pending_updates", "False")
//
// 		_ = mw.Close()
//
// 		resp, err2 := c.client.PostSetWebhookWithBody(c.Context(), mw.FormDataContentType(), body)
// 		if err2 != nil {
// 			return fmt.Errorf("failed to request set webhook: %w", err2)
// 		}
//
// 		swr, err2 := api.ParsePostSetWebhookResponse(resp)
// 		_ = resp.Body.Close()
// 		if err2 != nil {
// 			return fmt.Errorf("failed to parse response of set webhook: %w", err2)
// 		}
//
// 		if swr.JSON200 == nil || !swr.JSON200.Ok {
// 			return fmt.Errorf("failed to set telegram webhook: %s", swr.JSONDefault.Description)
// 		}
//
// 		mux.HandleFunc(c.webhookConfig.Path, func(w http.ResponseWriter, r *http.Request) {
// 			dec := json.NewDecoder(r.Body)
// 			defer func() { _ = r.Body.Close() }()
//
// 			var dest struct {
// 				Ok     bool         `json:"ok"`
// 				Result []api.Update `json:"result"`
// 			}
//
// 			if err3 := dec.Decode(&dest); err3 != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				c.Logger().E("failed to decode webhook payload", log.Error(err3))
// 				return
// 			}
//
// 			_, err3 := c.onTelegramUpdate(dest.Result)
// 			if err3 != nil {
// 				w.WriteHeader(http.StatusInternalServerError)
// 				c.Logger().E("failed to process telegram update", log.Error(err3))
// 				return
// 			}
// 		})
// 	} else {
// 		// delete webhook if exists
// 		resp, err2 := c.client.PostGetWebhookInfo(c.Context())
// 		if err2 != nil {
// 			return fmt.Errorf("failed to check bot webhook status: %w", err2)
// 		}
//
// 		info, err2 := api.ParsePostGetWebhookInfoResponse(resp)
// 		_ = resp.Body.Close()
// 		if err2 != nil {
// 			return fmt.Errorf("failed to parse bot webhook info: %w", err2)
// 		}
//
// 		// telegram will return a result with non-empty url if webhook is set
// 		// https://core.telegram.org/bots/api/#getwebhookinfo
// 		if info.JSON200 == nil || !info.JSON200.Ok {
// 			return fmt.Errorf("telegram: get webhook info failed: %s", info.JSONDefault.Description)
// 		}
//
// 		if len(info.JSON200.Result.Url) != 0 {
// 			// TODO: handle error
// 			resp, err2 = c.client.PostDeleteWebhook(c.Context(), api.PostDeleteWebhookJSONRequestBody{
// 				DropPendingUpdates: constant.False(),
// 			})
// 			if err2 != nil {
// 				return fmt.Errorf("failed to delete webhook: %w", err2)
// 			}
//
// 			wd, err3 := api.ParsePostDeleteWebhookResponse(resp)
// 			_ = resp.Body.Close()
// 			if err3 != nil {
// 				return fmt.Errorf("failed to parse webhook deletion response: %w", err3)
// 			}
//
// 			if wd.JSON200 == nil || !wd.JSON200.Ok {
// 				return fmt.Errorf("telegram: delete webhook failed: %s", wd.JSONDefault.Description)
// 			}
// 		}
//
// 		// long polling
// 		go func() {
// 			tk := time.NewTicker(2 * time.Second)
// 			defer tk.Stop()
//
// 			offset := 0
// 			for {
// 				select {
// 				case <-tk.C:
// 					// poll and ignore error
// 					offsetPtr := &offset
// 					if offset == 0 {
// 						offsetPtr = nil
// 					}
//
// 					resp, err3 := c.client.PostGetUpdates(
// 						c.Context(),
// 						api.PostGetUpdatesJSONRequestBody{
// 							AllowedUpdates: &allowedUpdates,
// 							Offset:         offsetPtr,
// 						},
// 					)
// 					if err3 != nil {
// 						c.Logger().I("failed to poll updates", log.Error(err3))
// 						continue
// 					}
//
// 					updates, err3 := api.ParsePostGetUpdatesResponse(resp)
// 					_ = resp.Body.Close()
//
// 					if err3 != nil {
// 						c.Logger().I("failed to parse updates", log.Error(err3))
// 						continue
// 					}
//
// 					if updates.JSON200 == nil || !updates.JSON200.Ok {
// 						c.Logger().I(
// 							"telegram: get updates failed",
// 							log.String("reason", updates.JSONDefault.Description),
// 						)
// 						continue
// 					}
//
// 					if len(updates.JSON200.Result) == 0 {
// 						c.Logger().V("no message update got")
// 						continue
// 					}
//
// 					// see https://core.telegram.org/bots/api/#getupdates
// 					// 		An update is considered confirmed as soon as getUpdates is called
// 					// 		with an offset higher than its update_id.
// 					maxID, err3 := c.onTelegramUpdate(updates.JSON200.Result)
// 					if err3 != nil {
// 						c.Logger().I("failed to process telegram update", log.Error(err3))
// 						continue
// 					}
// 					offset = maxID + 1
// 				case <-c.Context().Done():
// 					return
// 				}
// 			}
// 		}()
// 	}
// }
