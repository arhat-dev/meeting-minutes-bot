package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"arhat.dev/pkg/log"

	"arhat.dev/meeting-minutes-bot/pkg/conf"
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

	fu, err := storage.NewDriver(opts.Storage.Driver, opts.Storage.Config)
	if err != nil {
		return fmt.Errorf("failed to create file uploader: %w", err)
	}

	wa, err := webarchiver.NewDriver(opts.WebArchiver.Driver, opts.WebArchiver.Config)
	if err != nil {
		return fmt.Errorf("failed to create web archiver: %w", err)
	}

	gen, err := generator.NewDriver(opts.Generator.Driver, opts.Generator.Config)
	if err != nil {
		return fmt.Errorf("failed to create post generator: %w", err)
	}

	tgBot, err := createTelegramBot(
		ctx,
		log.Log.WithFields(log.String("bot", "telegram")),
		opts.PublicBaseURL,
		mux,
		fu,
		wa,
		gen,
		opts.Publisher.Driver,
		func() (publisher.Interface, publisher.UserConfig, error) {
			return publisher.NewDriver(opts.Publisher.Driver, opts.Publisher.Config)
		},
		&bots.Telegram,
	)
	if err != nil {
		return fmt.Errorf("failed to create telegram bot: %w", err)
	}

	_ = tgBot

	srv := &http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
		TLSConfig:   tlsConfig,
		Handler:     mux,
		Addr:        opts.Listen,
	}

	return srv.ListenAndServe()
}
