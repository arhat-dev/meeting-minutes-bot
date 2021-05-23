package fileuploader

var _ Interface = (*NopUploader)(nil)

type NopUploader struct {
}

func (u *NopUploader) Login(config interface{}) error {
	return nil
}

func (u *NopUploader) Upload(filename string, data []byte) (url string, err error) {
	return "", nil
}
