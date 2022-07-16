package gotemplate

import (
	"bytes"
	"fmt"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

// nolint:revive
const (
	Name = "gotemplate"
)

var _ generator.Interface = (*Driver)(nil)

type Driver struct {
	templates tplExecutor
}

func (g *Driver) Name() string { return Name }

func (g *Driver) RenderPageHeader() (_ []byte, err error) {
	var (
		buf bytes.Buffer
	)

	err = g.templates.ExecuteTemplate(&buf, "page.header", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute page header template: %w", err)
	}

	return buf.Next(buf.Len()), nil
}

func (g *Driver) RenderPageBody(msgs []*rt.Message) (_ []byte, err error) {
	var (
		buf bytes.Buffer
	)

	data := generator.TemplateData{
		Messages: msgs,
	}
	err = g.templates.ExecuteTemplate(&buf, "page.body", &data)
	if err != nil {
		return nil, fmt.Errorf("failed to execute page template: %w", err)
	}

	return buf.Next(buf.Len()), nil
}
