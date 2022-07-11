package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/conf"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
	"arhat.dev/pkg/log"
)

type nopMux struct{}

func (nopMux) HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request)) {}

func Run(ctx context.Context, opts *conf.Config) (err error) {
	var (
		mux   bot.Mux
		cache rt.Cache
	)

	if len(opts.App.CacheDir) == 0 {
		cache, err = rt.NewCache(".botcache")
	} else {
		cache, err = rt.NewCache(opts.App.CacheDir)
	}
	if err != nil {
		return fmt.Errorf("create caching: %w", err)
	}

	if len(opts.App.Listen) == 0 {
		mux = nopMux{}
	} else {
		mux = http.NewServeMux()
	}

	bots := make([]bot.Interface, 0, len(opts.Bots))
	for name, cfg := range opts.Bots {
		var b bot.Interface
		b, err = cfg.Create(name, rt.NewContext(ctx, log.Log.WithName(name), cache), &opts.Context)
		if err != nil {
			err = fmt.Errorf("create bot %q: %w", name, err)
			return
		}

		bots = append(bots, b)
	}

	for _, b := range bots {
		err = b.Configure()
		if err != nil {
			return fmt.Errorf("configure bot %q: %w", b.Name(), err)
		}

		err = b.Start(opts.App.PublicBaseURL, mux)
		if err != nil {
			return fmt.Errorf("start bot %q: %w", b.Name(), err)
		}
	}

	if len(opts.App.Listen) == 0 {
		<-ctx.Done()
		return nil
	}

	tlsConfig, err := opts.App.TLS.GetTLSConfig(true)
	if err != nil {
		return fmt.Errorf("load tls config: %w", err)
	}

	srv := http.Server{
		BaseContext: func(net.Listener) context.Context { return ctx },
		TLSConfig:   tlsConfig,
		Handler:     mux.(*http.ServeMux),
		Addr:        opts.App.Listen,
	}

	return srv.ListenAndServe()
}
