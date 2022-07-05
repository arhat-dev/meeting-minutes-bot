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
	for _, st := range opts.Storage {
		var cfg storage.Config
		for _, cfg = range st.Config {
			break
		}

		err = storageMgr.Add(st.MIMEMatch, st.MaxUploadSize, cfg)
		if err != nil {
			return fmt.Errorf("failed to add storage driver: %w", err)
		}
	}

	var (
		wa     webarchiver.Interface
		gen    generator.Interface
		pubCfg publisher.Config
	)

	for _, cfg := range opts.WebArchiver.Config {
		wa, err = cfg.Create()
		if err != nil {
			return fmt.Errorf("failed to create web archiver: %w", err)
		}

		break
	}

	for _, cfg := range opts.Generator.Config {
		gen, err = cfg.Create()
		if err != nil {
			return fmt.Errorf("failed to create post generator: %w", err)
		}

		break
	}

	for _, cfg := range opts.Publisher.Config {
		_, _, err = cfg.Create()
		if err != nil {
			return fmt.Errorf("failed to pre check publisher creation: %w", err)
		}

		pubCfg = cfg
		break
	}

	if bots.Telegram.Enabled {
		cmds, newToOld := getCommands(bots.GlobalCommandMapping, bots.Telegram.CommandsMapping)
		tgBot, err := telegram.Create(
			ctx,
			log.Log.WithFields(log.String("bot", "telegram")),
			storageMgr,
			wa,
			gen,
			pubCfg.Create,
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
		switch cmd {
		case constant.CommandStart:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.Start, botMapping.Start,
				),
			)
		case constant.CommandDiscuss:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.Discuss, botMapping.Discuss,
				),
			)
		case constant.CommandIgnore:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.Ignore, botMapping.Ignore,
				),
			)
		case constant.CommandInclude:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.Include, botMapping.Include,
				),
			)
		case constant.CommandEnd:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.End, botMapping.End,
				),
			)
		case constant.CommandCancel:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.Cancel, botMapping.Cancel,
				),
			)
		case constant.CommandContinue:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.Continue, botMapping.Continue,
				),
			)
		case constant.CommandEdit:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.Edit, botMapping.Edit,
				),
			)
		case constant.CommandList:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.List, botMapping.List,
				),
			)
		case constant.CommandDelete:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.Delete, botMapping.Delete,
				),
			)
		case constant.CommandHelp:
			setCmd(
				getCommandOverride(
					cmd,
					constant.BotCommandShortDescriptions[cmd],
					globalMapping.Help, botMapping.Help,
				),
			)
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
