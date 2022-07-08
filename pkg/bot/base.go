package bot

import (
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

func NewBotBase(name string, ctx rt.RTContext) BaseBot {
	return BaseBot{
		RTContext: ctx,
		name:      name,
	}
}

type BaseBot struct {
	rt.RTContext

	name string
}

func (b *BaseBot) Name() string { return b.name }
