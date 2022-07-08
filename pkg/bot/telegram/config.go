package telegram

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"arhat.dev/pkg/queue"
	"arhat.dev/rs"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	api "arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
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
	MaxConnections int32  `yaml:"maxConnections"`

	TLSPublicKey string `yaml:"tlsPublicKey"`
}

func (c *Config) Create(name string, rtCtx rt.RTContext, bctx *bot.Context) (ret bot.Interface, err error) {
	btk := strings.TrimSpace(c.BotToken)

	tgClient, err := api.NewClient(
		fmt.Sprintf("https://%s/bot%s/", strings.TrimSpace(c.Endpoint), btk),
		api.WithHTTPClient(&http.Client{
			// TODO: customize
		}),
		api.WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			return nil
		}),
	)
	if err != nil {
		err = fmt.Errorf("create api client: %w", err)
		return
	}

	workflows, err := c.CommonConfig.Resolve(bctx)
	if err != nil {
		err = fmt.Errorf("resolve workflow contexts: %w", err)
		return
	}

	ret = &telegramBot{
		BaseBot: bot.NewBotBase(name, rtCtx),

		botToken:    btk,
		botUsername: "", // set in Configure()

		client:         *tgClient,
		SessionManager: manager.NewSessionManager(rtCtx.Context()),

		wfSet:         workflows,
		webhookConfig: c.Webhook.webhookConfig,

		msgDelQ: *queue.NewTimeoutQueue[msgDeleteKey, struct{}](),
	}

	return
}
