package webarchiver

var _ Interface = (*Nop)(nil)

type NopConfig struct{}

type Nop struct{}

func (a *Nop) Login(config interface{}) error {
	return nil
}

func (a *Nop) Archive(
	url string,
) (
	archiveURL string,
	screenshot []byte,
	screenshotFileExt string,
	err error,
) {
	return "", nil, "", nil
}
