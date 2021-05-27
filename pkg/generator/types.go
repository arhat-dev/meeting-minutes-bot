package generator

import (
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

type Interface interface {
	Name() string

	// FormatPageHeader render page.header
	FormatPageHeader() ([]byte, error)

	// FormatPageBody render page.body
	FormatPageBody(messages []message.Interface) ([]byte, error)
}

type TemplateData struct {
	Messages []message.Interface
}

type FuncMap map[string]interface{}
