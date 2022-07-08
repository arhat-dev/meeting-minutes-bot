// Package telegraph provides telegraph as a storage backend
package telegraph

import (
	"arhat.dev/meeting-minutes-bot/pkg/storage"
)

const Name = "telegraph"

func init() {
	storage.Register(
		Name,
		func() storage.Config { return &Config{} },
	)
}
