// Package authorized implements a wrapper publisher providing authorized user for actual publisher
package authorized

import (
	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
	publisher.Interface
}

// RequireLogin overrides publisher.Interface
func (*Driver) RequireLogin(con rt.Conversation, cmd, params string, user publisher.User) (out rt.PublisherOutput, err error) {
	// nop
	return
}
