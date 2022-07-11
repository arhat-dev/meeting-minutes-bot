package gotemplate

import (
	"fmt"
	"os"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/rs"
)

func init() {
	generator.Register(
		Name,
		func() generator.Config { return &Config{} },
	)
}

type Config struct {
	rs.BaseField

	UseBuiltin string `yaml:"useBuiltin"`

	// Custom template

	// text or html
	OutputFormat string `yaml:"outputFormat"`

	// TemplatesDir is the dir template files stored
	TemplatesDir string `yaml:"templatesDir"`
}

func (c *Config) Create() (generator.Interface, error) {
	var (
		err error
		tpl tplExecutor
	)

	switch {
	case len(c.TemplatesDir) != 0:
		var base any

		switch f := c.OutputFormat; f {
		case "text":
			base = newTextTemplate()
		case "html":
			base = newHTMLTemplate()
		default:
			return nil, fmt.Errorf("unknown output format %q", f)
		}

		tpl, err = loadTemplatesFromFS(base, os.DirFS(c.TemplatesDir))
		if err != nil {
			return nil, fmt.Errorf("failed to load custom template from %q: %w", c.TemplatesDir, err)
		}
	case len(c.UseBuiltin) != 0:
		var spec templateLoadSpec

		switch c.UseBuiltin {
		case "text":
			spec = builtinTemplates[builtinTpl_Text]
		case "telegraph":
			spec = builtinTemplates[builtinTpl_Telegraph]
		case "beancount":
			spec = builtinTemplates[builtinTpl_Beancount]
		case "http-request-spec":
			spec = builtinTemplates[builtinTpl_HttpRequestSpec]
		default:
			return nil, fmt.Errorf("no such builtin template with name %q", c.UseBuiltin)
		}

		tpl, err = loadTemplatesFromFS(spec.factory(), spec.fs)
		if err != nil {
			return nil, fmt.Errorf("failed to load builtin template %q: %w", c.UseBuiltin, err)
		}
	default:
		return nil, fmt.Errorf("no template specified")
	}

	return &Driver{
		templates: tpl,
	}, nil
}
