package generator

import (
	"time"
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

type MessageEntity struct {
	Kind   FormatKind
	Text   string
	Params map[EntityParamKind]string
}

type Message interface {
	ID() string
	MessageURL() string

	Timestamp() time.Time

	ChatName() string
	ChatURL() string

	Author() string
	AuthorURL() string

	// message forwarded from other chat, use following info
	IsForwarded() bool
	OriginalChatName() string
	OriginalChatURL() string
	OriginalAuthor() string
	OriginalAuthorURL() string
	OriginalMessageURL() string

	IsPrivateMessage() bool

	IsReply() bool
	ReplyToMessageID() string

	Entities() []MessageEntity
}

type FuncMap map[string]interface{}

type Formatter interface {
	Name() string

	FormatPagePrefix() ([]byte, error)

	FormatPageContent(messages []Message, funcMap FuncMap) ([]byte, error)
}

// FormatKind of the text message
type FormatKind uint16

// Format kinds
const (
	// text and decorators
	KindText FormatKind = iota + 1
	KindBold
	KindItalic
	KindStrikethrough
	KindUnderline
	KindPre
	KindCode
	KindNewLine
	KindParagraph
	KindThematicBreak
	KindBlockquote

	// links
	KindEmail
	KindPhoneNumber
	KindURL

	// multi-media
	KindImage
	KindVideo
	KindAudio
	KindDocument
)

type EntityParamKind = string

// entity param keys
const (
	EntityParamURL                     EntityParamKind = "url"
	EntityParamWebArchiveURL           EntityParamKind = "web_archive_url"
	EntityParamWebArchiveScreenshotURL EntityParamKind = "web_archive_screenshot_url"
	EntityParamCaption                 EntityParamKind = "caption"
	EntityParamFilename                EntityParamKind = "filename"
)

type UserConfig interface {
	SetAuthToken(token string)
}

type MessageFindFunc func(id string) Message

func CreateFuncMap(findMessage MessageFindFunc) FuncMap {
	return map[string]interface{}{
		"entityIsText":          func(entity MessageEntity) bool { return entity.Kind == KindText },
		"entityIsBold":          func(entity MessageEntity) bool { return entity.Kind == KindBold },
		"entityIsItalic":        func(entity MessageEntity) bool { return entity.Kind == KindItalic },
		"entityIsStrikethrough": func(entity MessageEntity) bool { return entity.Kind == KindStrikethrough },
		"entityIsUnderline":     func(entity MessageEntity) bool { return entity.Kind == KindUnderline },
		"entityIsPre":           func(entity MessageEntity) bool { return entity.Kind == KindPre },
		"entityIsCode":          func(entity MessageEntity) bool { return entity.Kind == KindCode },
		"entityIsThematicBreak": func(entity MessageEntity) bool { return entity.Kind == KindThematicBreak },
		"entityIsBlockquote":    func(entity MessageEntity) bool { return entity.Kind == KindBlockquote },

		"entityIsEmail":       func(entity MessageEntity) bool { return entity.Kind == KindEmail },
		"entityIsPhoneNumber": func(entity MessageEntity) bool { return entity.Kind == KindPhoneNumber },
		"entityIsURL":         func(entity MessageEntity) bool { return entity.Kind == KindURL },

		"entityIsImage":    func(entity MessageEntity) bool { return entity.Kind == KindImage },
		"entityIsVideo":    func(entity MessageEntity) bool { return entity.Kind == KindVideo },
		"entityIsAudio":    func(entity MessageEntity) bool { return entity.Kind == KindAudio },
		"entityIsDocument": func(entity MessageEntity) bool { return entity.Kind == KindDocument },

		// to be overridden
		"findMessage": findMessage,
	}
}

type TemplateData struct {
	Messages []Message
}
