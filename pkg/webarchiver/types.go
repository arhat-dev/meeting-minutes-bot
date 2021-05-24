package webarchiver

type Interface interface {
	// Archive web page, return url of the archived page and screen shot
	Archive(url string) (
		archiveURL string,
		screenshot []byte,
		screenshotFileExt string,
		err error,
	)
}
