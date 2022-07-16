package generator

import (
	"fmt"

	"arhat.dev/meeting-minutes-bot/pkg/rt"
	"arhat.dev/pkg/textquery"
)

// FakeFuncMap creates a set of fake funcs with same function definitions as actual funcs
func FakeFuncMap() FuncMap {
	return map[string]any{
		"jq":          func(query string, data any) (string, error) { return "", nil },
		"findMessage": func(id uint64) *rt.Message { return nil },
	}
}

// CreateFuncMap creates actual funcs for templates
func CreateFuncMap(data *TemplateData) FuncMap {
	return map[string]any{
		"jq": func(query string, data any) (string, error) {
			switch t := data.(type) {
			case []byte:
				return textquery.JQ[byte](query, t)
			case string:
				return textquery.JQ[byte](query, t)
			default:
				return "", fmt.Errorf("unexpected non bytes nor string data %T", t)
			}
		},

		"findMessage": func(id rt.MessageID) *rt.Message {
			for _, m := range data.Messages {
				if m.ID == id {
					return m
				}
			}

			return nil
		},
	}
}
