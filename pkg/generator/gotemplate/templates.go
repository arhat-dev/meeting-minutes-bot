package gotemplate

import (
	"embed"
	"fmt"
	htmltemplate "html/template"
	"io/fs"
	texttemplate "text/template"
)

func newHTMLTemplate() templateExecutor {
	return htmltemplate.New("")
}

func newTextTemplate() templateExecutor {
	return texttemplate.New("")
}

type templateLoadSpec struct {
	name    string
	fs      fs.FS
	factory func() templateExecutor
}

var (
	//go:embed templates/text/*
	builtinTextTemplate embed.FS

	//go:embed templates/beancount/*
	builtinBeancountTemplate embed.FS

	//go:embed templates/telegraph/*
	builtinTelegraphTemplate embed.FS
)

var builtinTemplates = map[string]templateLoadSpec{
	"text": {
		name:    "text",
		fs:      &builtinTextTemplate,
		factory: newTextTemplate,
	},
	"beancount": {
		name:    "beancount",
		fs:      &builtinBeancountTemplate,
		factory: newTextTemplate,
	},
	"telegraph": {
		name:    "telegraph",
		fs:      &builtinTelegraphTemplate,
		factory: newHTMLTemplate,
	},
}

func init() {
	for _, spec := range builtinTemplates {
		_, err := loadTemplatesFromFS(spec.factory(), spec.fs)
		if err != nil {
			panic(fmt.Errorf("default %q templates not valid: %w", spec.name, err))
		}
	}
}
