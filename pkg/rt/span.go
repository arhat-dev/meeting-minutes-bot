package rt

import (
	"time"
)

// SpanFlag of the message
type SpanFlag uint32

// Format kinds
const (
	SpanFlag_PlainText SpanFlag = 0

	// text styles
	SpanFlag_Bold SpanFlag = 1 << (iota - 1)
	SpanFlag_Italic
	SpanFlag_Strikethrough
	SpanFlag_Underline
	SpanFlag_Pre
	SpanFlag_Code
	SpanFlag_Blockquote

	// links
	SpanFlag_Email       // email address
	SpanFlag_PhoneNumber // phone number
	SpanFlag_URL         // (named) url
	SpanFlag_Mention     // @foo
	SpanFlag_HashTag     // #foo

	// media
	SpanFlag_Image // image
	SpanFlag_Video // video
	SpanFlag_Audio // non-voice chat audio
	SpanFlag_Voice // voice chat message
	SpanFlag_File  // arbitrary file
)

const (
	SpanFlagColl_StyledText = SpanFlag_Bold | SpanFlag_Italic | SpanFlag_Strikethrough |
		SpanFlag_Underline | SpanFlag_Pre | SpanFlag_Code | SpanFlag_Blockquote

	SpanFlagColl_Link = SpanFlag_Email | SpanFlag_PhoneNumber | SpanFlag_URL |
		SpanFlag_Mention | SpanFlag_HashTag

	SpanFlagColl_Media = SpanFlag_Image | SpanFlag_Video | SpanFlag_Audio |
		SpanFlag_Voice | SpanFlag_File
)

func (f SpanFlag) IsPlainText() bool { return f&SpanFlag_PlainText != 0 }

func (f SpanFlag) IsStyledText() bool    { return f&SpanFlagColl_StyledText != 0 }
func (f SpanFlag) IsBold() bool          { return f&SpanFlag_Bold != 0 }
func (f SpanFlag) IsItalic() bool        { return f&SpanFlag_Italic != 0 }
func (f SpanFlag) IsStrikethrough() bool { return f&SpanFlag_Strikethrough != 0 }
func (f SpanFlag) IsUnderline() bool     { return f&SpanFlag_Underline != 0 }
func (f SpanFlag) IsPre() bool           { return f&SpanFlag_Pre != 0 }
func (f SpanFlag) IsCode() bool          { return f&SpanFlag_Code != 0 }
func (f SpanFlag) IsBlockquote() bool    { return f&SpanFlag_Blockquote != 0 }

func (f SpanFlag) IsLink() bool        { return f&SpanFlagColl_Link != 0 }
func (f SpanFlag) IsEmail() bool       { return f&SpanFlag_Email != 0 }
func (f SpanFlag) IsPhoneNumber() bool { return f&SpanFlag_PhoneNumber != 0 }
func (f SpanFlag) IsURL() bool         { return f&SpanFlag_URL != 0 }
func (f SpanFlag) IsMention() bool     { return f&SpanFlag_Mention != 0 }
func (f SpanFlag) IsHashTag() bool     { return f&SpanFlag_HashTag != 0 }

func (f SpanFlag) IsMedia() bool { return f&SpanFlagColl_Media != 0 }
func (f SpanFlag) IsImage() bool { return f&SpanFlag_Image != 0 }
func (f SpanFlag) IsVideo() bool { return f&SpanFlag_Video != 0 }
func (f SpanFlag) IsAudio() bool { return f&SpanFlag_Audio != 0 }
func (f SpanFlag) IsVoice() bool { return f&SpanFlag_Voice != 0 }
func (f SpanFlag) IsFile() bool  { return f&SpanFlag_File != 0 }

// Span represents a standalone section in a message
type Span struct {
	Flags SpanFlag `yaml:"flags"`

	// Text is the text value visible to user
	Text string `yaml:"text"`

	// Hint
	// when kind is pre, it's the programming language name of the text
	// when kind is mention, it's the value of mentioned name
	// when kind is audio, it's the title of the music
	Hint string `yaml:"hint"`

	// URL
	// when kind is link, it's the url
	// when kind is media, it's the url to the uploaded content
	URL string `yaml:"url"`

	// WebArchiveURL for archived web page for kind URL
	WebArchiveURL string `yaml:"webarchiveURL"`

	// WebArchiveScreenshotURL for screenshot of archived web page for kind URL
	WebArchiveScreenshotURL string `yaml:"webarchiveScreenshotURL"`

	SpanMediaOptions `yaml:",inline"`
}

type SpanMediaOptions struct {
	// Caption for media
	Caption []Span `yaml:"caption"`

	// Filename for media
	//
	// Optional
	Filename string `yaml:"filename"`

	// Data of media
	//
	// REQUIRED
	Data CacheReader `yaml:"cacheID"`

	// Size of Data
	//
	// REQUIRED
	Size int64 `yaml:"size"`

	// ContentType of Data
	//
	// REQUIRED, when it's unknown, set it to "application/octet-stream"
	ContentType string `yaml:"contentType"`

	// Duration of Data (video/audio/voice)
	Duration time.Duration `yaml:"duration"`
}

func (f *Span) IsPlainText() bool { return f.Flags.IsPlainText() }

func (f *Span) IsStyledText() bool    { return f.Flags.IsStyledText() }
func (f *Span) IsBold() bool          { return f.Flags.IsBold() }
func (f *Span) IsItalic() bool        { return f.Flags.IsItalic() }
func (f *Span) IsStrikethrough() bool { return f.Flags.IsStrikethrough() }
func (f *Span) IsUnderline() bool     { return f.Flags.IsUnderline() }
func (f *Span) IsPre() bool           { return f.Flags.IsPre() }
func (f *Span) IsCode() bool          { return f.Flags.IsCode() }
func (f *Span) IsBlockquote() bool    { return f.Flags.IsBlockquote() }

func (f *Span) IsLink() bool        { return f.Flags.IsLink() }
func (f *Span) IsEmail() bool       { return f.Flags.IsEmail() }
func (f *Span) IsPhoneNumber() bool { return f.Flags.IsPhoneNumber() }
func (f *Span) IsURL() bool         { return f.Flags.IsURL() }
func (f *Span) IsMention() bool     { return f.Flags.IsMention() }

func (f *Span) IsMedia() bool { return f.Flags.IsMedia() }
func (f *Span) IsImage() bool { return f.Flags.IsImage() }
func (f *Span) IsVideo() bool { return f.Flags.IsVideo() }
func (f *Span) IsAudio() bool { return f.Flags.IsAudio() }
func (f *Span) IsVoice() bool { return f.Flags.IsVoice() }
func (f *Span) IsFile() bool  { return f.Flags.IsFile() }
