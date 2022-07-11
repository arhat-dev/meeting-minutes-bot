package rt

import (
	"context"

	"arhat.dev/pkg/log"
)

func NewContext(ctx context.Context, logger log.Interface, cache Cache) RTContext {
	return RTContext{
		ctx:    ctx,
		logger: logger,
		cache:  cache,
	}
}

type RTContext struct {
	ctx    context.Context
	logger log.Interface
	cache  Cache
}

func (r *RTContext) Context() context.Context { return r.ctx }
func (r *RTContext) Logger() log.Interface    { return r.logger }
func (r *RTContext) Cache() Cache             { return r.cache }
