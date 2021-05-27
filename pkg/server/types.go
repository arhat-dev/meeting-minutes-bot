package server

import (
	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

type Client interface {
}

type Message interface {
	message.Interface

	// preProcess the message:
	// - sound to text
	// - archive web pages
	// - download picture/video/audio/document
	PreProcess(
		c Client,
		w webarchiver.Interface,
		u storage.Interface,
		previousMessages Message,
	) (chan error, error)

	// Ready returns true if the message has been pre-processed
	Ready() bool
}

type publisherFactoryFunc func() (publisher.Interface, publisher.UserConfig, error)
