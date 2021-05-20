package generator

type Interface interface {
	Login(config interface{}) error

	Retrieve(url string) (title string, _ error)

	Publish(title string, htmlContent []byte) (url string, _ error)

	Append(title string, htmlContent []byte) (url string, _ error)
}
