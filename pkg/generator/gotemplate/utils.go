package gotemplate

import (
	"fmt"
	htpl "html/template"
	"io"
	"io/fs"
	"path"
	ttpl "text/template"

	"github.com/Masterminds/sprig/v3"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
)

type tplExecutor interface {
	ExecuteTemplate(wr io.Writer, name string, data *generator.TemplateData) error
}

type hTemplate struct{ htpl.Template }

func (ht *hTemplate) ExecuteTemplate(wr io.Writer, name string, data *generator.TemplateData) error {
	clone, err := ht.Template.Clone()
	if err != nil {
		return err
	}

	return clone.Funcs(generator.CreateFuncMap(data)).ExecuteTemplate(wr, name, data)
}

type tTemplate struct{ ttpl.Template }

func (tt *tTemplate) ExecuteTemplate(wr io.Writer, name string, data *generator.TemplateData) error {
	clone, err := tt.Template.Clone()
	if err != nil {
		return err
	}

	return clone.Funcs(generator.CreateFuncMap(data)).ExecuteTemplate(wr, name, data)
}

func loadTemplatesFromFS(base any, dirFS fs.FS) (tplExecutor, error) {
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
	case *htpl.Template:
		ht, err := t.Funcs(sprig.HtmlFuncMap()).
			Funcs(htpl.FuncMap(generator.FakeFuncMap())).
			ParseFS(dirFS, pattern)
		if err != nil {
			return nil, err
		}

		return &hTemplate{*ht}, nil
	case *ttpl.Template:
		tt, err := t.Funcs(sprig.TxtFuncMap()).
			Funcs(ttpl.FuncMap(generator.FakeFuncMap())).
			ParseFS(dirFS, pattern)
		if err != nil {
			return nil, err
		}

		return &tTemplate{*tt}, nil
	default:
		return nil, fmt.Errorf("unknown template type %T", base)
	}
}
