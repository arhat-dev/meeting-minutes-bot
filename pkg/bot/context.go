package bot

import (
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/storage"
	"arhat.dev/mbot/pkg/webarchiver"
	"arhat.dev/rs"
)

type Context struct {
	rs.BaseField

	StorageSets  map[string]storage.Config     `yaml:"storage"`
	WebArchivers map[string]webarchiver.Config `yaml:"webarchivers"`
	Generators   map[string]generator.Config   `yaml:"generators"`
	Publishers   map[string]publisher.Config   `yaml:"publishers"`
}
