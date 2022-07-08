package bot

import (
	"fmt"
	"net/http"

	"arhat.dev/rs"

	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

// Mux is just an interface alternative to http.ServeMux
type Mux interface {
	HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request))
}

// Interface of a single bot
type Interface interface {
	Name() string

	// Configure the bot, prepare to start
	Configure() error

	// Start the bot in background
	//
	// NOTE: implementation MUST be non-blocking
	Start(baseURL string, mux Mux) error
}

type PublisherFactoryFunc = func() (publisher.Interface, publisher.UserConfig, error)

// Config type for single bot config
type Config interface {
	// Create a bot with this config
	Create(name string, ctx rt.RTContext, bctx *Context) (Interface, error)
}

// CommonConfig commonly used config for single bot config
type CommonConfig struct {
	rs.BaseField

	// Enabled set to true to enable this bot
	Enabled bool `yaml:"enabled"`

	// Workflows this bot accepts
	Workflows []WorkflowConfig `yaml:"workflows"`
}

// Resolve workflows
func (c *CommonConfig) Resolve(bctx *Context) (ret WorkflowSet, err error) {
	if !c.Enabled || len(c.Workflows) == 0 {
		return
	}

	ret = WorkflowSet{
		index:     make(map[string]int),
		Workflows: make([]Workflow, len(c.Workflows)),
	}

	for i, wfc := range c.Workflows {
		ret.Workflows[i], err = wfc.Resolve(bctx)
		if err != nil {
			err = fmt.Errorf("resolve #%d workflow: %w", i, err)
			return
		}

		for _, cmd := range ret.Workflows[i].BotCommands.Commands {
			if len(cmd) == 0 {
				continue
			}

			_, ok := ret.index[cmd]
			if ok {
				err = fmt.Errorf("duplicate cmd %q", cmd)
				return
			}

			ret.index[cmd] = i
		}
	}

	return
}

type configFactoryFunc = func() Config

var (
	supportedPlatforms = map[string]configFactoryFunc{}
)

func Register(platform string, cf configFactoryFunc) {
	// reserve empty name
	if platform == "" {
		return
	}

	supportedPlatforms[platform] = cf
}

func NewConfig(platform string) (any, error) {
	cf, ok := supportedPlatforms[platform]
	if !ok {
		return nil, fmt.Errorf("unknown bot platform %q", platform)
	}

	return cf(), nil
}
