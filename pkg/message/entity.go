package message

// EntityKind of the text message
type EntityKind uint32

// Format kinds
const (
	KindUnknown EntityKind = 0

	// rich text and decorators
	KindPlainText EntityKind = 1 << (iota - 1)
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
	KindFile
)

type EntityParamKey string

// entity param keys: providing extra details about the entity
const (
	EntityParamURL                     EntityParamKey = "url"                        // url of the link
	EntityParamWebArchiveURL           EntityParamKey = "web_archive_url"            // web page archive url
	EntityParamWebArchiveScreenshotURL EntityParamKey = "web_archive_screenshot_url" // web page screenshot url
	EntityParamCaption                 EntityParamKey = "caption"                    // caption for photo/video/audio
	EntityParamFilename                EntityParamKey = "filename"                   // filename of a document
	EntityParamData                    EntityParamKey = "data"                       // data of the photo/video/audio
)

type EntityParams map[EntityParamKey]any

// func (ep EntityParams) GetData() {
// 	r, ok := ep[EntityParamData]
// 	if !ok {
// 		return
// 	}
// }

type Entity struct {
	Kind EntityKind
	Text string

	Params map[EntityParamKey]any
}

func (m Entity) IsPlainText() bool     { return m.Kind&KindPlainText != 0 }
func (m Entity) IsBold() bool          { return m.Kind&KindBold != 0 }
func (m Entity) IsItalic() bool        { return m.Kind&KindItalic != 0 }
func (m Entity) IsStrikethrough() bool { return m.Kind&KindStrikethrough != 0 }
func (m Entity) IsUnderline() bool     { return m.Kind&KindUnderline != 0 }
func (m Entity) IsPre() bool           { return m.Kind&KindPre != 0 }
func (m Entity) IsCode() bool          { return m.Kind&KindCode != 0 }
func (m Entity) IsBlockquote() bool    { return m.Kind&KindBlockquote != 0 }

func (m Entity) IsEmail() bool       { return m.Kind&KindEmail != 0 }
func (m Entity) IsPhoneNumber() bool { return m.Kind&KindPhoneNumber != 0 }
func (m Entity) IsURL() bool         { return m.Kind&KindURL != 0 }

func (m Entity) IsImage() bool { return m.Kind&KindImage != 0 }
func (m Entity) IsVideo() bool { return m.Kind&KindVideo != 0 }
func (m Entity) IsAudio() bool { return m.Kind&KindAudio != 0 }
func (m Entity) IsFile() bool  { return m.Kind&KindFile != 0 }
