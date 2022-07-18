package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/storage"
)

func newConfigIfaceHandler() *configIfaceHandler {
	return &configIfaceHandler{
		generatorConfigType: reflect.TypeOf((*generator.Config)(nil)).Elem(),
		publisherConfigType: reflect.TypeOf((*publisher.Config)(nil)).Elem(),
		storageConfigType:   reflect.TypeOf((*storage.Config)(nil)).Elem(),
		botConfigType:       reflect.TypeOf((*bot.Config)(nil)).Elem(),
	}
}

type configIfaceHandler struct {
	generatorConfigType reflect.Type
	publisherConfigType reflect.Type
	storageConfigType   reflect.Type
	botConfigType       reflect.Type
}

func (c *configIfaceHandler) Create(typ reflect.Type, yamlKey string) (interface{}, error) {
	name, _, _ := strings.Cut(yamlKey, ":")

	switch typ {
	case c.generatorConfigType:
		return generator.NewConfig(name)
	case c.publisherConfigType:
		return publisher.NewConfig(name)
	case c.storageConfigType:
		return storage.NewConfig(name)
	case c.botConfigType:
		return bot.NewConfig(name)
	default:
		return nil, fmt.Errorf("unknown config type %s (key: %s)", typ.String(), yamlKey)
	}
}
