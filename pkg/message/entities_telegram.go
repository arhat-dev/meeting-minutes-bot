package message

import (
	"strings"
	"unicode/utf16"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
)

func ParseTelegramEntities(
	content string,
	entities *[]telegram.MessageEntity,
) *Entities {
	if entities == nil {
		return newMessageEntities([]Entity{{
			Kind: KindText,
			Text: content,
		}}, nil)
	}

	text := utf16.Encode([]rune(content))
	result := newMessageEntities(nil, nil)

	textIndex := 0
	for _, e := range *entities {
		if e.Offset > textIndex {
			// append previously unhandled plain text
			result.entities = append(result.entities, Entity{
				Kind:   KindText,
				Text:   string(utf16.Decode(text[textIndex:e.Offset])),
				Params: nil,
			})
		}

		data := string(utf16.Decode(text[e.Offset : e.Offset+e.Length]))
		textIndex = e.Offset + e.Length

		// handle entities without params
		kind, ok := map[telegram.MessageEntityType]EntityKind{
			telegram.MessageEntityTypeBotCommand: KindCode,
			telegram.MessageEntityTypeHashtag:    KindBold,
			telegram.MessageEntityTypeCashtag:    KindBold,
			telegram.MessageEntityTypeTextLink:   KindBold,

			telegram.MessageEntityTypeBold:          KindBold,
			telegram.MessageEntityTypeItalic:        KindItalic,
			telegram.MessageEntityTypeStrikethrough: KindStrikethrough,
			telegram.MessageEntityTypeUnderline:     KindUnderline,
			telegram.MessageEntityTypeCode:          KindCode,
			telegram.MessageEntityTypePre:           KindPre,

			telegram.MessageEntityTypeEmail:       KindEmail,
			telegram.MessageEntityTypePhoneNumber: KindPhoneNumber,
		}[e.Type]

		if ok {
			result.entities = append(result.entities, Entity{
				Kind:   kind,
				Text:   data,
				Params: nil,
			})

			continue
		}

		switch e.Type {
		case telegram.MessageEntityTypeUrl:
			result.entities = append(result.entities, Entity{
				Kind: KindURL,
				Text: data,
				Params: map[EntityParamKey]interface{}{
					EntityParamURL:                     data,
					EntityParamWebArchiveURL:           "",
					EntityParamWebArchiveScreenshotURL: "",
				},
			})

			result.urlsToArchive[data] = append(result.urlsToArchive[data], len(result.entities)-1)
		case telegram.MessageEntityTypeMention, telegram.MessageEntityTypeTextMention:
			url := "https://t.me/" + strings.TrimPrefix(data, "@")
			result.entities = append(result.entities, Entity{
				Kind: KindURL,
				Text: data,
				Params: map[EntityParamKey]interface{}{
					EntityParamURL:                     url,
					EntityParamWebArchiveURL:           "",
					EntityParamWebArchiveScreenshotURL: "",
				},
			})

			// TODO: do we really need to archive user page? (while reasonable to me)
			result.urlsToArchive[data] = append(result.urlsToArchive[data], len(result.entities)-1)
		default:
			// TODO: log error
			// client.logger.E("message entity unhandled", log.String("type", string(e.Type)))
		}
	}

	if textIndex < len(text)-1 {
		result.entities = append(result.entities, Entity{
			Kind:   KindText,
			Text:   string(utf16.Decode(text[textIndex:])),
			Params: nil,
		})
	}

	return result
}
