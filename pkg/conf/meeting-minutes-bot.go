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
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/tlshelper"
	"github.com/spf13/pflag"
)

type Config struct {
	App  AppConfig  `json:"app" yaml:"app"`
	Bots BotsConfig `json:"bots" yaml:"bots"`
}

type AppConfig struct {
	Log log.ConfigSet `json:"log" yaml:"log"`

	PublicBaseURL string `json:"publicBaseURL" yaml:"publicBaseURL"`

	Listen string              `json:"listen" yaml:"listen"`
	TLS    tlshelper.TLSConfig `json:"tls" yaml:"tls"`

	Storage     []StorageConfig   `json:"storage" yaml:"storage"`
	WebArchiver WebArchiverConfig `json:"webarchiver" yaml:"webarchiver"`
	Generator   GeneratorConfig   `json:"generator" yaml:"generator"`
	Publisher   PublisherConfig   `json:"publisher" yaml:"publisher"`
}

func FlagsForAppConfig(prefix string, config *AppConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("app", pflag.ExitOnError)

	fs.StringVar(&config.Listen, prefix+"listen", ":18080",
		"http server listen address, required if you need to serve webhook")
	fs.StringVar(&config.PublicBaseURL, prefix+"publicBaseURL", "",
		"url for external endpoints like telegram server to access")

	fs.AddFlagSet(tlshelper.FlagsForTLSConfig(prefix+"tls", &config.TLS))

	fs.StringVar(&config.WebArchiver.Driver,
		prefix+"webarchiver.driver", "", "set web archive service provider, one of [cdp], leave empty to disable")
	fs.StringVar(&config.Generator.Driver,
		prefix+"generator.driver", "",
		"set post generation engine, one of [gotemplate, file], leave empty to disable",
	)
	fs.StringVar(&config.Publisher.Driver,
		prefix+"publisher.driver", "",
		"set post publisher service provider, one of [telegraph, file, http, interpreter], leave empty to disable",
	)

	return fs
}

type BotsConfig struct {
	GlobalCommandMapping BotCommandsMappingConfig `json:"globalCommandsMapping" yaml:"globalCommandsMapping"`

	Telegram TelegramConfig `json:"telegram" yaml:"telegram"`
}

type BotCommandsMappingConfig struct {
	Discuss  *BotCommandMappingConfig `json:"/discuss" yaml:"/discuss"`
	Continue *BotCommandMappingConfig `json:"/continue" yaml:"/continue"`
	Ignore   *BotCommandMappingConfig `json:"/ignore" yaml:"/ignore"`
	Include  *BotCommandMappingConfig `json:"/include" yaml:"/include"`
	End      *BotCommandMappingConfig `json:"/end" yaml:"/end"`
	Cancel   *BotCommandMappingConfig `json:"/cancel" yaml:"/cancel"`

	Edit   *BotCommandMappingConfig `json:"/edit" yaml:"/edit"`
	List   *BotCommandMappingConfig `json:"/list" yaml:"/list"`
	Delete *BotCommandMappingConfig `json:"/delete" yaml:"/delete"`

	Help  *BotCommandMappingConfig `json:"/help" yaml:"/help"`
	Start *BotCommandMappingConfig `json:"/start" yaml:"/start"`
}

type BotCommandMappingConfig struct {
	As          string `json:"as" yaml:"as"`
	Description string `json:"description" yaml:"description"`
}

func FlagsForBotsConfig(prefix string, config *BotsConfig) *pflag.FlagSet {
	fs := pflag.NewFlagSet("bots", pflag.ExitOnError)

	fs.AddFlagSet(flagsForTelegramConfig(prefix+"telegram.", &config.Telegram))

	return fs
}
