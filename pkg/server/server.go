package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"arhat.dev/pkg/log"

	"arhat.dev/meeting-minutes-bot/pkg/bot/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/conf"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

func Run(ctx context.Context, opts *conf.AppConfig, bots *conf.BotsConfig) error {
	tlsConfig, err := opts.TLS.GetTLSConfig(true)
	if err != nil {
		return fmt.Errorf("failed to load tls config: %w", err)
	}

	mux := http.NewServeMux()

	storageMgr := storage.NewManager()
	for _, cfg := range opts.Storage {
		err = storageMgr.Add(cfg.Driver, cfg.MIMEMatch, cfg.MaxUploadSize, cfg.Config)
		if err != nil {
			return fmt.Errorf("failed to add storage driver %q: %w", cfg.Driver, err)
		}
	}

	webarchiver, err := webarchiver.NewDriver(opts.WebArchiver.Driver, opts.WebArchiver.Config)
	if err != nil {
		return fmt.Errorf("failed to create web archiver: %w", err)
	}

	generator, err := generator.NewDriver(opts.Generator.Driver, opts.Generator.Config)
	if err != nil {
		return fmt.Errorf("failed to create post generator: %w", err)
	}

	_, _, err = publisher.NewDriver(opts.Publisher.Driver, opts.Publisher.Config)
	if err != nil {
		return fmt.Errorf("failed to pre check publisher creation: %w", err)
	}

	if bots.Telegram.Enabled {
		cmds, newToOld := getCommands(bots.GlobalCommandMapping, bots.Telegram.CommandsMapping)
		tgBot, err := telegram.Create(
			ctx,
			log.Log.WithFields(log.String("bot", "telegram")),
			storageMgr,
			webarchiver,
			generator,
			func() (publisher.Interface, publisher.UserConfig, error) {
				return publisher.NewDriver(opts.Publisher.Driver, opts.Publisher.Config)
			},
			cmds,
			newToOld,
			&bots.Telegram,
		)
		if err != nil {
			return fmt.Errorf("failed to create telegram bot: %w", err)
		}

		err = tgBot.Configure()
		if err != nil {
			return fmt.Errorf("failed to configure telegram bot: %w", err)
		}

		err = tgBot.Start(opts.PublicBaseURL, mux)
		if err != nil {
			return fmt.Errorf("failed to start telegram bot: %w", err)
		}
	}

	srv := &http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
		TLSConfig:   tlsConfig,
		Handler:     mux,
		Addr:        opts.Listen,
	}

	return srv.ListenAndServe()
}

func getCommands(
	globalMapping, botMapping conf.BotCommandsMappingConfig,
) (
	oldToNew map[string]conf.BotCommandMappingConfig,
	newToOld map[string]string,
) {
	oldToNew = make(map[string]conf.BotCommandMappingConfig)
	newToOld = make(map[string]string)
	setCmd := func(originalCmd, cmd, description string, disabled bool) {
		if disabled {
			return
		}

		oldToNew[originalCmd] = conf.BotCommandMappingConfig{
			As:          cmd,
			Description: description,
		}

		newToOld[cmd] = originalCmd
	}

	// use loop to ensure no command skipped
	for _, cmd := range constant.AllBotCommands {
		originalDescription := constant.BotCommandShortDescriptions[cmd]
		switch cmd {
		case constant.CommandStart:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.Start, botMapping.Start))
		case constant.CommandDiscuss:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.Discuss, botMapping.Discuss))
		case constant.CommandIgnore:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.Ignore, botMapping.Ignore))
		case constant.CommandInclude:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.Include, botMapping.Include))
		case constant.CommandEnd:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.End, botMapping.End))
		case constant.CommandCancel:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.Cancel, botMapping.Cancel))
		case constant.CommandContinue:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.Continue, botMapping.Continue))
		case constant.CommandEdit:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.Edit, botMapping.Edit))
		case constant.CommandList:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.List, botMapping.List))
		case constant.CommandDelete:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.Delete, botMapping.Delete))
		case constant.CommandHelp:
			setCmd(getCommandOverride(cmd, originalDescription, globalMapping.Help, botMapping.Help))
		default:
			panic(fmt.Errorf("command %s not processed", cmd))
		}
	}

	return
}

func getCommandOverride(
	originalCmd, originalDescription string,
	globalOverride, botOverride *conf.BotCommandMappingConfig,
) (oldCmd, newCmd, description string, disabled bool) {
	newCmd = originalCmd
	description = originalDescription

	if globalOverride != nil {
		newCmd = globalOverride.As
		if len(description) != 0 {
			description = globalOverride.Description
		}
	}

	if botOverride != nil {
		newCmd = botOverride.As
		if len(description) != 0 {
			description = botOverride.Description
		}
	}

	return originalCmd, newCmd, description, len(newCmd) == 0
}
