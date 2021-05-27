package generator

import (
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

type Interface interface {
	Name() string

	Publisher

	Formatter
}

type PostInfo struct {
	Title string
	URL   string
}

type Publisher interface {
	// Login to platform
	Login(config UserConfig) (token string, _ error)

	// AuthURL return a one click url for external authorization
	AuthURL() (string, error)

	// Retrieve post and cache it locally according to the url
	Retrieve(url string) (title string, _ error)

	// Publish a new post
	Publish(title string, body []byte) (url string, _ error)

	// List all posts for this user
	List() ([]PostInfo, error)

	// Delete one post according to the url
	Delete(urls ...string) error

	// Append content to local post cache
	Append(title string, body []byte) (url string, _ error)
}

type FuncMap map[string]interface{}

type UserConfig interface {
	SetAuthToken(token string)
}

type Formatter interface {
	Name() string

	// FormatPageHeader render page.header
	FormatPageHeader() ([]byte, error)

	// FormatPageBody render page.body
	FormatPageBody(messages []message.Interface) ([]byte, error)
}

func CreateFuncMap() FuncMap {
	return map[string]interface{}{
		"findMessage": func(messages []message.Interface, id string) message.Interface {
			// TODO: use index?
			for _, m := range messages {
				if m.ID() == id {
					return m
				}
			}

			return nil
		},
	}
}

type TemplateData struct {
	Messages []message.Interface
}
