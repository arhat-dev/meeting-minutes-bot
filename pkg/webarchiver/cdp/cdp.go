package webarchiver

import (
	"context"

	"github.com/chromedp/chromedp"

	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

var _ webarchiver.Interface = (*Driver)(nil)

type Driver struct{}

func (d *Driver) Archive(ctx context.Context, url string) (webarchiver.Result, error) {
	chromedp.NewContext(ctx)
	return nil, nil
}
