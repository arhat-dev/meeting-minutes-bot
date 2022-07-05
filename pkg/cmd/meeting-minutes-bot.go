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

package cmd

import (
	"context"
	"fmt"
	"reflect"

	"arhat.dev/pkg/log"
	"arhat.dev/rs"
	"github.com/spf13/cobra"

	"arhat.dev/meeting-minutes-bot/pkg/conf"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/server"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

var (
	generatorConfigType   = reflect.TypeOf((*generator.Config)(nil)).Elem()
	publisherConfigType   = reflect.TypeOf((*publisher.Config)(nil)).Elem()
	storageConfigType     = reflect.TypeOf((*storage.Config)(nil)).Elem()
	webarchiverConfigType = reflect.TypeOf((*webarchiver.Config)(nil)).Elem()
)

type globalInterfaceTypeHandler struct{}

func (globalInterfaceTypeHandler) Create(typ reflect.Type, yamlKey string) (interface{}, error) {
	switch typ {
	case generatorConfigType:
		return generator.NewConfig(yamlKey)
	case publisherConfigType:
		return publisher.NewConfig(yamlKey)
	case storageConfigType:
		return storage.NewConfig(yamlKey)
	case webarchiverConfigType:
		return webarchiver.NewConfig(yamlKey)
	default:
		return nil, fmt.Errorf("unknown config type %s (key: %s)", typ.String(), yamlKey)
	}
}

func NewRootCmd() *cobra.Command {
	var (
		appCtx       context.Context
		configFile   string
		config       conf.Config
		cliLogConfig log.Config
	)

	rs.Init(&config, &rs.Options{
		InterfaceTypeHandler: globalInterfaceTypeHandler{},
	})

	rootCmd := &cobra.Command{
		Use:           "meeting-minutes-bot",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Use == "version" {
				return nil
			}

			var err error
			appCtx, err = conf.ReadConfig(cmd, &configFile, &cliLogConfig, &config)
			if err != nil {
				return err
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(appCtx, &config)
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringVarP(&configFile, "config", "c", constant.DefaultConfigFile,
		"path to the config file")
	flags.AddFlagSet(conf.FlagsForAppConfig("", &config.App))
	flags.AddFlagSet(conf.FlagsForBotsConfig("", &config.Bots))

	return rootCmd
}

func run(appCtx context.Context, config *conf.Config) error {
	return server.Run(appCtx, &config.App, &config.Bots)
}
