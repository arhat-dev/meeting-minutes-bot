package generator

import (
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

var (
	_ Config    = (*nopConfig)(nil)
	_ Interface = (*nop)(nil)
)

type nopConfig struct{}

func (nopConfig) Create() (Interface, error) { return nop{}, nil }

type nop struct{}

func (nop) Name() string { return "nop" }

func (nop) RenderPageHeader() ([]byte, error)                     { return nil, nil }
func (nop) RenderPageBody(messages []*rt.Message) ([]byte, error) { return nil, nil }
