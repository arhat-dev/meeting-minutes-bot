package rshelper

import (
	"bytes"
	"fmt"
	"text/template"

	"arhat.dev/rs"

	"arhat.dev/pkg/yamlhelper"
)

// TemplateHandler execute raw data as text/template
type TemplateHandler struct {
	CreateFuncMap func() template.FuncMap
}

func (h *TemplateHandler) RenderYaml(
	_ string, rawData interface{},
) ([]byte, error) {
	rawData, err := rs.NormalizeRawData(rawData)
	if err != nil {
		return nil, err
	}

	tplBytes, err := yamlhelper.ToYamlBytes(rawData)
	if err != nil {
		return nil, fmt.Errorf("failed to get data bytes of input: %w", err)
	}

	t := template.New("")
	if h.CreateFuncMap != nil {
		t = t.Funcs(h.CreateFuncMap())
	}

	t, err = t.Parse(string(tplBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %q: %w", string(tplBytes), err)
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, nil)

	return buf.Next(buf.Len()), err
}
