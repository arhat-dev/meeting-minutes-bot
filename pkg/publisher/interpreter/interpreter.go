package interpreter

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"text/template"

	"arhat.dev/pkg/textquery"
	"github.com/Masterminds/sprig/v3"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

// nolint:revive
const (
	Name = "interpreter"
)

func init() {
	publisher.Register(Name, func() publisher.Config { return &Config{} })
}

type Config struct {
	Bin  string   `json:"bin" yaml:"bin"`
	Args []string `json:"args" yaml:"args"`
}

func (c *Config) Create() (publisher.Interface, publisher.UserConfig, error) {
	var argTpls []*template.Template
	for _, arg := range c.Args {
		tpl, err := template.New("").
			Funcs(sprig.TxtFuncMap()).
			Funcs(map[string]interface{}{
				"jq":      textquery.JQ[byte, string],
				"jqBytes": textquery.JQ[byte, []byte],
			}).Parse(arg)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to parse arg template %q: %w", arg, err)
		}

		argTpls = append(argTpls, tpl)
	}

	return &Driver{
		bin:     c.Bin,
		argTpls: argTpls,
	}, &UserConfig{}, nil
}

var _ publisher.UserConfig = (*UserConfig)(nil)

type UserConfig struct{}

func (u *UserConfig) SetAuthToken(token string) {}

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	bin     string
	argTpls []*template.Template
}

func (d *Driver) Name() string {
	return Name
}

func (d *Driver) RequireLogin() bool {
	return false
}

func (d *Driver) Login(config publisher.UserConfig) (token string, _ error) {
	return "", fmt.Errorf("unimplemented")
}

func (d *Driver) AuthURL() (string, error) {
	return "", fmt.Errorf("unimplemented")
}

func (d *Driver) Retrieve(key string) ([]rt.Span, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) List() ([]publisher.PostInfo, error) {
	return nil, fmt.Errorf("unimplemented")
}

func (d *Driver) Delete(urls ...string) error {
	return fmt.Errorf("unimplemented")
}

func (d *Driver) Append(ctx context.Context, body *rt.Input) (_ []rt.Span, err error) {
	var (
		args []string
		buf  bytes.Buffer
	)

	content, err := body.String()
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

	cmd := exec.CommandContext(ctx, d.bin, args...)

	output, err := cmd.CombinedOutput()
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

func (d *Driver) Publish(title string, body *rt.Input) ([]rt.Span, error) {
	return []rt.Span{
		{
			Flags: rt.SpanFlag_PlainText,
			Text:  fmt.Sprintf("You are using %s interpreter.", d.bin),
		},
	}, nil
}
