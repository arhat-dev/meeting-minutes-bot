package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/conf"
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/storage"
	"arhat.dev/pkg/log"
)

type nopMux struct{}

func (nopMux) HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request)) {}

func Run(ctx context.Context, opts *conf.Config) (err error) {
	var (
		cache rt.Cache
		bctx  bot.CreationContext
	)

	bctx.Generators = make(map[string]generator.Interface, len(opts.Generators))
	for k, cfg := range opts.Generators {
		bctx.Generators[k], err = cfg.Create()
		if err != nil {
			err = fmt.Errorf("create generator %q: %w", k, err)
			return
		}
	}

	bctx.Storage = make(map[string]storage.Interface, len(opts.Storage))
	for k, cfg := range opts.Storage {
		bctx.Storage[k], err = cfg.Create()
		if err != nil {
			err = fmt.Errorf("create storage %q: %w", k, err)
			return
		}
	}

	for k, cfg := range opts.Publishers {
		_, _, err = cfg.Create()
		if err != nil {
			err = fmt.Errorf("check publisher creation %q: %w", k, err)
			return
		}
	}
	bctx.Publishers = opts.Publishers

	if len(opts.App.CacheDir) == 0 {
		cache, err = rt.NewCache(".botcache")
	} else {
		cache, err = rt.NewCache(opts.App.CacheDir)
	}
	if err != nil {
		return fmt.Errorf("create caching: %w", err)
	}

	type Pair struct {
		Name string
		Bot  bot.Interface
	}

	bots := make([]Pair, 0, len(opts.Bots))
	for name, cfg := range opts.Bots {
		var p Pair

		p.Name = name
		p.Bot, err = cfg.Create(rt.NewContext(ctx, log.Log.WithName(name), cache), &bctx)
		if err != nil {
			err = fmt.Errorf("create bot %q: %w", name, err)
			return
		}

		bots = append(bots, p)
	}

	bctx.Generators = nil
	bctx.Publishers = nil
	bctx.Storage = nil

	for _, b := range bots {
		err = b.Bot.Configure()
		if err != nil {
			return fmt.Errorf("configure bot %q: %w", b.Name, err)
		}
	}

	var mux rt.Mux
	if len(opts.App.Listen) == 0 {
		mux = nopMux{}
	} else {
		mux = http.NewServeMux()
	}

	for _, b := range bots {
		err = b.Bot.Start(opts.App.PublicBaseURL, mux)
		if err != nil {
			return fmt.Errorf("start bot %q: %w", b.Name, err)
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
