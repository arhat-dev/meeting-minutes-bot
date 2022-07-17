package bot

import (
	"arhat.dev/mbot/pkg/rt"
)

func NewBotBase(ctx rt.RTContext) BaseBot {
	return BaseBot{
		RTContext: ctx,
	}
}

type BaseBot struct {
	rt.RTContext
}
