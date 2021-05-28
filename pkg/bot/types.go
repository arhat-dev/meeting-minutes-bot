package bot

import (
	"net/http"

	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

// Mux is just a interface alternative to http.ServeMux
type Mux interface {
	HandleFunc(pattern string, handleFunc func(http.ResponseWriter, *http.Request))
}

type Interface interface {
	// Configure the bot (with bot api server)
	Configure() error

	// Start the bot in background (non-blocking)
	Start(baseURL string, mux Mux) error
}

type PublisherFactoryFunc func() (publisher.Interface, publisher.UserConfig, error)
