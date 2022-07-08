package cmd

import (
	"fmt"
	"reflect"
	"strings"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

var (
	generatorConfigType   = reflect.TypeOf((*generator.Config)(nil)).Elem()
	publisherConfigType   = reflect.TypeOf((*publisher.Config)(nil)).Elem()
	storageConfigType     = reflect.TypeOf((*storage.Config)(nil)).Elem()
	webarchiverConfigType = reflect.TypeOf((*webarchiver.Config)(nil)).Elem()
	botConfigType         = reflect.TypeOf((*bot.Config)(nil)).Elem()
)

type configIfaceHandler struct{}

func (configIfaceHandler) Create(typ reflect.Type, yamlKey string) (interface{}, error) {
	name, _, _ := strings.Cut(yamlKey, ":")

	switch typ {
	case generatorConfigType:
		return generator.NewConfig(name)
	case publisherConfigType:
		return publisher.NewConfig(name)
	case storageConfigType:
		return storage.NewConfig(name)
	case webarchiverConfigType:
		return webarchiver.NewConfig(name)
	case botConfigType:
		return bot.NewConfig(name)
	default:
		return nil, fmt.Errorf("unknown config type %s (key: %s)", typ.String(), yamlKey)
	}
}
