package gotemplate

import (
	"embed"
	"fmt"
	htmltemplate "html/template"
	"io/fs"
	texttemplate "text/template"
)

var (
	//go:embed telegraph-templates/*
	defaultTelegraphTemplates embed.FS
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
	defaultTemplates = map[string]templateLoadSpec{
		"telegraph": {
			name:    "telegraph",
			fs:      &defaultTelegraphTemplates,
			factory: newHTMLTemplate,
		},
	}
)

func init() {
	for _, spec := range defaultTemplates {
		_, err := loadTemplatesFromFS(spec.factory(), spec.fs)
		if err != nil {
			panic(fmt.Errorf("default %q templates not valid: %w", spec.name, err))
		}
	}
}
