package rt

import (
	"context"

	"arhat.dev/pkg/log"
)

func NewContext(ctx context.Context, logger log.Interface) RTContext {
	return RTContext{
		ctx:    ctx,
		logger: logger,
	}
}

type RTContext struct {
	ctx    context.Context
	logger log.Interface
}

func (c *RTContext) Context() context.Context { return c.ctx }
func (c *RTContext) Logger() log.Interface    { return c.logger }
