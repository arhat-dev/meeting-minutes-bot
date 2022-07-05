package gotemplate

import (
	"fmt"
	htmltemplate "html/template"
	"io"
	"io/fs"
	"path"
	texttemplate "text/template"

	"arhat.dev/pkg/textquery"
	"github.com/Masterminds/sprig/v3"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
)

type templateExecutor interface {
	ExecuteTemplate(wr io.Writer, name string, data interface{}) error
}

func loadTemplatesFromFS(base templateExecutor, dirFS fs.FS) (templateExecutor, error) {
	entries, err := fs.ReadDir(dirFS, ".")
	if err != nil {
		return nil, fmt.Errorf("load template: %w", err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("load template: no dir in fs")
	}

	first := entries[0]
	dirName := first.Name()
	pattern := path.Join(dirName, "**/*.tpl")
	if !first.IsDir() {
		pattern = "*"
	}

	switch t := base.(type) {
	case *htmltemplate.Template:
		return t.
			Funcs(sprig.HtmlFuncMap()).
			Funcs(htmltemplate.FuncMap(generator.CreateFuncMap())).
			Funcs(customFuncMap).
			ParseFS(dirFS, pattern)
	case *texttemplate.Template:
		return t.
			Funcs(sprig.TxtFuncMap()).
			Funcs(texttemplate.FuncMap(generator.CreateFuncMap())).
			Funcs(customFuncMap).
			ParseFS(dirFS, pattern)
	default:
		return nil, fmt.Errorf("unexpected no base template: %T", base)
	}
}

var customFuncMap = map[string]interface{}{
	"jq":      textquery.JQ[byte, string],
	"jqBytes": textquery.JQ[byte, []byte],
}
