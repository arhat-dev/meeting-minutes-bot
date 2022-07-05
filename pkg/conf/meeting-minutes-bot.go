/*
Copyright 2020 The arhat.dev Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package conf

import (
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/tlshelper"
	"arhat.dev/rs"
	"github.com/spf13/pflag"
)

type Config struct {
	rs.BaseField

	App  AppConfig  `yaml:"app"`
	Bots BotsConfig `yaml:"bots"`
}

type AppConfig struct {
	rs.BaseField

	Log log.ConfigSet `yaml:"log"`

	PublicBaseURL string `yaml:"publicBaseURL"`

	Listen string              `yaml:"listen"`
	TLS    tlshelper.TLSConfig `yaml:"tls"`

	Storage     []StorageConfig   `yaml:"storage"`
	WebArchiver WebArchiverConfig `yaml:"webarchiver"`
	Generator   GeneratorConfig   `yaml:"generator"`
	Publisher   PublisherConfig   `yaml:"publisher"`
}

type GeneratorConfig struct {
	rs.BaseField

	Config map[string]generator.Config `yaml:",inline"`
}

type WebArchiverConfig struct {
	rs.BaseField

	Config map[string]webarchiver.Config `yaml:",inline"`
}

type PublisherConfig struct {
	rs.BaseField

	Config map[string]publisher.Config `yaml:",inline"`
}

type StorageConfig struct {
	rs.BaseField

	MIMEMatch     string `yaml:"mimeMatch"`
	MaxUploadSize uint64 `yaml:"maxUploadSize"`

	Config map[string]storage.Config `yaml:",inline"`
}

func FlagsForAppConfig(prefix string, config *AppConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("app", pflag.ExitOnError)

	fs.StringVar(&config.Listen, prefix+"listen", ":18080",
		"http server listen address, required if you need to serve webhook")
	fs.StringVar(&config.PublicBaseURL, prefix+"publicBaseURL", "",
		"url for external endpoints like telegram server to access")

	return fs
}

type BotsConfig struct {
	rs.BaseField

	GlobalCommandMapping BotCommandsMappingConfig `yaml:"globalCommandsMapping"`

	Telegram TelegramConfig `yaml:"telegram"`
}

type BotCommandsMappingConfig struct {
	rs.BaseField

	Discuss  *BotCommandMappingConfig `yaml:"/discuss"`
	Continue *BotCommandMappingConfig `yaml:"/continue"`
	Ignore   *BotCommandMappingConfig `yaml:"/ignore"`
	Include  *BotCommandMappingConfig `yaml:"/include"`
	End      *BotCommandMappingConfig `yaml:"/end"`
	Cancel   *BotCommandMappingConfig `yaml:"/cancel"`

	Edit   *BotCommandMappingConfig `yaml:"/edit"`
	List   *BotCommandMappingConfig `yaml:"/list"`
	Delete *BotCommandMappingConfig `yaml:"/delete"`

	Help  *BotCommandMappingConfig `yaml:"/help"`
	Start *BotCommandMappingConfig `yaml:"/start"`
}

type BotCommandMappingConfig struct {
	As          string `yaml:"as"`
	Description string `yaml:"description"`
}

func FlagsForBotsConfig(prefix string, config *BotsConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("bots", pflag.ExitOnError)

	fs.AddFlagSet(flagsForTelegramConfig(prefix+"telegram.", &config.Telegram))

	return fs
}
