package telegram

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"arhat.dev/pkg/log"
	"arhat.dev/pkg/stringhelper"
	tm "github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"

	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

func encodeUint64Hex[T rt.ChatID | rt.UserID](n T) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(n))
	return hex.EncodeToString(buf[:])
}

func decodeUint64Hex[T rt.ChatID | rt.UserID](s string) (_ T, err error) {
	var buf [8]byte
	_, err = hex.Decode(buf[:], stringhelper.ToBytes[byte, byte](s))
	if err != nil {
		return 0, err
	}

	return T(binary.BigEndian.Uint64(buf[:])), nil
}

func formatMessageID(msgID rt.MessageID) string {
	return strconv.FormatUint(uint64(msgID), 10)
}

func translateSpans(spans []rt.Span) (ret []styling.StyledTextOption) {
	sz := len(spans)
	ret = make([]styling.StyledTextOption, 0, sz)

	for i := 0; i < sz; i++ {
		ent := &spans[i]

		switch {
		case ent.IsEmail():
			ret = append(ret, styling.Email(ent.Text))
		case ent.IsPhoneNumber():
			ret = append(ret, styling.Phone(ent.Text))
		case ent.IsURL(), ent.IsMention():
			if len(ent.URL) == 0 {
				ret = append(ret, styling.URL(ent.Text))
			} else {
				ret = append(ret, styling.TextURL(ent.Text, ent.URL))
			}
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
					styles[n] = entity.Pre(ent.Hint)
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
		case ent.IsMedia():
			// TODO: implement data uploading
		}

	}

	return
}

type msgDeleteKey struct {
	chatID rt.ChatID
	msgID  rt.MessageID
}

func (c *tgBot) scheduleMessageDelete(chat *chatSpec, after time.Duration, msgIDs ...rt.MessageID) {
	for _, msgID := range msgIDs {
		if msgID == 0 {
			// ignore invalid message id
			continue
		}

		_ = c.msgDelQ.OfferWithDelay(msgDeleteKey{
			chatID: chat.ID(),
			msgID:  msgID,
		}, chat.InputPeer(), after)
	}
}

func (c *tgBot) sendTextMessage(
	builder *tm.Builder,
	entities ...styling.StyledTextOption,
) (msgID rt.MessageID, err error) {
	updCls, err := builder.StyledText(c.Context(), entities...)
	if err != nil {
		c.Logger().E("failed to send message", log.Error(err))
		return
	}

	switch resp := updCls.(type) {
	case *tg.UpdatesTooLong:
		err = fmt.Errorf("too many updates")
	case *tg.UpdateShortMessage:
		msgID = rt.MessageID(resp.GetID())
	case *tg.UpdateShortChatMessage:
		msgID = rt.MessageID(resp.GetID())
	case *tg.UpdateShort:
		msgID, _ = extractMsgID(resp.GetUpdate())
	case *tg.UpdatesCombined:
		upds := resp.GetUpdates()
		for i := range upds {
			var ok bool
			msgID, ok = extractMsgID(upds[i])
			if ok {
				return
			}
		}
	case *tg.Updates:
		upds := resp.GetUpdates()
		for i := range upds {
			var ok bool
			msgID, ok = extractMsgID(upds[i])
			if ok {
				return
			}
		}
	case *tg.UpdateShortSentMessage:
		msgID = rt.MessageID(resp.GetID())
	default:
		err = fmt.Errorf("unexpected response type %T", resp)
		return
	}

	return
}

func extractMsgID(upd tg.UpdateClass) (msgID rt.MessageID, hasMsgID bool) {
	switch u := upd.(type) {
	case *tg.UpdateNewMessage:
		return rt.MessageID(u.GetMessage().GetID()), true
	case *tg.UpdateMessageID:
		return rt.MessageID(u.GetID()), true
	default:
		return
	}
}

func (c *tgBot) getChannelCreator(ch *tg.InputChannel) (ret *tg.User, err error) {
	const (
		BATCH_COUNT = 10
	)
	var (
		chAdmins tg.ChannelsChannelParticipantsClass
		offset   int
	)

QUERY:
	chAdmins, err = c.client.API().ChannelsGetParticipants(c.Context(), &tg.ChannelsGetParticipantsRequest{
		Channel: ch,
		Filter:  &tg.ChannelParticipantsAdmins{},
		Offset:  offset,
		Limit:   BATCH_COUNT,
	})
	if err != nil {
		return
	}

	offset += BATCH_COUNT

	switch adms := chAdmins.(type) {
	case *tg.ChannelsChannelParticipants:
		var (
			i          int
			creatorUID int64
		)
		pts := adms.GetParticipants()
		for i = range pts {
			creator, ok := pts[i].(*tg.ChannelParticipantCreator)
			if !ok {
				continue
			}

			creatorUID = creator.GetUserID()
		}
		if creatorUID == 0 {
			goto QUERY
		}

		users := adms.GetUsers()
		if len(users) > i && users[i].GetID() == creatorUID {
			var ok bool
			ret, ok = users[i].(*tg.User)
			if ok {
				return
			}
		}

		for _, user := range users {
			switch u := user.(type) {
			case *tg.User:
				if u.GetID() == creatorUID {
					ret = u
					return
				}
			}
		}

		err = fmt.Errorf("unexpected no user detail for creator")
		return
	case *tg.ChannelsChannelParticipantsNotModified:
		return nil, fmt.Errorf("no creator found")
	default:
		return nil, fmt.Errorf("unknown type_id %d", adms.TypeID())
	}
}
