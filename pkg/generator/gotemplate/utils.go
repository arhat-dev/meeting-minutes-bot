package gotemplate

import (
	"fmt"
	htmltemplate "html/template"
	"io"
	"io/fs"
	"path"
	texttemplate "text/template"

	"github.com/Masterminds/sprig/v3"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
)

type templateExecutor interface {
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

func loadTemplatesFromFS(base templateExecutor, dirFS fs.FS) (templateExecutor, error) {
	entries, err := fs.ReadDir(dirFS, ".")
	if err != nil {
		return nil, fmt.Errorf("load templates:: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("load templates: no dir in fs")
	}

	first := entries[0]
	dirName := first.Name()
	pattern := path.Join(dirName, "*")
	if !first.IsDir() {
		pattern = "*"
	}

	switch t := base.(type) {
	case *htmltemplate.Template:
		return t.
			Funcs(sprig.HtmlFuncMap()).
			Funcs(htmltemplate.FuncMap(generator.CreateFuncMap())).
			ParseFS(dirFS, pattern)
	case *texttemplate.Template:
		return t.
			Funcs(sprig.TxtFuncMap()).
			Funcs(texttemplate.FuncMap(generator.CreateFuncMap())).
			ParseFS(dirFS, pattern)
	default:
		return nil, fmt.Errorf("unexpected no base template: %T", base)
	}
}
