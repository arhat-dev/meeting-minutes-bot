package generator

import (
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

type Config interface {
	// Create a generation based on this config
	Create() (Interface, error)
}

type Interface interface {
	Name() string

	// RenderPageHeader render page.header
	RenderPageHeader() ([]byte, error)

	// RenderPageBody render page.body
	RenderPageBody(messages []message.Interface) ([]byte, error)
}

type TemplateData struct {
	Messages []message.Interface
}

type FuncMap map[string]interface{}
