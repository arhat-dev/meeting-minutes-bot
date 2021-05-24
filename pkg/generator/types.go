package generator

type Interface interface {
	Name() string

	Publisher

	Formatter
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

	// Append content to local post cache
	Append(title string, body []byte) (url string, _ error)
}

type Formatter interface {
	Format(kind FormatKind, text string, params ...string) string
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
)

type UserConfig interface {
	SetAuthToken(token string)
}
