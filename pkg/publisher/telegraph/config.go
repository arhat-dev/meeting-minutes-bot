package telegraph

import (
	"arhat.dev/meeting-minutes-bot/pkg/publisher"
)

var _ publisher.UserConfig = (*userConfig)(nil)

type userConfig struct {
	shortName  string
	authorName string
	authorURL  string

	authToken string
}

func (c *userConfig) SetAuthToken(token string) {
	c.authToken = token
}
