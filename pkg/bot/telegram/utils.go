package telegram

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/stringhelper"
	tm "github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"

	"arhat.dev/meeting-minutes-bot/pkg/message"
)

func encodeUint64Hex(n uint64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], n)
	return hex.EncodeToString(buf[:])
}

func decodeUint64Hex(s string) (_ uint64, err error) {
	var buf [8]byte
	_, err = hex.Decode(buf[:], stringhelper.ToBytes[byte, byte](s))
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint64(buf[:]), nil
}

func formatMessageID[T int | int64 | uint64](msgID T) string {
	return strconv.FormatUint(uint64(msgID), 10)
}

func translateEntities(entities []message.Span) (ret []styling.StyledTextOption) {
	ret = make([]styling.StyledTextOption, 0, len(entities))

	for _, ent := range entities {
		switch {
		case ent.IsEmail():
			ret = append(ret, styling.Email(ent.Text))
		case ent.IsPhoneNumber():
			ret = append(ret, styling.Phone(ent.Text))
		case ent.IsURL():
			ret = append(ret, styling.URL(ent.Text))
		case ent.IsStyledText():
			ret = append(ret, styling.Custom(func(b *entity.Builder) error {
				var (
					styles [8]entity.Formatter
					n      int
				)

				if ent.IsBlockquote() {
					styles[n] = entity.Blockquote()
					n++
				}

				if ent.IsBold() {
					styles[n] = entity.Bold()
					n++
				}

				if ent.IsItalic() {
					styles[n] = entity.Italic()
					n++
				}

				if ent.IsStrikethrough() {
					styles[n] = entity.Strike()
					n++
				}

				if ent.IsUnderline() {
					styles[n] = entity.Underline()
					n++
				}

				if ent.IsPre() {
					styles[n] = entity.Pre("")
					n++
				}

				if ent.IsCode() {
					styles[n] = entity.Code()
					n++
				}

				b.Format(ent.Text, styles[:n]...)
				return nil
			}))
		case ent.IsPlainText():
			ret = append(ret, styling.Plain(ent.Text))
		case ent.IsMultiMedia():
			// TODO: implement data uploading
		}

	}

	return
}

func (c *tgBot) scheduleMessageDelete(chat *chatSpec, after time.Duration, msgIDs ...uint64) {
	for _, msgID := range msgIDs {
		if msgID == 0 {
			// ignore invalid message id
			continue
		}

		_ = c.msgDelQ.OfferWithDelay(msgDeleteKey{
			chatID: uint64(chat.ID()),
			msgID:  msgID,
		}, chat.InputPeer(), after)
	}
}

func (c *tgBot) sendTextMessage(
	builder *tm.Builder, entities ...styling.StyledTextOption,
) (msgID int, err error) {
	updCls, err := builder.StyledText(c.Context(), entities...)
	if err != nil {
		c.Logger().E("failed to send message", log.Error(err))
		return
	}

	switch resp := updCls.(type) {
	case *tg.UpdateShortSentMessage:
		msgID = resp.GetID()
	default:
		err = fmt.Errorf("unexpected response type %T", resp)
		return
	}

	return
}

func parseTextEntities(text string, entities []tg.MessageEntityClass) (ret message.Entities) {
	if entities == nil {
		ret = append(ret, message.Span{
			SpanFlags: message.SpanFlags_PlainText,
			Text:      text,
		})
		return
	}

	var (
		this *message.Span

		start, end         int
		prevStart, prevEnd int
	)

	for _, e := range entities {
		start = e.GetOffset()
		end = start + e.GetLength()

		switch {
		case start > prevEnd:
			// append previously untouched plain text
			ret = append(ret, message.Span{
				SpanFlags: message.SpanFlags_PlainText,
				Text:      text[prevEnd:start],
			})

			fallthrough
		default: // start == lastEnd
			ret = append(ret, message.Span{
				Text: text[start:end],
			})
			this = &ret[len(ret)-1]

		case start < prevEnd:
			this = &ret[len(ret)-1]
			ret[len(ret)-2].Text = text[prevStart:start]
		}

		prevStart, prevEnd = start, end

		// handle entities without params
		switch t := e.(type) {
		case *tg.MessageEntityUnknown,
			*tg.MessageEntityCashtag,
			*tg.MessageEntitySpoiler:
			this.SpanFlags |= message.SpanFlags_PlainText

		case *tg.MessageEntityBotCommand,
			*tg.MessageEntityCode,
			*tg.MessageEntityBankCard:
			this.SpanFlags |= message.SpanFlags_Code

		case *tg.MessageEntityHashtag:
			this.SpanFlags |= message.SpanFlags_HashTag

		case *tg.MessageEntityBold:
			this.SpanFlags |= message.SpanFlags_Bold

		case *tg.MessageEntityItalic:
			this.SpanFlags |= message.SpanFlags_Italic

		case *tg.MessageEntityStrike:
			this.SpanFlags |= message.SpanFlags_Strikethrough

		case *tg.MessageEntityUnderline:
			this.SpanFlags |= message.SpanFlags_Underline

		case *tg.MessageEntityPre:
			this.SpanFlags |= message.SpanFlags_Pre

		case *tg.MessageEntityEmail:
			this.SpanFlags |= message.SpanFlags_Email

		case *tg.MessageEntityPhone:
			this.SpanFlags |= message.SpanFlags_PhoneNumber

		case *tg.MessageEntityBlockquote:
			this.SpanFlags |= message.SpanFlags_Blockquote

		case *tg.MessageEntityURL:
			this.SpanFlags |= message.SpanFlags_URL
			this.URL.Set(this.Text)

		case *tg.MessageEntityTextURL:
			this.SpanFlags |= message.SpanFlags_URL
			this.URL.Set(t.GetURL())

		case *tg.MessageEntityMention:
			this.SpanFlags |= message.SpanFlags_Mention
			this.URL.Set("https://t.me/" + strings.TrimPrefix(this.Text, "@"))

		case *tg.MessageEntityMentionName:
			this.SpanFlags |= message.SpanFlags_Mention
			this.URL.Set("https://t.me/" + strconv.FormatInt(t.GetUserID(), 10))
		default:
			// TODO: log error?
			// client.logger.E("message entity unhandled", log.String("type", string(e.Type)))
		}
	}

	// add untouched remainder in message
	if prevEnd < len(text)-1 {
		ret = append(ret, message.Span{
			SpanFlags: message.SpanFlags_PlainText,
			Text:      text[prevEnd:],
		})
	}

	return
}
