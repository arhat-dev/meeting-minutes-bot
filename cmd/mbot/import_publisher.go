package main

// data publishing drivers
import (
	_ "arhat.dev/mbot/pkg/publisher/file"
	_ "arhat.dev/mbot/pkg/publisher/http"
	_ "arhat.dev/mbot/pkg/publisher/interpreter"
	_ "arhat.dev/mbot/pkg/publisher/multipub"
	_ "arhat.dev/mbot/pkg/publisher/telegraph"
)
