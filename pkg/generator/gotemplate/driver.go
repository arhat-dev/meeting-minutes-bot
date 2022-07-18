// Package gotemplate implements a generator to gererate content using golang template language
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

// Peek implements generator.Interface
func (*Driver) Peek(con rt.Conversation, msg *rt.Message) (out rt.GeneratorOutput, err error) {
	return
}

// New implements generator.Interface
func (d *Driver) New(con rt.Conversation, cmd, params string) (out rt.GeneratorOutput, err error) {
	var buf strings.Builder

	data := Data{
		Command: cmd,
		Params:  params,
	}

	err = d.templates.ExecuteTemplate(&buf, "gen.new", &data)
	if err != nil {
		err = fmt.Errorf("execute template gen.new: %w", err)
		return
	}

	out.Data.Set(buf.String())
	return
}

// Continue implements generator.Interface
func (d *Driver) Continue(con rt.Conversation, cmd, params string) (out rt.GeneratorOutput, err error) {
	var buf strings.Builder

	data := Data{
		Command: cmd,
		Params:  params,
	}

	err = d.templates.ExecuteTemplate(&buf, "gen.continue", &data)
	if err != nil {
		err = fmt.Errorf("execute template gen.continue: %w", err)
		return
	}

	out.Data.Set(buf.String())
	return
}

type Data struct {
	Command  string
	Params   string
	Messages []*rt.Message
}

// RenderBody implements generator.Interface
func (d *Driver) Generate(con rt.Conversation, cmd, params string, msgs []*rt.Message) (out rt.GeneratorOutput, err error) {
	var buf strings.Builder

	data := Data{
		Messages: msgs,
	}

	err = d.templates.ExecuteTemplate(&buf, "gen.body", &data)
	if err != nil {
		err = fmt.Errorf("execute template gen.body: %w", err)
		return
	}

	out.Data.Set(buf.String())
	return
}
