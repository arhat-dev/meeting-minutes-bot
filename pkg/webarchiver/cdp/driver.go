package webarchiver

import (
	"github.com/chromedp/chromedp"

	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/webarchiver"
)

var _ webarchiver.Interface = (*Driver)(nil)

type Driver struct{}

func (d *Driver) Archive(con rt.Conversation, url string) (_ webarchiver.Result, _ error) {
	chromedp.NewContext(con.Context())
	return
}
