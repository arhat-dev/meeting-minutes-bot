package gotemplate

import (
	"bytes"
	"fmt"
	"os"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

// nolint:revive
const (
	Name = "gotemplate"
)

func init() {
	generator.Register(
		Name,
		func(config interface{}) (generator.Interface, error) {
			c, ok := config.(*Config)
			if !ok {
				return nil, fmt.Errorf("unexpected non gotemplate config: %T", config)
			}

			// TODO: move message template and page template to user config
			// 		 need to find a better UX for template editing

			var (
				err error
				tpl templateExecutor
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
				spec, ok := builtinTemplates[c.BuiltinTemplate]
				if !ok {
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
		},
		func() interface{} {
			return &Config{
				BuiltinTemplate: "",

				OutputFormat: "",
				TemplatesDir: "",
			}
		},
	)
}

type Config struct {
	BuiltinTemplate string `json:"builtinTemplate" yaml:"builtinTemplate"`

	// Custom template
	// text or html
	OutputFormat string `json:"outputFormat" yaml:"outputFormat"`
	TemplatesDir string `json:"templatesDir" yaml:"templatesDir"`
}

var _ generator.Interface = (*Driver)(nil)

type Driver struct {
	templates templateExecutor
}

func (g *Driver) Name() string {
	return Name
}

func (g *Driver) RenderPageHeader() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := g.templates.ExecuteTemplate(buf, "page.header", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to execute page header template: %w", err)
	}

	return buf.Bytes(), nil
}

func (g *Driver) RenderPageBody(
	messages []message.Interface,
) ([]byte, error) {
	var (
		buf = &bytes.Buffer{}
		err error
	)

	err = g.templates.ExecuteTemplate(
		buf,
		"page.body",
		&generator.TemplateData{
			Messages: messages,
		},
	)

	if err != nil {
		return nil, fmt.Errorf("failed to execute page template: %w", err)
	}

	return buf.Bytes(), nil
}
