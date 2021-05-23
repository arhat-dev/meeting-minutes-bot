package server

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf16"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/fileuploader"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"

	"arhat.dev/pkg/hashhelper"
	"github.com/h2non/filetype"
)

var _ Message = (*telegramMessage)(nil)

func newTelegramMessage(msg *telegram.Message) *telegramMessage {
	return &telegramMessage{
		id:   formatTelegramMessageID(msg.MessageId),
		msg:  msg,
		urls: make(map[string]string),

		fileURLs:    make(map[string]string),
		fileCaption: "",

		ready: 0,

		mu: &sync.RWMutex{},
	}
}

func formatTelegramMessageID(msgID int) string {
	return strconv.FormatInt(int64(msgID), 10)
}

type telegramMessage struct {
	id string

	msg *telegram.Message

	// text -> url or url -> archive url
	urls map[string]string

	// file id -> file url
	fileURLs    map[string]string
	fileCaption string // updated by the next message

	ready uint32

	mu *sync.RWMutex
}

func (m *telegramMessage) ID() string {
	return m.id
}

// ready
func (m *telegramMessage) Ready() bool {
	return atomic.LoadUint32(&m.ready) == 1
}

func (m *telegramMessage) markReady() {
	atomic.StoreUint32(&m.ready, 1)
}

func (m *telegramMessage) update(do func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	do()
}

// nolint:gocyclo
func (m *telegramMessage) PreProcess(
	c Client,
	w webarchiver.Interface,
	u fileuploader.Interface,
	previousMessage Message,
) (errCh chan error, err error) {
	client, ok := c.(*telegramBot)
	if !ok {
		return nil, fmt.Errorf("Unexpected client type: need telegram bot")
	}

	type fileRequest struct {
		id   string
		name string // can be empty
	}

	var (
		requestFiles []*fileRequest
	)

	switch {
	case m.msg.Text != nil, m.msg.Caption != nil:
		var (
			text     []uint16
			entities *[]telegram.MessageEntity
		)

		if m.msg.Text != nil {
			text = utf16.Encode([]rune(*m.msg.Text))
			entities = m.msg.Entities
		} else {
			if previousMessage != nil {
				prevMsg := previousMessage.(*telegramMessage)
				prevMsg.update(func() {
					prevMsg.fileCaption = *m.msg.Caption
				})
			}
			text = utf16.Encode([]rune(*m.msg.Caption))
			entities = m.msg.CaptionEntities
		}

		if entities == nil {
			m.markReady()
			return
		}

		var (
			urlsToArchive []string
		)

		urls := make(map[string]string)

		for _, e := range *entities {
			data := string(utf16.Decode(text[e.Offset : e.Offset+e.Length]))

			switch e.Type {
			case telegram.MessageEntityTypeUrl:
				urlsToArchive = append(urlsToArchive, data)
			case telegram.MessageEntityTypeMention, telegram.MessageEntityTypeTextMention:
				urls[data] = fmt.Sprintf("https://t.me/%s", strings.TrimPrefix(data, "@"))
			default:
				// TODO: check other type of text message we can pre-process
			}
		}

		if len(urlsToArchive) == 0 {
			m.markReady()

			return
		}

		errCh = make(chan error, 1)
		go func() {
			defer func() {
				close(errCh)

				m.markReady()
			}()

			for _, url := range urlsToArchive {
				archiveURL, err := w.Archive(url)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("Internal bot error: unable to archive web page %s: %w", url, err):
					case <-client.ctx.Done():
					}

					continue
				}

				urls[url] = archiveURL
			}

			m.update(func() {
				for k, v := range urls {
					m.urls[k] = v
				}
			})
		}()

		return
	case m.msg.Audio != nil:
		audio := m.msg.Audio
		req := &fileRequest{
			id: audio.FileId,
		}

		if audio.FileName != nil {
			req.name = *audio.FileName
		}

		requestFiles = append(requestFiles, req)
	case m.msg.Document != nil:
		doc := m.msg.Document
		req := &fileRequest{
			id: doc.FileId,
		}

		if doc.FileName != nil {
			req.name = *doc.FileName
		}

		requestFiles = append(requestFiles, req)
	case m.msg.Photo != nil:
		for _, photo := range *m.msg.Photo {
			req := &fileRequest{
				id: photo.FileId,
			}

			requestFiles = append(requestFiles, req)
		}
	case m.msg.Video != nil:
		video := m.msg.Video
		req := &fileRequest{
			id: video.FileId,
		}

		if video.FileName != nil {
			req.name = *video.FileName
		}

		requestFiles = append(requestFiles, req)
	case m.msg.Voice != nil:
		// TODO: sound to text
		voice := m.msg.Video
		req := &fileRequest{
			id: voice.FileId,
		}

		if voice.FileName != nil {
			req.name = *voice.FileName
		}

		requestFiles = append(requestFiles, req)
	default:
		// nothing to pre-process
		m.markReady()
		return nil, nil
	}

	errCh = make(chan error, len(requestFiles))

	wg := &sync.WaitGroup{}
	wg.Add(len(requestFiles))
	go func() {
		wg.Wait()

		close(errCh)

		m.markReady()
	}()

	for _, fr := range requestFiles {
		go func(fr fileRequest) {
			defer wg.Done()

			resp, err := client.client.PostGetFile(client.ctx, telegram.PostGetFileJSONRequestBody{
				FileId: fr.id,
			})
			if err != nil {
				select {
				case errCh <- fmt.Errorf("failed to request file: %w", err):
				case <-client.ctx.Done():
				}

				return
			}

			f, err := telegram.ParsePostGetFileResponse(resp)
			_ = resp.Body.Close()
			if err != nil {
				select {
				case errCh <- fmt.Errorf("failed to parse file response: %w", err):
				case <-client.ctx.Done():
				}

				return
			}

			if f.JSON200 == nil || !f.JSON200.Ok {
				select {
				case errCh <- fmt.Errorf("failed to request file: telegram: %s", f.JSONDefault.Description):
				case <-client.ctx.Done():
				}

				return
			}

			pathPtr := f.JSON200.Result.FilePath
			if pathPtr == nil {
				select {
				case errCh <- fmt.Errorf("Invalid empty path for telegram file downloading"):
				case <-client.ctx.Done():
				}

				return
			}

			downloadURL := fmt.Sprintf(
				"https://api.telegram.org/file/bot%s/%s",
				client.botToken, *pathPtr,
			)

			req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("Internal bot error: %w", err):
				case <-client.ctx.Done():
				}

				return
			}

			resp, err = client.client.Client.Do(req)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("Internal bot error: %w", err):
				case <-client.ctx.Done():
				}

				return
			}
			defer func() { _ = resp.Body.Close() }()

			fileContent, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("Internal bot error: failed to read file body: %w", err):
				case <-client.ctx.Done():
				}

				return
			}

			// TODO: upload file content to s3
			_ = fileContent

			var fileExt string
			filename := hex.EncodeToString(hashhelper.Sha256Sum(fileContent))
			if len(fr.name) != 0 {
				fileExt = filepath.Ext(fr.name)
			}

			if len(fileExt) == 0 {
				t, err2 := filetype.Match(fileContent)
				if err2 == nil {
					fileExt = t.Extension
				}
			}

			filename += fileExt
			fileURL, err := u.Upload(filename, fileContent)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("failed to upload file: %w", err):
				case <-client.ctx.Done():
				}

				return
			}

			m.update(func() {
				m.fileURLs[fr.id] = fileURL
			})
		}(*fr)
	}

	return errCh, nil
}

func (m *telegramMessage) formatText(
	fm generator.Formatter,
	content string,
	entities *[]telegram.MessageEntity,
	buf io.StringWriter,
) {
	// index is the position of plain text content
	index := 0

	text := utf16.Encode([]rune(content))

	msgEntityKindMapping := map[telegram.MessageEntityType]generator.FormatKind{
		telegram.MessageEntityTypeBotCommand: generator.KindText,
		telegram.MessageEntityTypeHashtag:    generator.KindText,
		telegram.MessageEntityTypeCashtag:    generator.KindText,
		telegram.MessageEntityTypeTextLink:   generator.KindText,

		telegram.MessageEntityTypeBold:          generator.KindBold,
		telegram.MessageEntityTypeItalic:        generator.KindItalic,
		telegram.MessageEntityTypeStrikethrough: generator.KindStrikethrough,
		telegram.MessageEntityTypeUnderline:     generator.KindUnderline,
		telegram.MessageEntityTypeCode:          generator.KindCode,
		telegram.MessageEntityTypePre:           generator.KindPre,

		telegram.MessageEntityTypeEmail:       generator.KindEmail,
		telegram.MessageEntityTypePhoneNumber: generator.KindPhoneNumber,
		telegram.MessageEntityTypeUrl:         generator.KindURL,

		// need pre-process
		telegram.MessageEntityTypeMention:     generator.KindURL,
		telegram.MessageEntityTypeTextMention: generator.KindURL,
	}

	var appendArchiveURLs []string

	if entities != nil {
		for _, e := range *entities {
			if index < e.Offset {
				_, _ = buf.WriteString(string(utf16.Decode(text[index:e.Offset])))
			}

			// mark next possible position of plain text
			index = e.Offset + e.Length

			data := string(utf16.Decode(text[e.Offset : e.Offset+e.Length]))

			var params []string
			kind, ok := msgEntityKindMapping[e.Type]
			if !ok {
				// unsupported type
				_, _ = buf.WriteString(data)
				continue
			}

			switch e.Type {
			case telegram.MessageEntityTypeMention, telegram.MessageEntityTypeTextMention:
				// find pre-processed url
				urlVal, ok := m.urls[data]
				if ok {
					params = append(params, urlVal)
				}
			case telegram.MessageEntityTypeUrl:
				// append as appendix
				appendArchiveURLs = append(appendArchiveURLs, data)
			}

			_, _ = buf.WriteString(fm.Format(kind, data, params...))
		}
	}

	// write tail plain text
	if index < len(text) {
		_, _ = buf.WriteString(string(utf16.Decode(text[index:])))
	}

	if len(appendArchiveURLs) != 0 {
		_, _ = buf.WriteString(fm.Format(generator.KindNewLine, ""))
		for _, url := range appendArchiveURLs {
			archiveURLVal, ok := m.urls[url]
			if ok && len(archiveURLVal) != 0 {
				_, _ = buf.WriteString(fm.Format(generator.KindURL, "* Archive of "+url, archiveURLVal))
			} else {
				_, _ = buf.WriteString(fm.Format(generator.KindText, fmt.Sprintf("* Archive of %s missing", url), ""))
			}
		}
	}
}

func (m *telegramMessage) Format(fm generator.Formatter) []byte {
	buf := &bytes.Buffer{}

	msgAuthorLink := ``
	switch {
	case m.msg.ForwardFrom != nil:
		originalUserText := m.msg.ForwardFrom.FirstName
		{
			if m.msg.ForwardFrom.LastName != nil {
				originalUserText += " " + *m.msg.ForwardFrom.LastName
			}

			if m.msg.ForwardFrom.Username != nil {
				msgAuthorLink += fm.Format(
					generator.KindURL,
					"(forwarded from) "+originalUserText,
					fmt.Sprintf("https://t.me/%s", *m.msg.ForwardFrom.Username),
				)
			} else {
				msgAuthorLink += "(forwarded from) " + originalUserText
			}

			originalChatText := ``
			if fc := m.msg.ForwardFromChat; fc != nil {
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
						msgAuthorLink += fm.Format(
							generator.KindURL,
							originalChatText,
							fmt.Sprintf("https://t.me/%s", *fc.Username),
						)
					} else {
						msgAuthorLink += originalChatText
					}
				}
			}
		}

		forwarderUserText := ""
		if m.msg.From != nil {
			forwarderUserText = m.msg.From.FirstName
			if m.msg.From.LastName != nil {
				forwarderUserText += " " + *m.msg.From.LastName
			}

			if m.msg.From.Username != nil {
				msgAuthorLink += fm.Format(
					generator.KindURL,
					" (via) "+forwarderUserText,
					fmt.Sprintf("https://t.me/%s", *m.msg.From.Username),
				)
			} else {
				msgAuthorLink += "(via) " + forwarderUserText
			}
		}
	case m.msg.From != nil:
		// not a forwarded message
		userText := m.msg.From.FirstName
		if m.msg.From.LastName != nil {
			userText += " " + *m.msg.From.LastName
		}

		if m.msg.From.Username != nil {
			msgAuthorLink += fm.Format(
				generator.KindURL,
				userText,
				fmt.Sprintf("https://t.me/%s", *m.msg.From.Username),
			)
		} else {
			msgAuthorLink += userText
		}
	}

	// write author link
	_, _ = buf.WriteString(msgAuthorLink + "\n")

	switch {
	case m.msg.Text != nil:
		m.formatText(fm, *m.msg.Text, m.msg.Entities, buf)
	case m.msg.Audio != nil:
		// waitingForCaption = true
	case m.msg.Document != nil:
		// waitingForCaption = true
	case m.msg.Photo != nil:
		// waitingForCaption = true
	case m.msg.Video != nil:
		// waitingForCaption = true
	case m.msg.Voice != nil:
		// waitingForCaption = true
	case m.msg.Caption != nil:
		m.formatText(fm, *m.msg.Caption, m.msg.CaptionEntities, buf)
	case m.msg.VideoNote != nil:
	case m.msg.Animation != nil:
		// waitingForCaption = true
	case m.msg.Sticker != nil,
		m.msg.Dice != nil,
		m.msg.Game != nil:
		// TODO: shall we just ignore them
		// TODO
	case m.msg.Poll != nil:
		// TODO
	case m.msg.Venue != nil:
		// TODO
	case m.msg.Location != nil:
		// TODO
	}

	return buf.Bytes()
}
