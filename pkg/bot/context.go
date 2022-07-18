package bot

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/storage"
)

type CreationContext struct {
	Storage    map[string]storage.Interface
	Generators map[string]generator.Interface
	Publishers map[string]publisher.Config
}
