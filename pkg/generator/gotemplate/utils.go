package gotemplate

import (
	"fmt"
	htpl "html/template"
	"io"
	"io/fs"
	"path"
	ttpl "text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/bmatcuk/doublestar/v4"

	"arhat.dev/pkg/stringhelper"
)

type tplExecutor interface {
	ExecuteTemplate(wr io.Writer, name string, data *TemplateData) error
}

type hTemplate struct{ htpl.Template }

func (ht *hTemplate) ExecuteTemplate(wr io.Writer, name string, data *TemplateData) error {
	clone, err := ht.Template.Clone()
	if err != nil {
		return err
	}

	return clone.Funcs(realFuncMap(data)).ExecuteTemplate(wr, name, data)
}

type tTemplate struct{ ttpl.Template }

func (tt *tTemplate) ExecuteTemplate(wr io.Writer, name string, data *TemplateData) error {
	clone, err := tt.Template.Clone()
	if err != nil {
		return err
	}

	return clone.Funcs(realFuncMap(data)).ExecuteTemplate(wr, name, data)
}

func loadTemplatesFromFS(base any, dirFS fs.FS) (tplExecutor, error) {
	files, err := doublestar.Glob(dirFS, "**/*.tpl")
	if err != nil {
		return nil, fmt.Errorf("no template file found: %w", err)
	}

	switch t := base.(type) {
	case *htpl.Template:
		ht, err := parseFiles(
			t.Funcs(sprig.HtmlFuncMap()).Funcs(htpl.FuncMap(fakeFuncMap())),
			dirFS,
			files,
		)
		if err != nil {
			return nil, err
		}

		return &hTemplate{*ht}, nil
	case *ttpl.Template:
		tt, err := parseFiles(
			t.Funcs(sprig.TxtFuncMap()).Funcs(ttpl.FuncMap(fakeFuncMap())),
			dirFS,
			files,
		)
		if err != nil {
			return nil, err
		}

		return &tTemplate{*tt}, nil
	default:
		return nil, fmt.Errorf("unknown template type %T", base)
	}
}

func parseFiles[T interface {
	New(string) T
	Name() string
	Parse(string) (T, error)
}](t T, dirFS fs.FS, files []string) (_ T, err error) {
	if len(files) == 0 {
		err = fmt.Errorf("no template file provided")
		return
	}

	var (
		data []byte
		name string
		tmpl T
	)

	for _, filename := range files {
		data, err = fs.ReadFile(dirFS, filename)
		if err != nil {
			return
		}

		name = path.Base(filename)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.

		if name == t.Name() {
			tmpl = t
		} else {
			tmpl = t.New(name)
		}

		_, err = tmpl.Parse(stringhelper.Convert[string, byte](data))
		if err != nil {
			return
		}
	}

	return t, nil
}
