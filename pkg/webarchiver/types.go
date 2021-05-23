package webarchiver

type Interface interface {
	Login(config interface{}) error

	Archive(url string) (archiveURL string, err error)
}
