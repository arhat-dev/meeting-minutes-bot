package bot

import (
	"fmt"

	"arhat.dev/rs"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

// Interface of a single bot
type Interface interface {
	// Configure the bot, prepare to start
	Configure() error

	// Start the bot in background
	//
	// NOTE: implementation MUST be non-blocking
	Start(baseURL string, mux rt.Mux) error
}

type PublisherFactoryFunc = func() (publisher.Interface, publisher.User, error)

// Config type for single bot config
type Config interface {
	// Create a bot with this config
	Create(ctx rt.RTContext, bctx *CreationContext) (Interface, error)
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
func (c *CommonConfig) Resolve(bctx *CreationContext) (ret WorkflowSet, err error) {
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

		cmds := &ret.Workflows[i].BotCommands.Commands
		for _, cmd := range cmds {
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
