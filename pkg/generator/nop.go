package generator

import "arhat.dev/meeting-minutes-bot/pkg/message"

type nopConfig struct{}

var _ Interface = (*nop)(nil)

type nop struct{}

func (a *nop) Name() string { return "nop" }

func (a *nop) RenderPageHeader() ([]byte, error)                           { return nil, nil }
func (a *nop) RenderPageBody(messages []message.Interface) ([]byte, error) { return nil, nil }
