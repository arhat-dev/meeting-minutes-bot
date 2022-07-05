package conf

import (
	"arhat.dev/rs"
	"github.com/spf13/pflag"
)

// TelegramConfig for telegram bot
type TelegramConfig struct {
	rs.BaseField

	Enabled bool `yaml:"enabled"`

	Endpoint string `yaml:"endpoint"`
	BotToken string `yaml:"botToken"`

	CommandsMapping BotCommandsMappingConfig `yaml:"commandsMapping"`

	Webhook struct {
		rs.BaseField

		Enabled        bool   `yaml:"enabled"`
		Path           string `yaml:"path"`
		MaxConnections int32  `yaml:"maxConnections"`

		TLSPublicKey string `yaml:"tlsPublicKey"`
	} `yaml:"webhook"`
}

// nolint:lll
func flagsForTelegramConfig(prefix string, config *TelegramConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("app", pflag.ExitOnError)

	fs.BoolVar(&config.Enabled, prefix+"enabled", true, "run telegram bot")
	fs.StringVar(&config.Endpoint, prefix+"endpoint", "api.telegram.org", "set telegram bot api server address")
	fs.StringVar(&config.BotToken, prefix+"botToken", "", "set bot token, e.g. 123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11")

	fs.BoolVar(&config.Webhook.Enabled, prefix+"webhook.enabled", true, "enabled webhook when server address is not empty")
	fs.StringVar(&config.Webhook.Path, prefix+"webhook.path", "", "set the http path for this webhook, relative to the server base url, https will be used, and server port can be one of [443, 80, 88, 8443]")
	fs.Int32Var(&config.Webhook.MaxConnections, prefix+"webhook.maxConnections", 40, "max concurrent requests to this bot, 1 - 100")
	fs.StringVar(&config.Webhook.TLSPublicKey, prefix+"webhook.tlsPublicKey", "", "public key of your self-signed certificate, only needed if you are using self-signed certs for this webhook")

	return fs
}
