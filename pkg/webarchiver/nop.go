package webarchiver

import (
	"context"
)

var _ Interface = (*nop)(nil)

type NopConfig struct{}

func (NopConfig) Create() (Interface, error) { return nop{}, nil }

type nop struct{}

func (nop) Archive(ctx context.Context, url string) (Result, error) { return nil, nil }
