package telegram

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"

	"arhat.dev/pkg/log"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
	"arhat.dev/meeting-minutes-bot/pkg/message"

	_ "embed"
)

var (
	// message template to render entities
	//go:embed message.tpl
	messageTemplate string
)

func init() {
	// check template error
	_, err := template.New("").Parse(messageTemplate)
	if err != nil {
		panic(err)
	}
}

func encodeUint64Hex(n uint64) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n)
	return hex.EncodeToString(buf)
}

func decodeUint64Hex(s string) (uint64, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(data), nil
}

func formatMessageID(msgID int) string {
	return strconv.FormatInt(int64(msgID), 10)
}

func (c *telegramBot) renderEntities(entities []message.Entity) string {
	buf := &bytes.Buffer{}
	err := c.msgTpl.ExecuteTemplate(buf, "message", entities)
	if err != nil {
		c.logger.E("failed to execute message template", log.Error(err))
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
			chatID:    chatID,
			messageID: msgID,
		}, struct{}{}, after)
	}
}

func (c *telegramBot) sendTextMessage(
	chatID uint64,
	disableNotification,
	disableWebPreview bool,
	replyTo int,
	text string,
	replyMarkup ...interface{},
) (msgID int, err error) {
	defer func() {
		if err != nil {
			c.logger.I("failed to send message", log.Error(err))
		}
	}()

	var replyToMsgIDPtr *int
	if replyTo > 0 {
		replyToMsgIDPtr = &replyTo
	}

	var replyMarkupPtr *interface{}
	if len(replyMarkup) > 0 {
		replyMarkupPtr = &replyMarkup[0]
	}

	var htmlStyle = "HTML"
	resp, err := c.client.PostSendMessage(
		c.ctx,
		telegram.PostSendMessageJSONRequestBody{
			AllowSendingWithoutReply: constant.True(),
			ChatId:                   chatID,
			DisableNotification:      &disableNotification,
			DisableWebPagePreview:    &disableWebPreview,
			ReplyToMessageId:         replyToMsgIDPtr,
			ParseMode:                &htmlStyle,
			Text:                     text,
			ReplyMarkup:              replyMarkupPtr,
		},
	)
	if err != nil {
		return 0, fmt.Errorf("failed to send message: %w", err)
	}

	result, err := telegram.ParsePostSendMessageResponse(resp)
	_ = resp.Body.Close()
	if err != nil {
		return 0, fmt.Errorf("failed to parse response of message send: %w", err)
	}

	if result.JSON200 == nil || !result.JSON200.Ok {
		return 0, fmt.Errorf("telegram: failed to send message: %s", result.JSONDefault.Description)
	}

	return result.JSON200.Result.MessageId, nil
}

func parseTelegramEntities(
	content string,
	entities *[]telegram.MessageEntity,
) *message.Entities {
	if entities == nil {
		return message.NewMessageEntities([]message.Entity{{
			Kind: message.KindText,
			Text: content,
		}}, nil)
	}

	text := utf16.Encode([]rune(content))
	result := message.NewMessageEntities(nil, nil)

	textIndex := 0
	for _, e := range *entities {
		if e.Offset > textIndex {
			// append previously unhandled plain text
			result.Append(message.Entity{
				Kind:   message.KindText,
				Text:   string(utf16.Decode(text[textIndex:e.Offset])),
				Params: nil,
			})
		}

		data := string(utf16.Decode(text[e.Offset : e.Offset+e.Length]))
		textIndex = e.Offset + e.Length

		// handle entities without params
		kind, ok := map[telegram.MessageEntityType]message.EntityKind{
			telegram.MessageEntityTypeBotCommand: message.KindCode,
			telegram.MessageEntityTypeHashtag:    message.KindBold,
			telegram.MessageEntityTypeCashtag:    message.KindBold,
			telegram.MessageEntityTypeTextLink:   message.KindBold,

			telegram.MessageEntityTypeBold:          message.KindBold,
			telegram.MessageEntityTypeItalic:        message.KindItalic,
			telegram.MessageEntityTypeStrikethrough: message.KindStrikethrough,
			telegram.MessageEntityTypeUnderline:     message.KindUnderline,
			telegram.MessageEntityTypeCode:          message.KindCode,
			telegram.MessageEntityTypePre:           message.KindPre,

			telegram.MessageEntityTypeEmail:       message.KindEmail,
			telegram.MessageEntityTypePhoneNumber: message.KindPhoneNumber,
		}[e.Type]

		if ok {
			result.Append(message.Entity{
				Kind:   kind,
				Text:   data,
				Params: nil,
			})

			continue
		}

		switch e.Type {
		case telegram.MessageEntityTypeUrl:
			result.Append(message.Entity{
				Kind: message.KindURL,
				Text: data,
				Params: map[message.EntityParamKey]interface{}{
					message.EntityParamURL:                     data,
					message.EntityParamWebArchiveURL:           "",
					message.EntityParamWebArchiveScreenshotURL: "",
				},
			})
		case telegram.MessageEntityTypeMention, telegram.MessageEntityTypeTextMention:
			url := "https://t.me/" + strings.TrimPrefix(data, "@")
			result.Append(message.Entity{
				Kind: message.KindURL,
				Text: data,
				Params: map[message.EntityParamKey]interface{}{
					message.EntityParamURL:                     url,
					message.EntityParamWebArchiveURL:           "",
					message.EntityParamWebArchiveScreenshotURL: "",
				},
			})
		default:
			// TODO: log error
			// client.logger.E("message entity unhandled", log.String("type", string(e.Type)))
		}
	}

	if textIndex < len(text)-1 {
		result.Append(message.Entity{
			Kind:   message.KindText,
			Text:   string(utf16.Decode(text[textIndex:])),
			Params: nil,
		})
	}

	return result
}
