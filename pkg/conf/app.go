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
	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/pkg/log"
	"arhat.dev/pkg/tlshelper"
	"arhat.dev/rs"
)

type Config struct {
	rs.BaseField

	App AppConfig `yaml:"app"`

	bot.Context `yaml:",inline"`

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
