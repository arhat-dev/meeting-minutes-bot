package telegram

import (
	"fmt"

	"arhat.dev/mbot/pkg/rt"
	"github.com/gotd/td/tg"
)

type messageSource struct {
	Chat chatSpec
	From authorSpec

	FwdChat rt.Optional[chatSpec]
	FwdFrom rt.Optional[authorSpec]
}

type chatFlag uint32

const (
	// one on one private message
	chatFlag_PM chatFlag = 1 << iota

	// megagroup, supergroup
	chatFlag_Group

	// channel
	chatFlag_Channel

	chatFlag_LegacyGroup
)

func (f chatFlag) IsPrivateChat() bool     { return f&chatFlag_PM != 0 }
func (f chatFlag) IsChannelChat() bool     { return f&chatFlag_Channel != 0 }
func (f chatFlag) IsGroupChat() bool       { return f&chatFlag_Group != 0 }
func (f chatFlag) IsLegacyGroupChat() bool { return f&chatFlag_LegacyGroup != 0 }

type commonSpec[ID rt.UserID | rt.ChatID] struct {
	id ID

	firstname, lastname string
	titile              string

	// username of the chat (for user and channel only)
	// https://t.me/{username}
	username string
}

func (cs *commonSpec[ID]) ID() ID            { return cs.id }
func (cs *commonSpec[ID]) Username() string  { return cs.username }
func (cs *commonSpec[ID]) Title() string     { return cs.titile }
func (cs *commonSpec[ID]) Firstname() string { return cs.firstname }
func (cs *commonSpec[ID]) Lastname() string  { return cs.lastname }

type chatSpec struct {
	chatFlag

	commonSpec[rt.ChatID]

	peer tg.InputPeerClass
}

func (cs *chatSpec) InputPeer() tg.InputPeerClass { return cs.peer }

func resolveChatSpec(chatPeer any) (ret chatSpec) {
	switch c := chatPeer.(type) {
	case *tg.User:
		ret.id = rt.ChatID(c.GetID())

		ret.chatFlag |= chatFlag_PM
		// TODO: set title as {FirstName} {LastName}
		ret.username, _ = c.GetUsername()
		ret.firstname, _ = c.GetFirstName()
		ret.lastname, _ = c.GetLastName()

		ret.peer = c.AsInputPeer()
	case *tg.Channel:
		ret.id = rt.ChatID(c.GetID())

		if c.GetBroadcast() {
			ret.chatFlag |= chatFlag_Channel
		} else {
			ret.chatFlag |= chatFlag_Group
		}

		ret.titile = c.GetTitle()
		ret.username, _ = c.GetUsername()

		ret.peer = c.AsInputPeer()
	case *tg.Chat:
		ret.id = rt.ChatID(c.GetID())

		ret.chatFlag |= chatFlag_LegacyGroup
		ret.titile = c.GetTitle()
		ret.peer = c.AsInputPeer()
	default:
		panic("unreachable")
	}

	return
}

type authorFlag uint32

const (
	authorFlag_User    authorFlag = 1 << iota
	authorFlag_Channel            // broadcast channel
	authorFlag_Group              // megagroup supergroup
	authorFlag_Bot
	authorFlag_Verified
)

func (f authorFlag) IsUser() bool { return f&authorFlag_User != 0 }

type authorSpec struct {
	authorFlag

	commonSpec[rt.UserID]

	user *tg.InputUser
}

func (as *authorSpec) InputUser() *tg.InputUser { return as.user }

func resolveAuthorSpec(peerFrom any) (ret authorSpec, err error) {
	switch p := peerFrom.(type) {
	case *tg.User:
		ret.id = rt.UserID(p.GetID())

		ret.authorFlag |= authorFlag_User
		if p.GetBot() {
			ret.authorFlag |= authorFlag_Bot
		}

		if p.GetVerified() {
			ret.authorFlag |= authorFlag_Verified
		}

		ret.username, _ = p.GetUsername()
		ret.firstname, _ = p.GetFirstName()
		ret.lastname, _ = p.GetLastName()

		ret.user = p.AsInput()
	case *tg.Channel:
		ret.id = rt.UserID(p.GetID())

		if p.GetBroadcast() {
			ret.authorFlag |= authorFlag_Channel
		} else {
			ret.authorFlag |= authorFlag_Group
		}

		if p.GetVerified() {
			ret.authorFlag |= authorFlag_Verified
		}

		ret.titile = p.GetTitle()
		ret.username, _ = p.GetUsername()
	case *tg.Chat: // TODO: is it possible?
		err = fmt.Errorf("unsupported chat as author")
		return
	default:
		err = fmt.Errorf("unknown type %T", p)
		return
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

type chatIDWrapper struct {
	chat tg.InputPeerClass
}

func (c chatIDWrapper) ID() rt.ChatID {
	switch this := c.chat.(type) {
	case *tg.InputPeerChat:
		return rt.ChatID(this.GetChatID())
	case *tg.InputPeerUser:
		return rt.ChatID(this.GetUserID())
	case *tg.InputPeerChannel:
		return rt.ChatID(this.GetChannelID())
	case *tg.InputPeerUserFromMessage:
		return rt.ChatID(this.GetUserID())
	case *tg.InputPeerChannelFromMessage:
		return rt.ChatID(this.GetChannelID())
	default:
		return 0
	}
}

func (c chatIDWrapper) Equals(o chatIDWrapper) bool {
	switch {
	case c.chat == nil && o.chat == nil:
		return true
	case c.chat == nil || o.chat == nil:
		return false

	// both are not nil
	case c.chat == o.chat:
		return true
	case c.chat.TypeID() != c.chat.TypeID():
		return false
	}

	switch this := c.chat.(type) {
	case *tg.InputPeerEmpty:
		return true
	case *tg.InputPeerSelf:
		return true
	case *tg.InputPeerChat:
		return this.GetChatID() == o.chat.(*tg.InputPeerChat).GetChatID()
	case *tg.InputPeerUser:
		return this.GetUserID() == o.chat.(*tg.InputPeerUser).GetUserID()
	case *tg.InputPeerChannel:
		return this.GetChannelID() == o.chat.(*tg.InputPeerChannel).GetChannelID()
	case *tg.InputPeerUserFromMessage:
		return this.GetUserID() == o.chat.(*tg.InputPeerUserFromMessage).GetUserID() &&
			this.GetMsgID() == o.chat.(*tg.InputPeerUserFromMessage).GetMsgID()
	case *tg.InputPeerChannelFromMessage:
		return this.GetChannelID() == o.chat.(*tg.InputPeerChannelFromMessage).GetChannelID() &&
			this.GetMsgID() == o.chat.(*tg.InputPeerUserFromMessage).GetMsgID()
	default:
		return false
	}
}
