package gotemplate

import (
	"testing"

	"arhat.dev/meeting-minutes-bot/pkg/rt"
	"github.com/stretchr/testify/assert"
)

func testMessages() []*rt.Message {
	return []*rt.Message{
		{ /* empty message */ },
		{ /* basic message */
			ID:       1,
			ChatName: "basic-chat-name",
			Author:   "basic-author-1",
			Spans: []rt.Span{
				{
					Flags: rt.SpanFlag_PlainText,
					Text:  "basic",
				},
			},
		},
		{ /* basic reply */
			ID:       2,
			ChatName: "basic-chat-name",
			Author:   "basic-author-2",
			Spans: []rt.Span{
				{
					Flags: rt.SpanFlag_Code,
					Text:  "reply-to-basic",
				},
			},
			IsReply:          true,
			ReplyToMessageID: 1,
		},
	}
}

func TestBuiltinTemplates(t *testing.T) {
	for _, test := range []struct {
		name   string
		config Config
	}{
		{
			name: "telegraph",
			config: Config{
				UseBuiltin: "telegraph",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			gen, err := test.config.Create()
			if !assert.NoError(t, err) {
				return
			}

			hdr, err := gen.RenderPageHeader()
			assert.NoError(t, err)
			t.Log("header", string(hdr))

			body, err := gen.RenderPageBody(testMessages())
			assert.NoError(t, err)
			t.Log("body", string(body))
		})
	}
}
