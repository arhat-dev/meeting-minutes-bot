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
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"

	"arhat.dev/mbot/pkg/rt"
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

type msgDeleteKey struct {
	chatID rt.ChatID
	msgID  rt.MessageID
}

func (c *tgBot) scheduleMessageDelete(chat *chatInfo, after time.Duration, msgIDs ...rt.MessageID) {
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

	var buf [1]rt.MessageID
	out := buf[0:0:1]
	err = handleMessageSent(updCls, &out)
	if len(out) == 1 {
		msgID = buf[0]
	}
	return
}

// getChannelCreator get creator of the channel, requires being channel admin
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
