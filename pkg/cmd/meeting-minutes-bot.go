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

	"arhat.dev/pkg/log"
	"arhat.dev/rs"
	"github.com/spf13/cobra"

	"arhat.dev/meeting-minutes-bot/pkg/conf"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/server"
)

func NewRootCmd() *cobra.Command {
	var (
		appCtx       context.Context
		configFile   string
		config       conf.Config
		cliLogConfig log.Config
	)

	rs.Init(&config, &rs.Options{
		InterfaceTypeHandler: configIfaceHandler{},
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

	return rootCmd
}

func run(appCtx context.Context, config *conf.Config) error {
	return server.Run(appCtx, config)
}
