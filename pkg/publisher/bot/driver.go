package bot

import (
	"context"

	"arhat.dev/mbot/pkg/publisher"
	"arhat.dev/mbot/pkg/rt"
)

var _ publisher.Interface = (*Driver)(nil)

type Driver struct {
}

// RequireLogin return true when the publisher requires login, if false
// there will be no login process presented to user
func (d Driver) RequireLogin() bool { return false }

// Login to platform
func (d *Driver) Login(config publisher.UserConfig) (token string, err error) { return }

// AuthURL return a one click url for external authorization
func (d *Driver) AuthURL() (_ string, err error) { return }

// Retrieve post and cache it locally according to the url
func (d *Driver) Retrieve(url string) (_ []rt.Span, err error) { return }

// Publish a new post
func (d *Driver) Publish(title string, body *rt.Input) (_ []rt.Span, err error) { return }

// List all posts for this user
func (d *Driver) List() (_ []publisher.PostInfo, err error) { return }

// Delete one post according to the url
func (d *Driver) Delete(urls ...string) (err error) { return }

// Append content to local post cache
func (d *Driver) Append(ctx context.Context, body *rt.Input) (_ []rt.Span, err error) { return }
