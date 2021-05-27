package message

// EntityKind of the text message
type EntityKind uint16

// Format kinds
const (
	// text and decorators
	KindText EntityKind = iota + 1
	KindBold
	KindItalic
	KindStrikethrough
	KindUnderline
	KindPre
	KindCode
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

type EntityParamKey = string

// entity param keys
const (
	EntityParamURL                     EntityParamKey = "url"
	EntityParamWebArchiveURL           EntityParamKey = "web_archive_url"
	EntityParamWebArchiveScreenshotURL EntityParamKey = "web_archive_screenshot_url"
	EntityParamCaption                 EntityParamKey = "caption"
	EntityParamFilename                EntityParamKey = "filename"
)

type Entity struct {
	Kind   EntityKind
	Text   string
	Params map[EntityParamKey]interface{}
}

func (m Entity) IsText() bool          { return m.Kind == KindText }
func (m Entity) IsBold() bool          { return m.Kind == KindBold }
func (m Entity) IsItalic() bool        { return m.Kind == KindItalic }
func (m Entity) IsStrikethrough() bool { return m.Kind == KindStrikethrough }
func (m Entity) IsUnderline() bool     { return m.Kind == KindUnderline }
func (m Entity) IsPre() bool           { return m.Kind == KindPre }
func (m Entity) IsCode() bool          { return m.Kind == KindCode }
func (m Entity) IsBlockquote() bool    { return m.Kind == KindBlockquote }

func (m Entity) IsEmail() bool       { return m.Kind == KindEmail }
func (m Entity) IsPhoneNumber() bool { return m.Kind == KindPhoneNumber }
func (m Entity) IsURL() bool         { return m.Kind == KindURL }

func (m Entity) IsImage() bool    { return m.Kind == KindImage }
func (m Entity) IsVideo() bool    { return m.Kind == KindVideo }
func (m Entity) IsAudio() bool    { return m.Kind == KindAudio }
func (m Entity) IsDocument() bool { return m.Kind == KindDocument }
