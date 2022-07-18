package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/rshelper"
	"arhat.dev/rs"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"arhat.dev/mbot/pkg/conf"
	"arhat.dev/mbot/pkg/server"
)

// nolint:revive
const (
	DefaultConfigFile = "/etc/mbot/config.yaml"
)

func NewRootCmd() *cobra.Command {
	var (
		appCtx       context.Context
		configFile   string
		config       conf.Config
		cliLogConfig log.Config
	)

	rs.Init(&config, &rs.Options{
		InterfaceTypeHandler: newConfigIfaceHandler(),
	})

	rootCmd := &cobra.Command{
		Use:           "mbot",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Use == "version" {
				return nil
			}

			var err error
			appCtx, err = readConfig(cmd, &configFile, &cliLogConfig, &config)
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

	flags.StringVarP(&configFile, "config", "c", DefaultConfigFile, "path to the config file")

	return rootCmd
}

func run(appCtx context.Context, config *conf.Config) error {
	return server.Run(appCtx, config)
}

func readConfig(
	cmd *cobra.Command,
	configFile *string,
	cliLogConfig *log.Config,
	config *conf.Config,
) (context.Context, error) {
	flags := cmd.Flags()
	configBytes, err := os.ReadFile(*configFile)
	if err != nil && flags.Changed("config") {
		return nil, fmt.Errorf("failed to read config file %s: %v", *configFile, err)
	}

	err = yaml.Unmarshal(configBytes, config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	osEnv := os.Environ()
	env := make(map[string]string, len(osEnv))

	for _, kv := range osEnv {
		k, v, _ := strings.Cut(kv, "=")
		env[k] = v
	}

	err = config.ResolveFields(rshelper.DefaultRenderingManager(env, nil), -1)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config: %w", err)
	}

	if len(config.App.Log) > 0 {
		if flags.Changed("log.format") {
			config.App.Log[0].Format = cliLogConfig.Format
		}

		if flags.Changed("log.level") {
			config.App.Log[0].Level = cliLogConfig.Level
		}

		if flags.Changed("log.file") {
			config.App.Log[0].File = cliLogConfig.File
		}
	} else {
		config.App.Log = append(config.App.Log, *cliLogConfig)
	}

	if err = cmd.ParseFlags(os.Args); err != nil {
		return nil, err
	}

	err = log.SetDefaultLogger(config.App.Log)
	if err != nil {
		return nil, fmt.Errorf("failed to set default logger: %w", err)
	}

	appCtx, exit := context.WithCancel(context.Background())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		exitCount := 0
		for sig := range sigCh {
			switch sig {
			case os.Interrupt, syscall.SIGTERM:
				exitCount++
				if exitCount == 1 {
					exit()
				} else {
					os.Exit(1)
				}
				//case syscall.SIGHUP:
				//	// force reload
			}
		}
	}()

	return appCtx, nil
}
