package fileuploader

type Interface interface {
	Login(config interface{}) error

	Upload(filename string, data []byte) (url string, err error)
}
