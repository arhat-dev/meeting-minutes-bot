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

// New implements generator.Interface
func (d *Driver) New(con rt.Conversation, cmd, params string) (_ string, err error) {
	var buf strings.Builder

	data := Data{
		Command: cmd,
		Params:  params,
	}

	err = d.templates.ExecuteTemplate(&buf, "gen.new", &data)
	if err != nil {
		return "", fmt.Errorf("execute template gen.new: %w", err)
	}

	return buf.String(), nil
}

// Continue implements generator.Interface
func (d *Driver) Continue(con rt.Conversation, cmd string, params string) (_ string, err error) {
	var buf strings.Builder

	data := Data{
		Command: cmd,
		Params:  params,
	}

	err = d.templates.ExecuteTemplate(&buf, "gen.new", &data)
	if err != nil {
		return "", fmt.Errorf("execute template gen.continue: %w", err)
	}

	return buf.String(), nil
}

type Data struct {
	Command  string
	Params   string
	Messages []*rt.Message
}

// RenderBody implements generator.Interface
func (d *Driver) RenderBody(con rt.Conversation, msgs []*rt.Message) (_ string, err error) {
	var buf strings.Builder

	data := Data{
		Messages: msgs,
	}

	err = d.templates.ExecuteTemplate(&buf, "gen.body", &data)
	if err != nil {
		return "", fmt.Errorf("execute template gen.body: %w", err)
	}

	return buf.String(), nil
}
