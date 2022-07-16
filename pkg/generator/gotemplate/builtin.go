package gotemplate

import (
	"embed"
	htpl "html/template"
	"io/fs"
	ttpl "text/template"
)

func newHTMLTemplate() any { return htpl.New("") }
func newTextTemplate() any { return ttpl.New("") }

type templateLoadSpec struct {
	name    string
	fs      fs.FS
	factory func() any
}

var (
	//go:embed templates/text
	builtinTextTemplate embed.FS

	//go:embed templates/beancount
	builtinBeancountTemplate embed.FS

	//go:embed templates/telegraph
	builtinTelegraphTemplate embed.FS

	//go:embed templates/http-req-spec
	builtinHTTPRequestSpecTemplate embed.FS
)

const (
	builtinTpl_Text = iota
	builtinTpl_Beancount
	builtinTpl_Telegraph
	builtinTpl_HttpRequestSpec
)

var builtinTemplates = [...]templateLoadSpec{
	builtinTpl_Text: {
		name:    "text",
		fs:      &builtinTextTemplate,
		factory: newTextTemplate,
	},
	builtinTpl_Beancount: {
		name:    "beancount",
		fs:      &builtinBeancountTemplate,
		factory: newTextTemplate,
	},
	builtinTpl_Telegraph: {
		name:    "telegraph",
		fs:      &builtinTelegraphTemplate,
		factory: newHTMLTemplate,
	},
	builtinTpl_HttpRequestSpec: {
		name:    "http-req-spec",
		fs:      &builtinHTTPRequestSpecTemplate,
		factory: newTextTemplate,
	},
}
