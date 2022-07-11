package message

import (
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

// SpanFlags of the message
type SpanFlags uint32

// Format kinds
const (
	SpanFlags_Unknown SpanFlags = 0

	SpanFlags_PlainText SpanFlags = 1 << (iota - 1)

	// text styles
	SpanFlags_Bold
	SpanFlags_Italic
	SpanFlags_Strikethrough
	SpanFlags_Underline
	SpanFlags_Pre
	SpanFlags_Code
	SpanFlags_Blockquote

	// links
	SpanFlags_Email       // email address
	SpanFlags_PhoneNumber // phone number
	SpanFlags_URL         // (named) url
	SpanFlags_Mention     // @foo
	SpanFlags_HashTag     // #foo

	// multi-media
	SpanFlags_Image // image
	SpanFlags_Video // video
	SpanFlags_Audio // non-voice chat audio
	SpanFlags_Voice // voice chat message
	SpanFlags_File  // arbitrary file
)

const (
	SpanFlagsColl_StyledText = SpanFlags_Bold | SpanFlags_Italic | SpanFlags_Strikethrough |
		SpanFlags_Underline | SpanFlags_Pre | SpanFlags_Code | SpanFlags_Blockquote

	SpanFlagsColl_Link = SpanFlags_Email | SpanFlags_PhoneNumber | SpanFlags_URL |
		SpanFlags_Mention | SpanFlags_HashTag

	SpanFlagsColl_MultiMedia = SpanFlags_Image | SpanFlags_Video | SpanFlags_Audio |
		SpanFlags_Voice | SpanFlags_File
)

// Span represents a standalone section in a message
type Span struct {
	SpanFlags

	// Text is the plain text of the message
	Text string

	// URL
	// when kind is KindURL, it's the original url, text is the decoded url
	// when kind is multi-media, it's the url to the uploaded content
	URL rt.Optional[string]

	// WebArchiveURL for archived web page
	WebArchiveURL rt.Optional[string]

	// WebArchiveScreenshotURL for screenshot of archived web page
	WebArchiveScreenshotURL rt.Optional[string]

	// Caption for non-text messages
	Caption rt.Optional[string]

	// Filename for multi-media message
	Filename rt.Optional[string]

	// Data of non-text message
	//
	// REQUIRED when it's a multi-media Span
	Data rt.CacheReader

	// Size of Data
	//
	// REQUIRED when it's a multi-media Span
	Size int64

	// ContentType of Data
	//
	// REQUIRED when it's a multi-media Span
	ContentType string

	// Duration of Data (video/audio/voice)
	Duration rt.Optional[time.Duration]
}

func (f SpanFlags) IsPlainText() bool { return f&SpanFlags_PlainText != 0 }

func (f SpanFlags) IsStyledText() bool    { return f&SpanFlagsColl_StyledText != 0 }
func (f SpanFlags) IsBold() bool          { return f&SpanFlags_Bold != 0 }
func (f SpanFlags) IsItalic() bool        { return f&SpanFlags_Italic != 0 }
func (f SpanFlags) IsStrikethrough() bool { return f&SpanFlags_Strikethrough != 0 }
func (f SpanFlags) IsUnderline() bool     { return f&SpanFlags_Underline != 0 }
func (f SpanFlags) IsPre() bool           { return f&SpanFlags_Pre != 0 }
func (f SpanFlags) IsCode() bool          { return f&SpanFlags_Code != 0 }
func (f SpanFlags) IsBlockquote() bool    { return f&SpanFlags_Blockquote != 0 }

func (f SpanFlags) IsLink() bool        { return f&SpanFlagsColl_Link != 0 }
func (f SpanFlags) IsEmail() bool       { return f&SpanFlags_Email != 0 }
func (f SpanFlags) IsPhoneNumber() bool { return f&SpanFlags_PhoneNumber != 0 }
func (f SpanFlags) IsURL() bool         { return f&SpanFlags_URL != 0 }
func (f SpanFlags) IsMention() bool     { return f&SpanFlags_Mention != 0 }

func (f SpanFlags) IsMultiMedia() bool { return f&SpanFlagsColl_MultiMedia != 0 }
func (f SpanFlags) IsImage() bool      { return f&SpanFlags_Image != 0 }
func (f SpanFlags) IsVideo() bool      { return f&SpanFlags_Video != 0 }
func (f SpanFlags) IsAudio() bool      { return f&SpanFlags_Audio != 0 }
func (f SpanFlags) IsVoice() bool      { return f&SpanFlags_Voice != 0 }
func (f SpanFlags) IsFile() bool       { return f&SpanFlags_File != 0 }

type Entities []Span

// NeedPreProcess returns true when any of following is true:
// - there is url to be archived
func (me Entities) NeedPreProcess() bool {
	for i := range me {
		if !me[i].URL.IsNil() {
			return true
		}
	}

	return false
}
