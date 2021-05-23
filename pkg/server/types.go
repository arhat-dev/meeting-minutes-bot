package server

import (
	"arhat.dev/meeting-minutes-bot/pkg/fileuploader"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

type Client interface {
}

type Message interface {
	ID() string

	// preProcess the message:
	// - sound to text
	// - archive web pages
	// - download picture/video/audio/document
	PreProcess(
		c Client,
		w webarchiver.Interface,
		u fileuploader.Interface,
		previousMessages Message,
	) (chan error, error)

	// Ready returns true if the message has been pre-processed
	Ready() bool

	// Format message with target formatter
	Format(fm generator.Formatter) []byte
}
