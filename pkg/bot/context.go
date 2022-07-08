package bot

import (
	"fmt"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
	"arhat.dev/rs"
)

type Context struct {
	rs.BaseField

	StorageSets  map[string]storageSet         `yaml:"storageSets"`
	WebArchivers map[string]webarchiver.Config `yaml:"webarchivers"`
	Generators   map[string]generator.Config   `yaml:"generators"`
	Publishers   map[string]publisher.Config   `yaml:"publishers"`
}

type (
	storageSet   []storageEntry
	storageEntry struct {
		rs.BaseField

		Config map[string]storage.Config `yaml:",inline"`
	}
)

func (ss storageSet) Create() (_ storage.Interface, err error) {
	storageMgr := storage.NewManager()
	for _, st := range ss {
		for k, cfg := range st.Config {
			err = storageMgr.Add(cfg.MIMEMatch(), cfg.MaxSize(), cfg)
			if err != nil {
				err = fmt.Errorf("add storage driver %q: %w", k, err)
				return
			}
		}
	}

	return storageMgr, nil
}
