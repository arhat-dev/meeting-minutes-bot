package interpreter

import (
	"bytes"
	"fmt"
	"os/exec"
	"text/template"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	bin     string
	argTpls []*template.Template
}

// CreateNew implements publisher.Interface
func (d *Driver) CreateNew(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) ([]rt.Span, error) {
	return []rt.Span{
		{
			Flags: rt.SpanFlag_PlainText,
			Text:  fmt.Sprintf("You are using %s interpreter.", d.bin),
		},
	}, nil
}

// AppendToExisting implements publisher.Interface
func (d *Driver) AppendToExisting(con rt.Conversation, cmd, params string, fromGenerator *rt.Input) (_ []rt.Span, err error) {
	var (
		args []string
		buf  bytes.Buffer
	)

	content, err := fromGenerator.String()
	if err != nil {
		return
	}

	for i, tpl := range d.argTpls {
		buf.Reset()
		err := tpl.Execute(&buf, content)
		if err != nil {
			return nil, fmt.Errorf("failed to execute #%d arg template: %w", i, err)
		}
		args = append(args, buf.String())
	}

	execcmd := exec.CommandContext(con.Context(), d.bin, args...)

	output, err := execcmd.CombinedOutput()
	if err != nil {
		if len(output) != 0 {
			return []rt.Span{
				{
					Flags: rt.SpanFlag_Pre,
					Text:  fmt.Sprintf("%s\n%v", output, err),
				},
			}, nil
		}

		return []rt.Span{
			{
				Flags: rt.SpanFlag_Pre,
				Text:  err.Error(),
			},
		}, nil
	}

	return []rt.Span{
		{
			Flags: rt.SpanFlag_Pre,
			Text:  string(output),
		},
	}, nil
}

func (d *Driver) RequireLogin(con rt.Conversation, cmd, params string) (rt.LoginFlow, error) {
	return rt.LoginFlow_None, nil
}

func (d *Driver) Login(con rt.Conversation, user publisher.User) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) RequestExternalAccess(con rt.Conversation) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) Retrieve(con rt.Conversation, cmd, params string) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) List(con rt.Conversation) ([]publisher.PostInfo, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) Delete(con rt.Conversation, cmd, params string) error {
	return fmt.Errorf("unimplemented")
}
