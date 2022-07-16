package telegram

import (
	"context"
	"fmt"
	"strings"

	"arhat.dev/pkg/log"
	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/tg"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

func (c *tgBot) Configure() (err error) {
	stop, err := bg.Connect(c.client, bg.WithContext(c.Context()))
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

	var self *tg.User
	switch t := auth.GetUser().(type) {
	case *tg.UserEmpty:
	case *tg.User:
		self = t
	default:
		self, err = c.client.Self(c.Context())
		if err != nil {
			return fmt.Errorf("recognize self: %w", err)
		}
	}

	c.isBot = self.GetBot()
	c.username, _ = self.GetUsername()

	c.Logger().D("recognized self",
		log.Int64("user_id", self.GetID()),
		log.String("username", c.username),
		log.Bool("is_bot", c.isBot),
	)

	if self.Bot {
		var (
			req   tg.BotsSetBotCommandsRequest
			scope tg.BotCommandScopeDefault
		)

		req.Scope = &scope

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

func (c *tgBot) onNewEncryptedMessage(ctx context.Context, e tg.Entities, update *tg.UpdateNewEncryptedMessage) error {
	switch m := update.GetMessage().(type) {
	case *tg.EncryptedMessage: // encryptedMessage#ed18c118
		c.Logger().V("new encrypted message", log.Uint32("type_id", m.TypeID()))
	case *tg.EncryptedMessageService: // encryptedMessageService#23734b06
		c.Logger().V("new encrypted service message", log.Uint32("type_id", m.TypeID()))
	default:
		c.Logger().I("unknown encrypted message type", log.Uint32("type_id", m.TypeID()))
	}

	return nil
}

func (c *tgBot) onNewChannelMessage(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
	return c.handleNewMessage(e, update.GetMessage())
}

func (c *tgBot) onNewLegacyMessage(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
	return c.handleNewMessage(e, update.GetMessage())
}

func (c *tgBot) handleNewMessage(e tg.Entities, msg tg.MessageClass) error {
	switch m := msg.(type) {
	case *tg.MessageEmpty:
		c.Logger().V("new empty message", log.Uint32("type_id", m.TypeID()))
		return nil
	case *tg.MessageService:
		c.Logger().V("new service message", log.Uint32("type_id", m.TypeID()))
		return nil
	case *tg.Message:
		var src messageSource

		chat, err := extractPeer(e, m.GetPeerID())
		if err != nil {
			c.Logger().E("bad chat for new message", log.Error(err))
			return err
		}
		src.Chat = resolveChatSpec(chat)

		c.Logger().V("new message", log.Uint32("type_id", m.TypeID()), log.Bool("pm", src.Chat.IsPrivateChat()))

		fromID, ok := m.GetFromID()
		if ok {
			var from any
			from, err = extractPeer(e, fromID)
			if err != nil {
				c.Logger().E("bad sender of new message", log.Error(err))
				return err
			}

			src.From, err = resolveAuthorSpec(from)
			if err != nil {
				c.Logger().E("unresolable sender", log.Error(err))
				return err
			}
		} else if src.Chat.IsPrivateChat() {
			src.From, err = resolveAuthorSpec(chat)
			if err != nil {
				c.Logger().E("unresolable sender", log.Error(err))
				return err
			}
		} else {
			c.Logger().E("unexpected message sent from anonymous user", rt.LogChatID(src.Chat.ID()))
			return nil
		}

		fwd, ok := m.GetFwdFrom()
		if ok {
			var fwdFrom authorSpec
			if fwdChatID, ok := fwd.GetFromID(); ok {
				var peer any
				peer, err = extractPeer(e, fwdChatID)
				if err != nil {
					c.Logger().E("bad fwd chat", log.Error(err))
					return err
				}

				fwdChat := resolveChatSpec(peer)
				switch {
				case fwdChat.IsChannelChat(), fwdChat.IsGroupChat():
					fwdFrom.username, ok = fwd.GetPostAuthor()
					if ok {
						fwdFrom.authorFlag |= authorFlag_User
					} else if fwdChat.IsChannelChat() {
						fwdFrom.authorFlag |= authorFlag_Channel
					} else {
						fwdFrom.authorFlag |= authorFlag_Group
					}

					src.FwdFrom.Set(fwdFrom)
				case fwdChat.IsLegacyGroupChat():
					// TODO
				case fwdChat.IsPrivateChat():
					fwdFrom, err = resolveAuthorSpec(peer)
					if err != nil {
						c.Logger().E("bad fwd user", log.Error(err))
						return err
					}
					src.FwdFrom.Set(fwdFrom)
				}

				src.FwdChat.Set(fwdChat)
			}
		}

		err = c.dispatchNewMessage(&src, m)
		if err != nil {
			c.Logger().I("bad message", log.Error(err))
		}

		return err
	default:
		c.Logger().E("unknown new message", log.Uint32("type_id", msg.TypeID()))
		return nil
	}
}
