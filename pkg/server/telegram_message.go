package server

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unicode/utf16"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/generator/telegraph"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"

	"arhat.dev/pkg/hashhelper"
	"arhat.dev/pkg/log"
	"github.com/h2non/filetype"
)

var _ Message = (*telegramMessage)(nil)

func newTelegramMessage(msg *telegram.Message) *telegramMessage {
	return &telegramMessage{
		id:   formatTelegramMessageID(msg.MessageId),
		msg:  msg,
		urls: make(map[string]string),

		fileURL: "",

		ready: 0,

		mu: &sync.Mutex{},
	}
}

func formatTelegramMessageID(msgID int) string {
	return strconv.FormatInt(int64(msgID), 10)
}

type telegramMessage struct {
	id string

	msg *telegram.Message

	// text -> url (named links)
	// or
	// url -> archive url (for archived pages)
	urls map[string]string

	// url -> image url (for archived pages)
	archiveScreenshotURLs map[string]string

	// file url
	fileURL  string
	fileName string

	ready uint32

	mu *sync.Mutex
}

func (m *telegramMessage) ID() string {
	return m.id
}

// ready for content generation
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
	u storage.Interface,
	previousMessage Message,
) (errCh chan error, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := c.(*telegramBot)
	if !ok {
		return nil, fmt.Errorf("Unexpected client type: need telegram bot")
	}

	var (
		requestFileID   string
		requestFileName string // can be empty
	)

	switch {
	case m.msg.Text != nil:
		var (
			text     []uint16
			entities *[]telegram.MessageEntity
		)

		if m.msg.Text != nil {
			text = utf16.Encode([]rune(*m.msg.Text))
			entities = m.msg.Entities
		} else {
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

			screenshotURLs := make(map[string]string)
			for _, url := range urlsToArchive {
				archiveURL, screenshot, ext, err := w.Archive(url)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("Internal bot error: unable to archive web page %s: %w", url, err):
					case <-client.ctx.Done():
					}

					continue
				}

				urls[url] = archiveURL

				if len(screenshot) == 0 {
					// no screenshot created
					continue
				}

				filename := hex.EncodeToString(hashhelper.Sha256Sum(screenshot)) + ext
				screenshotURL, err := u.Upload(client.ctx, filename, screenshot)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("unable to upload web page screenshot: %w", err):
					case <-client.ctx.Done():
					}

					continue
				}

				screenshotURLs[url] = screenshotURL
			}

			m.update(func() {
				for k, v := range urls {
					m.urls[k] = v
				}

				for k, v := range screenshotURLs {
					m.archiveScreenshotURLs[k] = v
				}
			})
		}()

		return
	case m.msg.Audio != nil:
		audio := m.msg.Audio

		requestFileID = audio.FileId
		if audio.FileName != nil {
			requestFileName = *audio.FileName
		}
	case m.msg.Document != nil:
		doc := m.msg.Document

		requestFileID = doc.FileId
		if doc.FileName != nil {
			requestFileName = *doc.FileName
		}
	case m.msg.Photo != nil:
		var (
			maxSize int
			id      string
		)

		for _, photo := range *m.msg.Photo {
			if size := photo.FileSize; size != nil {
				if *size > maxSize {
					id = photo.FileId
				}
			}
		}

		requestFileID = id
	case m.msg.Video != nil:
		video := m.msg.Video

		requestFileID = video.FileId
		if video.FileName != nil {
			requestFileName = *video.FileName
		}
	case m.msg.Voice != nil:
		// TODO: sound to text
		voice := m.msg.Voice
		requestFileID = voice.FileId
	default:
		// nothing to pre-process
		m.markReady()
		return nil, nil
	}

	errCh = make(chan error, 1)
	go func() {
		defer func() {
			close(errCh)

			m.markReady()
		}()

		logger := client.logger.WithFields(
			log.String("filename", requestFileName),
			log.String("id", requestFileID),
		)

		logger.V("requesting file info")
		resp, err := client.client.PostGetFile(client.ctx, telegram.PostGetFileJSONRequestBody{
			FileId: requestFileID,
		})
		if err != nil {
			logger.I("failed to request get file", log.Error(err))

			select {
			case errCh <- fmt.Errorf("failed to request file: %w", err):
			case <-client.ctx.Done():
			}

			return
		}

		f, err := telegram.ParsePostGetFileResponse(resp)
		_ = resp.Body.Close()
		if err != nil {
			logger.I("failed to parse get file response", log.Error(err))

			select {
			case errCh <- fmt.Errorf("failed to parse file response: %w", err):
			case <-client.ctx.Done():
			}

			return
		}

		if f.JSON200 == nil || !f.JSON200.Ok {
			logger.I("telegram get file error", log.String("reason", f.JSONDefault.Description))

			select {
			case errCh <- fmt.Errorf("failed to request file: telegram: %s", f.JSONDefault.Description):
			case <-client.ctx.Done():
			}

			return
		}

		pathPtr := f.JSON200.Result.FilePath
		if pathPtr == nil {
			logger.I("telegram file path not found")

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
			logger.I("failed to create bot file download request", log.Error(err))

			select {
			case errCh <- fmt.Errorf("Internal bot error: %w", err):
			case <-client.ctx.Done():
			}

			return
		}

		resp, err = client.client.Client.Do(req)
		if err != nil {
			logger.I("failed to request file download", log.Error(err))

			select {
			case errCh <- fmt.Errorf("Internal bot error: %w", err):
			case <-client.ctx.Done():
			}

			return
		}
		defer func() { _ = resp.Body.Close() }()

		fileContent, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.I("failed to download file", log.Error(err))

			select {
			case errCh <- fmt.Errorf("Internal bot error: failed to read file body: %w", err):
			case <-client.ctx.Done():
			}

			return
		}

		var fileExt string
		filename := hex.EncodeToString(hashhelper.Sha256Sum(fileContent))
		if len(requestFileName) != 0 {
			fileExt = path.Ext(requestFileName)
		}

		if len(fileExt) == 0 {
			t, err2 := filetype.Match(fileContent)
			if err2 == nil {
				fileExt = "." + t.Extension
			}
		}

		filename += fileExt
		logger.V("uploading file",
			log.String("upload_name", filename),
			log.Int("size", len(fileContent)),
		)
		fileURL, err := u.Upload(client.ctx, filename, fileContent)
		if err != nil {
			select {
			case errCh <- fmt.Errorf("failed to upload file: %w", err):
			case <-client.ctx.Done():
			}

			return
		}

		logger.V("file uploaded", log.String("url", fileURL))

		m.update(func() {
			m.fileURL = fileURL
			m.fileName = requestFileName
		})
	}()

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
			}

			_, _ = buf.WriteString(fm.Format(kind, data, params...))

			switch e.Type {
			case telegram.MessageEntityTypeUrl:
				archiveURL, hasArchiveURL := m.urls[data]
				if hasArchiveURL && len(archiveURL) != 0 {
					_, _ = buf.WriteString(fm.Format(generator.KindURL, " [archive]", archiveURL))
				} else {
					_, _ = buf.WriteString(fm.Format(generator.KindText, " [archive missing]"))
				}

				screenshotURL, hasScreenshotURL := m.archiveScreenshotURLs[data]
				if hasScreenshotURL && len(screenshotURL) != 0 {
					_, _ = buf.WriteString(fm.Format(generator.KindURL, " [screenshot]", screenshotURL))
				} else {
					_, _ = buf.WriteString(fm.Format(generator.KindText, " [screenshot missing]"))
				}
			default:
				// other cases
			}
		}
	}

	// write tail plain text
	if index < len(text) {
		_, _ = buf.WriteString(string(utf16.Decode(text[index:])))
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
	case m.msg.Audio != nil, m.msg.Voice != nil:
		_, _ = buf.WriteString(
			fm.Format(
				generator.KindAudio, m.fileURL, m.formatFileCaptionText(fm),
			),
		)
	case m.msg.Document != nil:
		linkText := `[File]`
		if len(m.fileName) != 0 {
			linkText += " " + m.fileName
		}

		if m.msg.Caption != nil {
			linkText += " (" + m.formatFileCaptionText(fm) + ")"
		}

		_, _ = buf.WriteString(
			fm.Format(generator.KindURL, linkText, m.fileURL),
		)
	case m.msg.Photo != nil:
		_, _ = buf.WriteString(
			fm.Format(
				generator.KindImage, m.fileURL, m.formatFileCaptionText(fm),
			),
		)
	case m.msg.Video != nil:
		_, _ = buf.WriteString(
			fm.Format(
				generator.KindVideo, m.fileURL, m.formatFileCaptionText(fm),
			),
		)
	case m.msg.Caption != nil:
		// should have been consumed by audio/photo/video
	case m.msg.VideoNote != nil:
		// TODO
	case m.msg.Animation != nil:
		// waitingForCaption = true
	case m.msg.Sticker != nil,
		m.msg.Dice != nil,
		m.msg.Game != nil:
		// TODO: shall we just ignore them
	case m.msg.Poll != nil:
		// TODO
	case m.msg.Venue != nil:
		// TODO
	case m.msg.Location != nil:
		// TODO
	}

	return buf.Bytes()
}

func (m *telegramMessage) formatFileCaptionText(fm generator.Formatter) string {
	switch fm.Name() {
	case telegraph.Name:
		// telegraph doesn't support html tags in caption area
		return fm.Format(generator.KindText, *m.msg.Caption)
	default:
		var caption string
		if m.msg.Caption != nil {
			captionBuf := &bytes.Buffer{}
			m.formatText(fm, *m.msg.Caption, m.msg.CaptionEntities, captionBuf)
			caption = captionBuf.String()
		}
		return caption
	}
}
