package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/queue"
	"github.com/gotd/contrib/bg"
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
	sender     *message.Sender
	uploader   *uploader.Uploader
	dispatcher tg.UpdateDispatcher
	downloader downloader.Downloader

	sessions session.Manager[chatIDWrapper]

	wfSet   bot.WorkflowSet
	msgDelQ *queue.TimeoutQueue[msgDeleteKey, tg.InputPeerClass]
}

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
		const (
			reqDefault = iota
			reqUsers
			reqChats
			reqChatAdmins

			REQ_COUNT
		)

		var (
			requests [REQ_COUNT]tg.BotsSetBotCommandsRequest

			defaultScope    tg.BotCommandScopeDefault
			usersScope      tg.BotCommandScopeUsers
			chatsScope      tg.BotCommandScopeChats
			chatAdminsScope tg.BotCommandScopeChatAdmins

			// TODO: support peer scopes?
			// peerScope       tg.BotCommandScopePeer
			// peerAdminsScope tg.BotCommandScopePeerAdmins
			// peerUserScope   tg.BotCommandScopePeerUser
		)

		requests[reqDefault].Scope = &defaultScope
		requests[reqUsers].Scope = &usersScope
		requests[reqChats].Scope = &chatsScope
		requests[reqChatAdmins].Scope = &chatAdminsScope

		for _, wf := range c.wfSet.Workflows {
			for i, cmd := range wf.BotCommands.Commands {
				if len(cmd) == 0 || len(wf.BotCommands.Descriptions[i]) == 0 {
					continue
				}

				cmdSpec := tg.BotCommand{
					Command:     strings.TrimPrefix(cmd, "/"),
					Description: wf.BotCommands.Descriptions[i],
				}

				for j := 0; j < REQ_COUNT; j++ {
					switch {
					case j == reqChats && wf.RequireAdmin():
						// skip
					default:
						requests[j].Commands = append(requests[j].Commands, cmdSpec)
					}
				}
			}
		}

		for i := 0; i < REQ_COUNT; i++ {
			_, err = c.client.API().BotsSetBotCommands(c.Context(), &requests[i])
			if err != nil {
				return fmt.Errorf("set bot commands: %w", err)
			}
		}

		c.Logger().D("bot commands updated", log.Any("commands", requests[reqDefault].Commands))
	}

	return nil
}

// nolint:gocyclo
func (c *tgBot) Start(baseURL string, mux rt.Mux) error {
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

func (c *tgBot) onNewTelegramEncryptedMessage(ctx context.Context, e tg.Entities, update *tg.UpdateNewEncryptedMessage) error {
	switch m := update.GetMessage().(type) {
	case *tg.EncryptedMessage:
		c.Logger().V("new encrypted message", log.Uint32("type_id", m.TypeID()))
	case *tg.EncryptedMessageService:
		c.Logger().V("new encrypted service message", log.Uint32("type_id", m.TypeID()))
	default:
		c.Logger().I("unknown encrypted message type", log.Uint32("type_id", m.TypeID()))
	}

	return nil
}

func (c *tgBot) onNewTelegramChannelMessage(ctx context.Context, e tg.Entities, update *tg.UpdateNewChannelMessage) error {
	return c.handleTelegramMessage(e, update.GetMessage())
}

func (c *tgBot) onNewTelegramLegacyMessage(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
	return c.handleTelegramMessage(e, update.GetMessage())
}

type messageContext struct {
	con    conversationImpl
	src    source
	msg    *tg.Message
	logger log.Interface
}

func (c *tgBot) handleTelegramMessage(e tg.Entities, msg tg.MessageClass) error {
	switch m := msg.(type) {
	case *tg.MessageEmpty:
		c.Logger().V("new empty message", log.Uint32("type_id", m.TypeID()))
		return nil
	case *tg.MessageService:
		c.Logger().V("new service message", log.Uint32("type_id", m.TypeID()))
		return nil
	case *tg.Message:
		var (
			mc messageContext
		)

		// resolve chat
		chat, err := extractPeer(e, m.GetPeerID())
		if err != nil {
			c.Logger().E("bad chat of new message", log.Error(err))
			return err
		}
		mc.src.Chat = resolveChatSpec(chat)

		mc.con = conversationImpl{
			bot:  c,
			peer: mc.src.Chat.InputPeer(),
		}

		c.Logger().V("new message", log.Uint32("type_id", m.TypeID()), log.Bool("pm", mc.src.Chat.IsPrivateChat()))

		// resolve sender
		fromID, ok := m.GetFromID()
		if ok {
			var from any
			from, err = extractPeer(e, fromID)
			if err != nil {
				c.Logger().E("bad sender of new message", log.Error(err))
				return err
			}

			mc.src.From, err = resolveAuthorSpec(from)
			if err != nil {
				c.Logger().E("unresolable sender", log.Error(err))
				return err
			}
		} else if mc.src.Chat.IsPrivateChat() {
			mc.src.From, err = resolveAuthorSpec(chat)
			if err != nil {
				c.Logger().E("unresolable sender", log.Error(err))
				return err
			}
		} else {
			c.Logger().E("unexpected message sent from anonymous user", rt.LogChatID(mc.src.Chat.ID()))
			return nil
		}

		// resolve orignial chat and sender
		fwd, ok := m.GetFwdFrom()
		if ok {
			var fwdFrom authorInfo
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

					mc.src.FwdFrom.Set(fwdFrom)
				case fwdChat.IsLegacyGroupChat():
					// TODO
				case fwdChat.IsPrivateChat():
					fwdFrom, err = resolveAuthorSpec(peer)
					if err != nil {
						c.Logger().E("bad fwd user", log.Error(err))
						return err
					}
					mc.src.FwdFrom.Set(fwdFrom)
				}

				mc.src.FwdChat.Set(fwdChat)
			}
		}

		mc.msg = m
		mc.logger = c.Logger().WithFields(
			rt.LogChatID(mc.src.Chat.ID()),
			rt.LogSenderID(mc.src.From.ID()),
		)

		err = c.dispatchNewMessage(&mc)
		if err != nil {
			c.Logger().I("bad message", log.Error(err))
		}

		return err
	default:
		c.Logger().E("unknown new message", log.Uint32("type_id", msg.TypeID()))
		return nil
	}
}

// nolint:gocyclo
func (c *tgBot) dispatchNewMessage(mc *messageContext) error {
	mc.logger.V("dispatch message")
	if len(mc.msg.GetMessage()) == 0 {
		return c.appendSessionMessage(mc)
	}

	var (
		isCmd   bool
		cmd     string
		params  string
		mention string
	)

	entities, _ := mc.msg.GetEntities()
	for _, v := range entities {
		e, ok := v.(*tg.MessageEntityBotCommand)
		if !ok {
			continue
		}

		isCmd = true
		cmd = mc.msg.Message[e.GetOffset() : e.GetOffset()+e.GetLength()]
		cmd, mention, ok = strings.Cut(cmd, "@")
		if ok /* found '@' */ && mention != c.username {
			continue
		}

		params = strings.TrimSpace(mc.msg.Message[e.GetOffset()+e.GetLength():])
		break
	}

	if isCmd {
		return c.handleBotCmd(mc, cmd, params)
	}

	// not containing bot command
	// filter private message for special replies to this bot
	if mc.src.Chat.IsPrivateChat() {
		replyToHdr, ok := mc.msg.GetReplyTo()
		mc.logger.V("check potential input", log.Bool("do_check", ok))
		if ok {
			replyTo := rt.MessageID(replyToHdr.ReplyToMsgID)
			for _, handle := range [...]inputHandleFunc{
				c.tryToHandleInputForDiscussOrContinue,
				c.tryToHandleInputForEditing,
				c.tryToHandleInputForListing,
				c.tryToHandleInputForDeleting,
			} {
				done, err := handle(mc, replyTo)
				if done {
					return err
				}
			}
		}
	}

	return c.appendSessionMessage(mc)
}

func (c *tgBot) appendSessionMessage(mc *messageContext) (err error) {
	mc.logger.V("check active session")
	s, ok := c.sessions.GetActiveSession(mc.src.Chat.ID())
	if !ok {
		return nil
	}

	mc.logger.V("append session message")
	m := newMessageFromTelegramMessage(mc)

	err = c.fillMessageSpans(mc, s.Workflow(), m)
	if err != nil {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(err.Error()),
		)

		return
	}

	s.AppendMessage(m)

	return nil
}

// handleBotCmd handles single command with all params as a single string
// nolint:gocyclo
func (c *tgBot) handleBotCmd(
	mc *messageContext, cmd, params string,
) error {
	mc.logger.V("handle bot command", log.String("cmd", cmd))
	if !mc.src.From.IsUser() {
		_, _ = c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Anonymous user not allowed"),
		)

		return nil
	}

	wf, ok := c.wfSet.WorkflowFor(cmd)
	if !ok {
		msgID, _ := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Unknown command: "),
			styling.Code(cmd),
		)

		c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))
		return nil
	}

	if wf.RequireAdmin() && !mc.src.Chat.IsPrivateChat() {
		// ensure only admin can use this bot in group

		if mc.src.Chat.IsLegacyGroupChat() {
			_, _ = c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Plain("Unsupported chat type: please consider upgrading this chat to supergroup."),
			)

			return nil
		}

		pt, err := c.client.API().ChannelsGetParticipant(c.Context(), &tg.ChannelsGetParticipantRequest{
			Channel: &tg.InputChannel{
				ChannelID:  mc.src.Chat.InputPeer().(*tg.InputPeerChannel).GetChannelID(),
				AccessHash: mc.src.Chat.InputPeer().(*tg.InputPeerChannel).GetAccessHash(),
			},
			Participant: &tg.InputPeerUser{
				UserID:     mc.src.From.InputUser().GetUserID(),
				AccessHash: mc.src.From.InputUser().GetAccessHash(),
			},
		})
		if err != nil || len(pt.Users) != 1 {
			_, _ = c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Plain("Internal bot error: "),
				styling.Code("unable to verify the initiator."),
			)
			return err
		}

		switch pt.GetParticipant().(type) {
		case *tg.ChannelParticipantCreator:
		case *tg.ChannelParticipantAdmin:
		default:
			msgID, _ := c.sendTextMessage(
				c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
				styling.Plain("Only administrators can use this bot in group chat"),
			)

			c.scheduleMessageDelete(&mc.src.Chat, 5*time.Second, msgID, rt.MessageID(mc.msg.GetID()))
			return nil
		}
	}

	switch bc := wf.BotCommands.Parse(cmd); bc {
	case rt.BotCmd_New:
		return c.handleBotCmd_Discuss(mc, wf, cmd, params)
	case rt.BotCmd_Resume:
		return c.handleBotCmd_Continue(mc, wf, cmd, params)
	case rt.BotCmd_Start:
		return c.handleBotCmd_Start(mc, wf, params)
	case rt.BotCmd_Cancel:
		return c.handleBotCmd_Cancel(mc, wf, cmd, params)
	case rt.BotCmd_End:
		return c.handleBotCmd_End(mc, wf, cmd, params)
	case rt.BotCmd_Include:
		return c.handleBotCmd_Include(mc, wf, cmd, params)
	case rt.BotCmd_Ignore:
		return c.handleBotCmd_Ignore(mc, wf, cmd, params)
	case rt.BotCmd_Edit:
		return c.handleBotCmd_Edit(mc, wf, cmd, params)
	case rt.BotCmd_Delete:
		return c.handleBotCmd_Delete(mc, wf, cmd, params)
	case rt.BotCmd_List:
		return c.handleBotCmd_List(mc, wf, cmd, params)
	case rt.BotCmd_Help:
		return c.handleBotCmd_Help(mc)
	default:
		mc.logger.E("unhandled cmd", log.String("cmd", cmd))

		_, err := c.sendTextMessage(
			c.sender.To(mc.src.Chat.InputPeer()).NoWebpage().Silent().Reply(mc.msg.GetID()),
			styling.Plain("Internal bot error: "),
			styling.Bold(cmd),
			styling.Bold(" not handled."),
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
