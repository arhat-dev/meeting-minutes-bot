package main

// storage drivers
import (
	_ "arhat.dev/mbot/pkg/storage/router"
	_ "arhat.dev/mbot/pkg/storage/s3"
	_ "arhat.dev/mbot/pkg/storage/telegraph"
)
