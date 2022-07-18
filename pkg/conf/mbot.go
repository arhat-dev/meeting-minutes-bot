package conf

import (
	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/storage"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/tlshelper"
	"arhat.dev/rs"
)

type Config struct {
	rs.BaseField

	App AppConfig `yaml:"app"`

	Storage    map[string]storage.Config   `yaml:"storage"`
	Generators map[string]generator.Config `yaml:"generators"`
	Publishers map[string]publisher.Config `yaml:"publishers"`

	// Bots are combinations of storage, generator, publisher
	Bots map[string]bot.Config `yaml:"bots"`
}

type AppConfig struct {
	rs.BaseField

	// CacheDir to store multi-media/files received from chat
	CacheDir string `yaml:"cacheDir"`

	Log log.ConfigSet `yaml:"log"`

	PublicBaseURL string `yaml:"publicBaseURL"`

	Listen string              `yaml:"listen"`
	TLS    tlshelper.TLSConfig `yaml:"tls"`
}
