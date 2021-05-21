package server

import (
	"bytes"
	"fmt"
	"html"
	"strconv"
	"sync"
	"unicode/utf16"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
)

func newSession(topic, defaultChatUsername string, gen generator.Interface) *session {
	return &session{
		Topic:    topic,
		Messages: make([]*telegram.Message, 0, 16),

		defaultChatUsername: defaultChatUsername,
		generator:           gen,
		msgIdx:              make(map[int]int),
		mu:                  &sync.RWMutex{},
	}
}

type session struct {
	Topic    string
	Messages []*telegram.Message

	defaultChatUsername string
	generator           generator.Interface
	msgIdx              map[int]int
	mu                  *sync.RWMutex
}

func (s *session) appendMessage(msg *telegram.Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Messages = append(s.Messages, msg)
	s.msgIdx[msg.MessageId] = len(s.Messages) - 1
}

func (s *session) deleteMessage(msgID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, ok := s.msgIdx[msgID]
	if !ok {
		// no such id, ignore
		return false
	}

	delete(s.msgIdx, msgID)
	s.Messages = append(s.Messages[:idx], s.Messages[idx+1:]...)
	return true
}

func (s *session) deleteFirstNMessage(n int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if n < len(s.Messages) {
		for _, msg := range s.Messages[n:] {
			delete(s.msgIdx, msg.MessageId)
		}

		s.Messages = s.Messages[n:]
	} else {
		s.Messages = make([]*telegram.Message, 0, 16)
		s.msgIdx = make(map[int]int)
	}
}

// nolint:gocyclo
func (s *session) generateHTMLContent() (msgOutCount int, _ []byte) {
	s.mu.RLock()
	msgCopy := make([]*telegram.Message, 0, len(s.Messages))
	msgCopy = append(msgCopy, s.Messages...)
	s.mu.RUnlock()

	buf := &bytes.Buffer{}
	for _, msg := range msgCopy {
		waitingForCaption := false

		if waitingForCaption {
			waitingForCaption = false
			if msg.Caption != nil {
				buf.WriteString(`<figcaption>` + *msg.Caption + `</figcaption></figure>`)
			} else {
				buf.WriteString(`<figcaption></figcaption></figure>`)
			}
		}

		chartUsername := s.defaultChatUsername
		msgID := int64(msg.MessageId)

		// TODO: should we try to find original link?

		// switch {
		// case msg.SenderChat != nil && msg.SenderChat.Username != nil:
		// 	chartUsername = *msg.SenderChat.Username
		// case msg.ForwardFromChat != nil && msg.ForwardFromChat.Username != nil:
		// 	chartUsername = *msg.ForwardFromChat.Username
		// 	msgID = int64(*msg.ForwardFromMessageId)
		// }

		msgAuthorLink := ``
		switch {
		case msg.ForwardFrom != nil:
			originalUserText := msg.ForwardFrom.FirstName
			{
				if msg.ForwardFrom.LastName != nil {
					originalUserText += " " + *msg.ForwardFrom.LastName
				}

				if msg.ForwardFrom.Username != nil {
					originalUserLink := fmt.Sprintf(`<a href="https://t.me/%s">`, *msg.ForwardFrom.Username)
					msgAuthorLink += originalUserLink + "(forwarded from) " + originalUserText + `</a>`
				} else {
					msgAuthorLink += "(forwarded from) " + originalUserText
				}

				originalChatText := ``
				if fc := msg.ForwardFromChat; fc != nil {
					if fc.FirstName != nil {
						originalChatText += *fc.FirstName
					}

					if fc.LastName != nil {
						if len(originalChatText) != 0 {
							originalChatText += " "
						}

						originalChatText += *fc.LastName
					}

					if len(originalChatText) != 0 {
						originalChatText += " @ "
						if fc.Username != nil {
							originalChatLink := fmt.Sprintf(`<a href="https://t.me/%s">`, *fc.Username)
							msgAuthorLink += originalChatLink + originalChatText + `</a>`
						} else {
							msgAuthorLink += originalChatText
						}
					}
				}
			}

			forwarderUserText := ""
			if msg.From != nil {
				forwarderUserText = msg.From.FirstName
				if msg.From.LastName != nil {
					forwarderUserText += " " + *msg.From.LastName
				}

				if msg.From.Username != nil {
					forwarderUserLink := fmt.Sprintf(`<a href="https://t.me/%s">`, *msg.From.Username)
					msgAuthorLink += forwarderUserLink + " (via) " + forwarderUserText + `</a>`
				} else {
					msgAuthorLink += "(via) " + forwarderUserText
				}
			}
		case msg.From != nil:
			// not a forwarded message
			userText := msg.From.FirstName
			if msg.From.LastName != nil {
				userText += " " + *msg.From.LastName
			}

			if msg.From.Username != nil {
				msgAuthorLink += fmt.Sprintf(`<a href="https://t.me/%s">`, *msg.From.Username) + userText + `</a>`
			} else {
				msgAuthorLink += userText
			}
		}

		buf.WriteString(
			fmt.Sprintf(
				`<p>%s<a href="https://t.me/%s/%s"><br><blockquote>`,
				msgAuthorLink,
				chartUsername,
				strconv.FormatInt(msgID, 10),
			),
		)

		switch {
		case msg.Text != nil:

			// index is the position of plain text content
			index := 0

			text := utf16.Encode([]rune(*msg.Text))
			if msg.Entities != nil {
				for _, e := range *msg.Entities {
					if index < e.Offset {
						buf.WriteString(string(utf16.Decode(text[index:e.Offset])))
					}

					// mark next possible position of plain text
					index = e.Offset + e.Length

					data := string(utf16.Decode(text[e.Offset : e.Offset+e.Length]))

					switch e.Type {
					case telegram.MessageEntityTypeBold:
						buf.WriteString(`<strong>` + data + `</strong>`)
					case telegram.MessageEntityTypeBotCommand:
						// TODO
						buf.WriteString(html.EscapeString(data))
					case telegram.MessageEntityTypeCashtag:
						// TODO
						buf.WriteString(html.EscapeString(data))
					case telegram.MessageEntityTypeCode:
						buf.WriteString(`<code>` + data + `</code>`)
					case telegram.MessageEntityTypeEmail:
						buf.WriteString(fmt.Sprintf(`<a href="mailto:%s">`, data) + data + `</a>`)
					case telegram.MessageEntityTypeHashtag:
						buf.WriteString(`<b>#` + data + `</b>`)
					case telegram.MessageEntityTypeItalic:
						buf.WriteString(`<em>` + data + `</em>`)
					case telegram.MessageEntityTypeMention:
						buf.WriteString(fmt.Sprintf(`<a href="https://t.me/%s">@`, data) + data + `</a>`)
					case telegram.MessageEntityTypePhoneNumber:
					case telegram.MessageEntityTypePre:
						buf.WriteString(`<pre>` + data + `</pre>`)
					case telegram.MessageEntityTypeStrikethrough:
						buf.WriteString(`<del>` + data + `</del>`)
					case telegram.MessageEntityTypeTextLink:
						buf.WriteString(fmt.Sprintf(`<a href="%s">`, data) + data + `</a>`)
					case telegram.MessageEntityTypeTextMention:
						buf.WriteString(fmt.Sprintf(`<a href="https://t.me/%s">@`, data) + data + `</a>`)
					case telegram.MessageEntityTypeUnderline:
						buf.WriteString(`<u>` + data + `</u>`)
					case telegram.MessageEntityTypeUrl:
						// TODO: parse telegraph supported media url
						buf.WriteString(fmt.Sprintf(`<a href="%s">`, data) + data + `</a>`)
					}
				}
			}

			// write tail plain text
			if index < len(text) {
				buf.WriteString(string(utf16.Decode(text[index:])))
			}
		case msg.Audio != nil:
			// TODO: sound to text
			waitingForCaption = true
		case msg.Voice != nil:
			// TODO: sound to text
			waitingForCaption = true
		case msg.Video != nil:
			// TODO
			waitingForCaption = true
			// nolint:lll,gosimple
			buf.WriteString(fmt.Sprintf(`<figure><iframe src="/embed/youtube?url={URL_ESCAPED}" width="640" height="360" frameborder="0" allowtransparency="true" allowfullscreen="true" scrolling="no"></iframe>"`))
		case msg.Photo != nil:
			// TODO
			waitingForCaption = true
		case msg.Animation != nil:
			// TODO
			waitingForCaption = true
		case msg.Document != nil:
			// TODO
			waitingForCaption = true
		case msg.Game != nil:
			// TODO
		case msg.Dice != nil:
			// TODO
		}

		buf.WriteString(`</a></blockquote></p><hr>`)
	}

	return len(msgCopy) + 1, buf.Bytes()
}

func newSessionManager() *SessionManager {
	return &SessionManager{
		sessions: &sync.Map{},
	}
}

type SessionManager struct {
	// chart_id -> session
	sessions *sync.Map
}

func (c *SessionManager) getActiveSession(chartID int64) (*session, bool) {
	sVal, ok := c.sessions.Load(chartID)
	if !ok {
		return nil, false
	}

	return sVal.(*session), true
}

func (c *SessionManager) setSessionState(
	chartID int64, active bool,
	topic, defaultChatUsername string,
	gen generator.Interface,
) (_ *session, ok bool) {
	if active {
		newS := newSession(topic, defaultChatUsername, gen)
		sVal, loaded := c.sessions.LoadOrStore(chartID, newS)
		if !loaded {
			return newS, true
		}

		return sVal.(*session), false
	}

	sVal, loaded := c.sessions.LoadAndDelete(chartID)
	if loaded {
		return sVal.(*session), true
	}

	return nil, false
}
