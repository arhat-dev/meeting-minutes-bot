package rt

import "strings"

const (
	// discrete types
	MIMEType_Application = "application"
	MIMEType_Image       = "image"
	MIMEType_Video       = "video"
	MIMEType_Audio       = "auido"
	MIMEType_Text        = "text"

	// composite types
	MIMEType_Message   = "message"
	MIMEType_Multipart = "multipart"
)

func NewMIME(value string) MIME {
	sep := strings.IndexByte(value, '/')
	if sep == -1 {
		sep = len(value)
	}

	return MIME{
		sep:   sep,
		value: value,
	}
}

type MIME struct {
	sep   int
	value string
}

func (m MIME) Type() string    { return m.value[:m.sep] }
func (m MIME) Subtype() string { return strings.TrimPrefix(m.value[m.sep:], "/") }
func (m MIME) Value() string   { return m.value }
