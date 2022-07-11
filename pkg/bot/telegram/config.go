package telegram

import (
	"context"
	"fmt"
	"strings"

	"arhat.dev/pkg/queue"
	"arhat.dev/rs"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/manager"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

const Platform = "telegram"

func init() {
	bot.Register(Platform, func() bot.Config { return &Config{} })
}

// Config for telegram bot
type Config struct {
	rs.BaseField

	bot.CommonConfig `yaml:",inline"`

	// Endpoint of telegram api (e.g. api.telegram.org)
	Endpoint string `yaml:"endpoint"`
	// BotToken of the telegram bot (fetched from BotFather)
	BotToken string `yaml:"botToken"`

	Webhook struct {
		rs.BaseField

		webhookConfig `yaml:",inline"`
	} `yaml:"webhook"`
}

type webhookConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Path           string `yaml:"path"`
	MaxConnections int32  `yaml:"maxConn"`

	TLSPublicKey string `yaml:"tlsPublicKey"`
}

func (c *Config) Create(name string, rtCtx rt.RTContext, bctx *bot.Context) (bot.Interface, error) {
	workflows, err := c.CommonConfig.Resolve(bctx)
	if err != nil {
		return nil, fmt.Errorf("resolve workflow contexts: %w", err)
	}

	tb := &tgBot{
		BaseBot: bot.NewBotBase(name, rtCtx),

		botToken:    strings.TrimSpace(c.BotToken),
		botUsername: "", // set in Configure()

		dispatcher: tg.NewUpdateDispatcher(),

		SessionManager: manager.NewSessionManager[*Message](rtCtx.Context()),

		wfSet:         workflows,
		webhookConfig: c.Webhook.webhookConfig,

		msgDelQ: *queue.NewTimeoutQueue[msgDeleteKey, tg.InputPeerClass](),
	}

	tb.client = *telegram.NewClient(0, "", telegram.Options{
		UpdateHandler: tb.dispatcher,
	})

	tb.sender = *message.NewSender(tb.client.API())
	tb.downloader = *downloader.NewDownloader()
	tb.uploader = *uploader.NewUploader(tb.client.API())

	tb.dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, update *tg.UpdateNewMessage) error {
		switch m := update.GetMessage().(type) {
		case *tg.MessageEmpty:
		case *tg.MessageService:
		case *tg.Message:
			var (
				fwdChat any
				from    any
				fwdFrom any

				src messageSource
			)

			chat, err := extractPeer(e, m.GetPeerID())
			if err != nil {
				return fmt.Errorf("bad chat: %w", err)
			}
			src.Chat = resolveChatSpec(chat)

			fromID, ok := m.GetFromID()
			if ok {
				from, err = extractPeer(e, fromID)
				if err != nil {
					return fmt.Errorf("bad msg from: %w", err)
				}

				src.From.Set(resolveAuthorSpec(from))
			}

			fwdFromHdr, ok := m.GetFwdFrom()
			if ok {
				{
					fwdFromID, ok := fwdFromHdr.GetFromID()
					if ok {
						fwdFrom, err = extractPeer(e, fwdFromID)
						if err != nil {
							return fmt.Errorf("bad msg fwd from: %w", err)
						}

						src.FwdFrom.Set(resolveAuthorSpec(fwdFrom))
					}
				}
				{
					fwdChatID, ok := fwdFromHdr.GetSavedFromPeer()
					if ok {
						fwdChat, err = extractPeer(e, fwdChatID)
						if err != nil {
							return fmt.Errorf("bad fwd chat: %w", err)
						}

						src.FwdChat.Set(resolveChatSpec(fwdChat))
					}
				}
			}

			return tb.handleNewMessage(&src, m)
		default:
			panic("unreachable")
		}

		return nil
	})

	return tb, nil
}
