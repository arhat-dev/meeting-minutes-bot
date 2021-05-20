package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"arhat.dev/pkg/log"

	"arhat.dev/meeting-minutes-bot/pkg/conf"
)

func Run(ctx context.Context, opts *conf.AppConfig, bots *conf.BotsConfig) error {
	tlsConfig, err := opts.TLS.GetTLSConfig(true)
	if err != nil {
		return fmt.Errorf("failed to load tls config: %w", err)
	}

	mux := http.NewServeMux()
	tgBot, err := createTelegramBot(
		ctx,
		log.Log.WithFields(log.String("bot", "telegram")),
		opts.PublicBaseURL,
		mux,
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
