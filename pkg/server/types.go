package server

import (
	"fmt"
	"reflect"
	"strings"

	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/storage"
	"arhat.dev/mbot/pkg/webarchiver"
)

var (
	generatorResultType   = reflect.TypeOf((*generator.Output)(nil)).Elem()
	publisherResultType   = reflect.TypeOf((*publisher.Result)(nil)).Elem()
	storageResultType     = reflect.TypeOf((*storage.Result)(nil)).Elem()
	webarchiverResultType = reflect.TypeOf((*webarchiver.Result)(nil)).Elem()
)

type resultIfaceHandler struct{}

func (resultIfaceHandler) Create(typ reflect.Type, yamlKey string) (interface{}, error) {
	name, _, _ := strings.Cut(yamlKey, ":")
	switch typ {
	case generatorResultType:
		return generator.NewConfig(name)
	case publisherResultType:
		return publisher.NewConfig(name)
	case storageResultType:
		return storage.NewConfig(name)
	case webarchiverResultType:
		return webarchiver.NewConfig(name)
	default:
		return nil, fmt.Errorf("unknown result type %s (key: %s)", typ.String(), yamlKey)
	}
}
