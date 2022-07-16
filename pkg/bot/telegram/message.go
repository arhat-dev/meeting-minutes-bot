package telegram

import (
	"strconv"
	"strings"
	"time"

	"arhat.dev/mbot/pkg/rt"
	"github.com/gotd/td/tg"
)

func newTelegramMessage(src *messageSource, msg *tg.Message, msgs *[]*rt.Message) (ret rt.Message) {
	var (
		buf strings.Builder
	)

	ret.ID = rt.MessageID(msg.GetID())
	// TODO: set tz by user location
	ret.Timestamp = time.Unix(int64(msg.GetDate()), 0).UTC()

	ret.IsPrivateMessage = src.Chat.IsPrivateChat()

	if replyTo, ok := msg.GetReplyTo(); ok {
		ret.IsReply = true
		ret.ReplyToMessageID = rt.MessageID(replyTo.GetReplyToMsgID())
	}

	if !src.FwdChat.IsNil() {
		ret.IsForwarded = true

		fwdChat := src.FwdChat.GetPtr()

		ret.OriginalChatName = fwdChat.Title()
		if len(ret.OriginalChatName) == 0 {
			buf.Reset()
			buf.WriteString(fwdChat.Firstname())
			if buf.Len() != 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(fwdChat.Lastname())
			ret.OriginalChatName = buf.String()
		}

		if len(fwdChat.Username()) != 0 {
			buf.Reset()
			buf.WriteString("https://t.me/")
			buf.WriteString(fwdChat.Username())
			ret.OriginalChatLink = buf.String()

			if fwdHdr, ok := msg.GetFwdFrom(); ok {
				fwdMsgID, ok := fwdHdr.GetChannelPost()
				if ok {
					buf.WriteString("/")
					buf.WriteString(formatMessageID(rt.MessageID(fwdMsgID)))
					ret.OriginalMessageLink = buf.String()
				}
			}
		}
	}

	ret.ChatName = src.Chat.Title()
	if len(ret.ChatName) == 0 {
		buf.Reset()
		buf.WriteString(src.Chat.Firstname())
		if buf.Len() != 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(src.Chat.Lastname())

		ret.ChatName = buf.String()
	}

	if !ret.IsPrivateMessage && len(src.Chat.Username()) != 0 {
		buf.Reset()
		buf.WriteString("https://t.me/")
		buf.WriteString(src.Chat.Username())
		ret.ChatLink = buf.String()

		buf.WriteString("/")
		buf.WriteString(formatMessageID(ret.ID))
		ret.MessageLink = buf.String()
	}

	ret.Author = src.From.Title()
	if len(ret.Author) == 0 {
		buf.Reset()
		buf.WriteString(src.From.Firstname())
		if buf.Len() != 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(src.From.Lastname())
		ret.Author = buf.String()
	}

	if len(src.From.Username()) != 0 {
		buf.Reset()
		buf.WriteString("https://t.me/")
		buf.WriteString(src.From.Username())
		ret.AuthorLink = buf.String()
	}

	if !src.FwdFrom.IsNil() {
		fwdFrom := src.FwdFrom.GetPtr()

		ret.OriginalAuthor = fwdFrom.Title()
		if len(ret.OriginalAuthor) == 0 {
			buf.Reset()
			buf.WriteString(fwdFrom.Firstname())
			if buf.Len() != 0 {
				buf.WriteString(" ")
			}
			buf.WriteString(fwdFrom.Lastname())
			ret.OriginalAuthor = buf.String()
		}

		if len(fwdFrom.Username()) != 0 {
			buf.Reset()
			buf.WriteString("https://t.me/")
			buf.WriteString(fwdFrom.Username())
			ret.OriginalAuthorLink = buf.String()
		}
	}

	return
}

func parseTextEntities(text string, entities []tg.MessageEntityClass) (spans []rt.Span) {
	if entities == nil {
		spans = append(spans, rt.Span{
			Flags: rt.SpanFlag_PlainText,
			Text:  text,
		})
		return
	}

	var (
		this *rt.Span

		start, end int
		prevEnd    int
	)

	for _, e := range entities {
		start = e.GetOffset()
		end = start + e.GetLength()

		switch {
		case start > prevEnd:
			// append previously untouched plain text
			spans = append(spans, rt.Span{
				Flags: rt.SpanFlag_PlainText,
				Text:  text[prevEnd:start],
			})

			fallthrough
		default: // start == lastEnd
			spans = append(spans, rt.Span{
				Text: text[start:end],
			})

			fallthrough
		case start < prevEnd:
			this = &spans[len(spans)-1]
		}

		prevEnd = end

		// handle entities without params
		switch t := e.(type) {
		case *tg.MessageEntityUnknown,
			*tg.MessageEntityCashtag,
			*tg.MessageEntitySpoiler:
			this.Flags |= rt.SpanFlag_PlainText

		case *tg.MessageEntityBotCommand,
			*tg.MessageEntityCode,
			*tg.MessageEntityBankCard:
			this.Flags |= rt.SpanFlag_Code

		case *tg.MessageEntityHashtag:
			this.Flags |= rt.SpanFlag_HashTag

		case *tg.MessageEntityBold:
			this.Flags |= rt.SpanFlag_Bold

		case *tg.MessageEntityItalic:
			this.Flags |= rt.SpanFlag_Italic

		case *tg.MessageEntityStrike:
			this.Flags |= rt.SpanFlag_Strikethrough

		case *tg.MessageEntityUnderline:
			this.Flags |= rt.SpanFlag_Underline

		case *tg.MessageEntityPre:
			this.Flags |= rt.SpanFlag_Pre

		case *tg.MessageEntityEmail:
			this.Flags |= rt.SpanFlag_Email

		case *tg.MessageEntityPhone:
			this.Flags |= rt.SpanFlag_PhoneNumber

		case *tg.MessageEntityBlockquote:
			this.Flags |= rt.SpanFlag_Blockquote

		case *tg.MessageEntityURL:
			this.Flags |= rt.SpanFlag_URL
			this.URL = this.Text

		case *tg.MessageEntityTextURL:
			this.Flags |= rt.SpanFlag_URL
			this.URL = t.GetURL()

		case *tg.MessageEntityMention:
			this.Flags |= rt.SpanFlag_Mention
			this.URL = "https://t.me/" + strings.TrimPrefix(this.Text, "@")

		case *tg.MessageEntityMentionName:
			this.Flags |= rt.SpanFlag_Mention
			this.URL = "https://t.me/" + strconv.FormatInt(t.GetUserID(), 10)
		default:
			// TODO: log error?
			// client.logger.E("message entity unhandled", log.String("type", string(e.Type)))
		}
	}

	// add untouched remainder in message
	if prevEnd < len(text)-1 {
		spans = append(spans, rt.Span{
			Flags: rt.SpanFlag_PlainText,
			Text:  text[prevEnd:],
		})
	}

	return
}
