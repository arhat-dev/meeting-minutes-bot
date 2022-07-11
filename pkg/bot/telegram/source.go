package telegram

import (
	"fmt"

	"arhat.dev/meeting-minutes-bot/pkg/rt"
	"github.com/gotd/td/tg"
)

type messageSource struct {
	Chat chatSpec
	From rt.Optional[authorSpec]

	FwdChat rt.Optional[chatSpec]
	FwdFrom rt.Optional[authorSpec]
}

type chatFlag uint32

const (
	// one on one private message
	chatFlag_PM chatFlag = 1 << iota

	// group chat
	chatFlag_Group

	// channel messages
	chatFlag_Channel
)

type commonSpec struct {
	id int64

	firstname, lastname string
	titile              string

	// username of the chat (for user and channel only)
	// https://t.me/{username}
	username string
}

func (cs *commonSpec) ID() int64         { return cs.id }
func (cs *commonSpec) Username() string  { return cs.username }
func (cs *commonSpec) Title() string     { return cs.titile }
func (cs *commonSpec) Firstname() string { return cs.firstname }
func (cs *commonSpec) Lastname() string  { return cs.lastname }

type chatSpec struct {
	chatFlag

	commonSpec

	accessHash int64
}

func (cs *chatSpec) InputPeer() tg.InputPeerClass {
	switch {
	case cs.chatFlag&chatFlag_PM != 0:
		return &tg.InputPeerUser{
			UserID:     cs.id,
			AccessHash: cs.accessHash,
		}
	case cs.chatFlag&chatFlag_Group != 0:
		return &tg.InputPeerChannel{
			ChannelID:  cs.id,
			AccessHash: cs.accessHash,
		}
	case cs.chatFlag&chatFlag_Channel != 0:
		return &tg.InputPeerChat{
			ChatID: cs.id,
		}
	default:
		panic("unreachable")
	}
}

func (cs *chatSpec) IsPrivateChat() bool { return cs.chatFlag&chatFlag_PM != 0 }

func resolveChatSpec(chat any) (ret chatSpec) {
	switch c := chat.(type) {
	case *tg.User:
		ret.id = c.GetID()

		ret.chatFlag |= chatFlag_PM
		// TODO: set title as {FirstName} {LastName}
		ret.username, _ = c.GetUsername()
		ret.firstname, _ = c.GetFirstName()
		ret.lastname, _ = c.GetLastName()

		ret.accessHash, _ = c.GetAccessHash()
	case *tg.Channel:
		ret.id = c.GetID()

		ret.chatFlag |= chatFlag_Channel
		ret.titile = c.GetTitle()
		ret.username, _ = c.GetUsername()

		ret.accessHash, _ = c.GetAccessHash()
	case *tg.Chat:
		ret.id = c.GetID()

		ret.chatFlag |= chatFlag_Group
		ret.titile = c.GetTitle()
	default:
		panic("unreachable")
	}

	return
}

type authorFlag uint32

const (
	authorFlag_User authorFlag = 1 << iota
	authorFlag_Channel
	authorFlag_Group
	authorFlag_Bot
)

type authorSpec struct {
	authorFlag

	commonSpec
}

func resolveAuthorSpec(from any) (ret authorSpec) {
	switch f := from.(type) {
	case *tg.User:
		ret.id = f.GetID()

		ret.authorFlag |= authorFlag_User
		if f.GetBot() {
			ret.authorFlag |= authorFlag_Bot
		}

		ret.username, _ = f.GetUsername()
		ret.firstname, _ = f.GetFirstName()
		ret.lastname, _ = f.GetLastName()
	case *tg.Channel:
		ret.id = f.GetID()

		ret.authorFlag |= authorFlag_Channel
		ret.username, _ = f.GetUsername()
		ret.titile = f.GetTitle()
	case *tg.Chat:
		ret.id = f.GetID()

		ret.authorFlag |= authorFlag_Group
		ret.titile = f.GetTitle()
	default:
		panic("unreachable")
	}

	return
}

func extractPeer(e tg.Entities, p tg.PeerClass) (any, error) {
	var id int64

	switch t := p.(type) {
	case *tg.PeerUser: // private message
		id = t.GetUserID()

		if len(e.Users) != 0 {
			u, ok := e.Users[id]
			if ok {
				return u, nil
			}
		}

		return nil, fmt.Errorf("unknown user %d", id)
	case *tg.PeerChannel: // channel/subgroup
		id = t.GetChannelID()

		if len(e.Channels) != 0 {
			u, ok := e.Channels[id]
			if ok {
				return u, nil
			}
		}

		return nil, fmt.Errorf("unknown channel %d", id)
	case *tg.PeerChat: // group chat
		id = t.GetChatID()

		if len(e.Chats) != 0 {
			u, ok := e.Chats[id]
			if !ok {
				return u, nil
			}
		}

		return nil, fmt.Errorf("unknown chat %d", id)
	default:
		return nil, fmt.Errorf("unknown peer type: %T", t)
	}
}
