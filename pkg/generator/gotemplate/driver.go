package gotemplate

import (
	"fmt"
	"strings"

	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
)

// nolint:revive
const (
	Name = "gotemplate"
)

var _ generator.Interface = (*Driver)(nil)

type Driver struct {
	templates tplExecutor
}

func (g *Driver) RenderPageHeader() (_ string, err error) {
	var buf strings.Builder

	err = g.templates.ExecuteTemplate(&buf, "page.header", nil)
	if err != nil {
		return "", fmt.Errorf("execute page header template: %w", err)
	}

	return buf.String(), nil
}

type TemplateData struct {
	Messages []*rt.Message
}

func (g *Driver) RenderPageBody(msgs []*rt.Message) (_ string, err error) {
	var buf strings.Builder

	data := TemplateData{
		Messages: msgs,
	}
	err = g.templates.ExecuteTemplate(&buf, "page.body", &data)
	if err != nil {
		return "", fmt.Errorf("execute page body template: %w", err)
	}

	return buf.String(), nil
}
