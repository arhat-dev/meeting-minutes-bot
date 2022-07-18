package telegram

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"strings"
	"time"

	"arhat.dev/pkg/queue"
	"arhat.dev/pkg/stringhelper"
	"arhat.dev/rs"
	tds "github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/dcs"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/session"
)

const Platform = "telegram"

func init() {
	bot.Register(Platform, func() bot.Config { return &Config{} })
}

type Endpoint struct {
	rs.BaseField

	// Address (with port) of telegram server (e.g. 1.2.3.4:443)
	Address string `yaml:"address"`
	Test    bool   `yaml:"test"`
}

// Config for telegram bot
type Config struct {
	rs.BaseField

	bot.CommonConfig `yaml:",inline"`

	DC int `yaml:"dc"`
	// TODO: support custom servers
	Servers []Endpoint `yaml:"servers"`

	// Telegram app info obtained from https://my.telegram.org/apps
	AppID     int    `yaml:"appID"`
	AppHash   string `yaml:"appHash"`
	AppPubKey string `yaml:"appPubKey"`

	// BotToken of the telegram bot (fetched from BotFather)
	BotToken string `yaml:"botToken"`
}

func (c *Config) Create(rtCtx rt.RTContext, bctx *bot.CreationContext) (bot.Interface, error) {
	var (
		publicKeys []telegram.PublicKey
		dcList     dcs.List
	)

	if len(c.AppPubKey) != 0 {
		blk, _ := pem.Decode(stringhelper.ToBytes[byte, byte](c.AppPubKey))
		key, err := x509.ParsePKCS1PublicKey(blk.Bytes)
		if err != nil {
			return nil, err
		}

		publicKeys = []telegram.PublicKey{
			{RSA: key},
		}
	}

	workflows, err := c.CommonConfig.Resolve(bctx)
	if err != nil {
		return nil, fmt.Errorf("resolve workflow contexts: %w", err)
	}

	tb := &tgBot{
		BaseBot: bot.NewBotBase(rtCtx),

		botToken: strings.TrimSpace(c.BotToken),
		username: "", // set in Configure()

		dispatcher: tg.NewUpdateDispatcher(),

		sessions: session.NewManager[chatIDWrapper](rtCtx.Context()),

		wfSet: workflows,

		msgDelQ: queue.NewTimeoutQueue[msgDeleteKey, tg.InputPeerClass](),
	}

	tb.dispatcher.OnNewMessage(tb.onNewTelegramLegacyMessage)
	tb.dispatcher.OnNewChannelMessage(tb.onNewTelegramChannelMessage)
	tb.dispatcher.OnNewEncryptedMessage(tb.onNewTelegramEncryptedMessage)

	tb.client = telegram.NewClient(c.AppID, strings.TrimSpace(c.AppHash), telegram.Options{
		UpdateHandler:  tb.dispatcher,
		DC:             c.DC,
		DCList:         dcList,
		PublicKeys:     publicKeys,
		MaxRetries:     15,
		RetryInterval:  5 * time.Second,
		SessionStorage: &tds.StorageMemory{},
	})

	tb.sender = message.NewSender(tb.client.API())
	_ = tb.downloader.WithPartSize(512 * 1024)
	tb.uploader = uploader.NewUploader(tb.client.API()).WithThreads(3)

	return tb, nil
}
