package webarchiver

var _ Interface = (*NopArchiver)(nil)

type NopArchiver struct {
}

func (a *NopArchiver) Login(config interface{}) error {
	return nil
}

func (a *NopArchiver) Archive(url string) (archiveURL string, err error) {
	return "", nil
}
