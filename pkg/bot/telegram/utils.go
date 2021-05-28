package telegram

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/constant"
)

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

func (c telegramBot) scheduleMessageDelete(chatID uint64, after time.Duration, msgIDs ...uint64) {
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
) (int, error) {
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
