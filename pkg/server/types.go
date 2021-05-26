package server

import (
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

type Client interface {
}

type Message interface {
	generator.Message

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

type generatorFactoryFunc func() (generator.Interface, generator.UserConfig, error)
