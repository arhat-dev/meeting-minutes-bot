package generator

import "arhat.dev/meeting-minutes-bot/pkg/message"

type nopConfig struct{}

var _ Interface = (*nop)(nil)

type nop struct{}

func (a *nop) Name() string { return "nop" }

func (a *nop) FormatPageHeader() ([]byte, error)                           { return nil, nil }
func (a *nop) FormatPageBody(messages []message.Interface) ([]byte, error) { return nil, nil }
