// Package telegraph provides telegraph as a storage backend
package telegraph

import (
	"arhat.dev/mbot/pkg/storage"
)

const Name = "telegraph"

func init() {
	storage.Register(Name, func() storage.Config { return &Config{} })
}
