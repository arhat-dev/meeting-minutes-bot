package webarchiver

import (
	"context"

	"github.com/chromedp/chromedp"

	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

var _ webarchiver.Interface = (*ChromeDevToolsProtocol)(nil)

type ChromeDevToolsProtocol struct {
}

func (a *ChromeDevToolsProtocol) Archive(
	url string,
) (
	archiveURL string,
	screenshot []byte,
	screenshotFileExt string,
	err error,
) {
	chromedp.NewContext(
		context.TODO(),
	)

	return "", nil, "", nil
}
