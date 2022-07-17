package mime

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

// New creates a new MIME struct by indexing first slash ('/') in value.
//
// if there is no slash in value, SlashIndex is set to len(value)
func New(value string) MIME {
	sep := strings.IndexByte(value, '/')
	if sep == -1 {
		sep = len(value)
	}

	return MIME{
		SlashIndex: sep,
		Value:      value,
	}
}

type MIME struct {
	SlashIndex int
	Value      string
}

func (m MIME) Type() string { return m.Value[:m.SlashIndex] }
func (m MIME) Subtype() string {
	if m.SlashIndex >= len(m.Value) {
		return ""
	}

	return m.Value[m.SlashIndex+1:]
}
