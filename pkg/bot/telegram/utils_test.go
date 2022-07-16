package telegram

import (
	"testing"

	"arhat.dev/mbot/pkg/rt"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/stretchr/testify/assert"
)

func TestParseTextEntities(t *testing.T) {
	for _, test := range []struct {
		name     string
		entities []styling.StyledTextOption

		expected []rt.Span
	}{
		{
			name:     "Simple",
			entities: []styling.StyledTextOption{styling.Plain("test")},
			expected: []rt.Span{
				{Text: "test", Flags: rt.SpanFlag_PlainText},
			},
		},
		{
			name: "Non-overlapped Styles",
			entities: []styling.StyledTextOption{
				styling.Plain("plainStart"),
				styling.Blockquote("blockquote"),
				styling.Bold("bold"),
				styling.Plain("plainEnd"),
			},
			expected: []rt.Span{
				{Text: "plainStart", Flags: rt.SpanFlag_PlainText},
				{Text: "blockquote", Flags: rt.SpanFlag_Blockquote},
				{Text: "bold", Flags: rt.SpanFlag_Bold},
				{Text: "plainEnd", Flags: rt.SpanFlag_PlainText},
			},
		},
		{
			name: "Overlapped Styles",
			entities: []styling.StyledTextOption{
				styling.Mention("mentionStart"),
				styling.Custom(func(eb *entity.Builder) error {
					eb.Format("italic-url-underline", entity.Italic(), entity.TextURL("foo"), entity.Underline())
					return nil
				}),
				styling.Strike("strikeEnd"),
			},
			expected: []rt.Span{
				{
					Text:  "mentionStart",
					Flags: rt.SpanFlag_Mention,
					URL:   "https://t.me/mentionStart",
				},
				{
					Text:  "italic-url-underline",
					Flags: rt.SpanFlag_Italic | rt.SpanFlag_URL | rt.SpanFlag_Underline,
					URL:   "foo",
				},
				{
					Text:  "strikeEnd",
					Flags: rt.SpanFlag_Strikethrough,
				},
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			var builder entity.Builder

			assert.NoError(t, styling.Perform(&builder, test.entities...))
			assert.EqualValues(t, test.expected, parseTextEntities(builder.Raw()))
		})
	}
}
