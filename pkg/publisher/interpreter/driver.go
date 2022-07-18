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
func (d *Driver) CreateNew(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{
				Flags: rt.SpanFlag_PlainText,
				Text:  fmt.Sprintf("You are using %s interpreter.", d.bin),
			},
		},
	})

	return
}

// AppendToExisting implements publisher.Interface
func (d *Driver) AppendToExisting(con rt.Conversation, cmd, params string, in *rt.GeneratorOutput) (out rt.PublisherOutput, err error) {
	var (
		args []string
		buf  bytes.Buffer
	)

	content := in.Data.Get()

	for i, tpl := range d.argTpls {
		buf.Reset()
		err = tpl.Execute(&buf, content)
		if err != nil {
			err = fmt.Errorf("execute #%d arg template: %w", i, err)
			return
		}
		args = append(args, buf.String())
	}

	execcmd := exec.CommandContext(con.Context(), d.bin, args...)

	output, err := execcmd.CombinedOutput()
	if err != nil {
		if len(output) != 0 {
			out.SendMessage.Set(rt.SendMessageOptions{
				Body: []rt.Span{
					{
						Flags: rt.SpanFlag_Pre,
						Text:  fmt.Sprintf("%s\n%v", output, err),
					},
				},
			})
			return
		}

		out.SendMessage.Set(rt.SendMessageOptions{
			Body: []rt.Span{
				{
					Flags: rt.SpanFlag_Pre,
					Text:  err.Error(),
				},
			},
		})
		return
	}

	out.SendMessage.Set(rt.SendMessageOptions{
		Body: []rt.Span{
			{
				Flags: rt.SpanFlag_Pre,
				Text:  string(output),
			},
		},
	})

	return
}

func (d *Driver) RequireLogin(con rt.Conversation, cmd, params string) (rt.LoginFlow, error) {
	return rt.LoginFlow_None, nil
}

func (d *Driver) Login(con rt.Conversation, user publisher.User) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) RequestExternalAccess(con rt.Conversation) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) Retrieve(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) List(con rt.Conversation) (out rt.PublisherOutput, err error) {
	return
}

func (d *Driver) Delete(con rt.Conversation, cmd, params string) (out rt.PublisherOutput, err error) {
	return
}
