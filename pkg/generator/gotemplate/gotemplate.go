package gotemplate

import (
	"bytes"
	"fmt"
	"os"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/rs"
)

// nolint:revive
const (
	Name = "gotemplate"
)

func init() {
	generator.Register(
		Name,
		func() generator.Config { return &Config{} },
	)
}

type Config struct {
	rs.BaseField

	BuiltinTemplate string `yaml:"builtinTemplate"`

	// Custom template
	// text or html
	OutputFormat string `yaml:"outputFormat"`
	TemplatesDir string `yaml:"templatesDir"`
}

func (c *Config) Create() (generator.Interface, error) {
	// TODO: move message template and page template to user config
	// 		 need to find a better UX for template editing

	var (
		err error
		tpl tplExecutor
	)

	switch {
	case len(c.TemplatesDir) != 0:
		switch f := c.OutputFormat; f {
		case "text":
			tpl = newTextTemplate()
		case "html":
			tpl = newHTMLTemplate()
		default:
			return nil, fmt.Errorf("unknown output format: %q", f)
		}

		tpl, err = loadTemplatesFromFS(tpl, os.DirFS(c.TemplatesDir))
		if err != nil {
			return nil, fmt.Errorf("failed to load custom templates: %w", err)
		}
	case len(c.BuiltinTemplate) != 0:
		var spec templateLoadSpec

		switch c.BuiltinTemplate {
		case "text":
			spec = builtinTemplates[builtinTpl_Text]
		case "telegraph":
			spec = builtinTemplates[builtinTpl_Telegraph]
		case "beancount":
			spec = builtinTemplates[builtinTpl_Beancount]
		case "http-request-spec":
			spec = builtinTemplates[builtinTpl_HttpRequestSpec]
		default:
			return nil, fmt.Errorf("no such builtin template with name %q", c.BuiltinTemplate)
		}

		tpl, err = loadTemplatesFromFS(spec.factory(), spec.fs)
		if err != nil {
			return nil, fmt.Errorf("failed to load custom templates: %w", err)
		}
	default:
		return nil, fmt.Errorf("no template specified")
	}

	return &Driver{
		templates: tpl,
	}, nil
}

var _ generator.Interface = (*Driver)(nil)

type Driver struct {
	templates tplExecutor
}

func (g *Driver) Name() string {
	return Name
}

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

func (g *Driver) RenderPageBody(
	messages []message.Interface,
) (_ []byte, err error) {
	var (
		buf bytes.Buffer
	)

	err = g.templates.ExecuteTemplate(
		&buf,
		"page.body",
		&generator.TemplateData{
			Messages: messages,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to execute page template: %w", err)
	}

	return buf.Next(buf.Len()), nil
}
