package telegram

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/stringhelper"

	api "arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/message"

	_ "embed"
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

func formatMessageID(msgID int) string {
	return strconv.FormatInt(int64(msgID), 10)
}

func renderEntities(entities []message.Entity) string {
	var (
		buf strings.Builder
	)

	for _, ent := range entities {
		switch {
		case ent.IsEmail():
			buf.WriteString(`<a href="mailto:`)
			buf.WriteString(ent.Text)
			buf.WriteString(`">`)
		case ent.IsPhoneNumber():
			buf.WriteString(`<a href="tel:`)
			buf.WriteString(ent.Text)
			buf.WriteString(`">`)
		case ent.IsURL():
			buf.WriteString(`<a href="`)
			buf.WriteString(ent.Params[message.EntityParamURL].(string))
			buf.WriteString(`">`)
		default:
			if ent.IsBold() {
				buf.WriteString("<strong>")
			}

			if ent.IsItalic() {
				buf.WriteString("<em>")
			}

			if ent.IsStrikethrough() {
				buf.WriteString("<del>")
			}

			if ent.IsUnderline() {
				buf.WriteString("<u>")
			}

			if ent.IsPre() {
				buf.WriteString("<pre>")
			}

			if ent.IsCode() {
				buf.WriteString("<code>")
			}
		}

		buf.WriteString(ent.Text)

		switch {
		case ent.Kind&(message.KindURL|message.KindPhoneNumber|message.KindEmail) != 0:
			buf.WriteString("</a>")
		default:
			if ent.IsCode() {
				buf.WriteString("</code>")
			}

			if ent.IsPre() {
				buf.WriteString("</pre>")
			}

			if ent.IsUnderline() {
				buf.WriteString("</u>")
			}

			if ent.IsStrikethrough() {
				buf.WriteString("</del>")
			}

			if ent.IsItalic() {
				buf.WriteString("</em>")
			}

			if ent.IsBold() {
				buf.WriteString("</strong>")
			}
		}
	}

	return buf.String()
}

func (c *telegramBot) scheduleMessageDelete(chatID uint64, after time.Duration, msgIDs ...uint64) {
	for _, msgID := range msgIDs {
		if msgID == 0 {
			// ignore invalid message id
			continue
		}

		_ = c.msgDelQ.OfferWithDelay(msgDeleteKey{
			chatID: chatID,
			msgID:  msgID,
		}, struct{}{}, after)
	}
}

func (c *telegramBot) sendTextMessage(
	chatID uint64,
	disableNotification,
	disableWebPreview bool,
	replyTo int,
	text string,
	replyMarkup ...any,
) (msgID int, err error) {
	defer func() {
		if err != nil {
			c.Logger().I("failed to send message", log.Error(err))
		}
	}()

	var (
		replyToMsgIDPtr *int
		replyMarkupPtr  *any
		parseMode       = "HTML"
	)

	if replyTo > 0 {
		replyToMsgIDPtr = &replyTo
	}

	if len(replyMarkup) > 0 {
		replyMarkupPtr = &replyMarkup[0]
	}

	resp, err := c.client.PostSendMessage(
		c.Context(),
		api.PostSendMessageJSONRequestBody{
			AllowSendingWithoutReply: constant.True(),
			ChatId:                   chatID,
			DisableNotification:      &disableNotification,
			DisableWebPagePreview:    &disableWebPreview,
			ReplyToMessageId:         replyToMsgIDPtr,
			ParseMode:                &parseMode,
			Text:                     text,
			ReplyMarkup:              replyMarkupPtr,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("send message: %w", err)
	}

	result, err := api.ParsePostSendMessageResponse(resp)
	_ = resp.Body.Close()
	if err != nil {
		return 0, fmt.Errorf("parse response of message send: %w", err)
	}

	if result.JSON200 == nil || !result.JSON200.Ok {
		return 0, fmt.Errorf("send message: %s", result.JSONDefault.Description)
	}

	return result.JSON200.Result.MessageId, nil
}

func parseTelegramEntities(
	content string,
	entities *[]api.MessageEntity,
) (ret message.Entities) {
	ret.Init()

	if entities == nil {
		ret.Append(message.Entity{
			Kind: message.KindPlainText,
			Text: content,
		})
		return
	}

	text := utf16.Encode([]rune(content))

	lastIndex := 0
	for _, e := range *entities {
		switch {
		case e.Offset > lastIndex:
			// append previously untouched plain text
			ret.Append(message.Entity{
				Kind:   message.KindPlainText,
				Text:   string(utf16.Decode(text[lastIndex:e.Offset])),
				Params: nil,
			})
		case e.Offset < lastIndex:
			// TODO: handle multiple styles on overlapped range of text
		}

		data := string(utf16.Decode(text[e.Offset : e.Offset+e.Length]))
		lastIndex = e.Offset + e.Length

		// handle entities without params
		var kind message.EntityKind
		switch e.Type {
		case api.MessageEntityTypeBotCommand,
			api.MessageEntityTypeCode:
			kind = message.KindCode
		case api.MessageEntityTypeHashtag,
			api.MessageEntityTypeCashtag,
			api.MessageEntityTypeTextLink,
			api.MessageEntityTypeBold:
			kind = message.KindBold
		case api.MessageEntityTypeItalic:
			kind = message.KindItalic
		case api.MessageEntityTypeStrikethrough:
			kind = message.KindStrikethrough
		case api.MessageEntityTypeUnderline:
			kind = message.KindUnderline
		case api.MessageEntityTypePre:
			kind = message.KindPre

		case api.MessageEntityTypeEmail:
			kind = message.KindEmail
		case api.MessageEntityTypePhoneNumber:
			kind = message.KindPhoneNumber

		case api.MessageEntityTypeUrl:
			ret.Append(message.Entity{
				Kind: message.KindURL,
				Text: data,
				Params: map[message.EntityParamKey]interface{}{
					message.EntityParamURL:                     data,
					message.EntityParamWebArchiveURL:           "",
					message.EntityParamWebArchiveScreenshotURL: "",
				},
			})
			continue
		case api.MessageEntityTypeMention, api.MessageEntityTypeTextMention:
			url := "https://t.me/" + strings.TrimPrefix(data, "@")
			ret.Append(message.Entity{
				Kind: message.KindURL,
				Text: data,
				Params: map[message.EntityParamKey]interface{}{
					message.EntityParamURL:                     url,
					message.EntityParamWebArchiveURL:           "",
					message.EntityParamWebArchiveScreenshotURL: "",
				},
			})
			continue
		default:
			// TODO: log error
			// client.logger.E("message entity unhandled", log.String("type", string(e.Type)))
			continue
		}

		ret.Append(message.Entity{
			Kind:   kind,
			Text:   data,
			Params: nil,
		})
	}

	// add untouched remainder in message
	if lastIndex < len(text)-1 {
		ret.Append(message.Entity{
			Kind:   message.KindPlainText,
			Text:   string(utf16.Decode(text[lastIndex:])),
			Params: nil,
		})
	}

	return
}
