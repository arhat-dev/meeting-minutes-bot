package server

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unicode/utf16"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"

	"arhat.dev/pkg/hashhelper"
	"arhat.dev/pkg/log"
	"github.com/h2non/filetype"
)

var _ Message = (*telegramMessage)(nil)

func newTelegramMessage(msg *telegram.Message, botUsername string) *telegramMessage {
	return &telegramMessage{
		id: formatTelegramMessageID(msg.MessageId),

		botUsername: botUsername,

		msg: msg,

		entities: make([]generator.MessageEntity, 0, 1),

		ready: 0,

		mu: &sync.Mutex{},
	}
}

func formatTelegramMessageID(msgID int) string {
	return strconv.FormatInt(int64(msgID), 10)
}

type telegramMessage struct {
	id string

	botUsername string

	msg *telegram.Message

	entities []generator.MessageEntity

	ready uint32

	mu *sync.Mutex
}

func (m *telegramMessage) ID() string {
	return m.id
}

func (m *telegramMessage) MessageURL() string {
	url := m.ChatURL()
	if len(url) == 0 {
		return ""
	}

	return url + "/" + formatTelegramMessageID(m.msg.MessageId)
}

func (m *telegramMessage) Timestamp() time.Time {
	// TODO
	return time.Time{}
}

func (m *telegramMessage) ChatName() string {
	var name string

	if cfn := m.msg.Chat.FirstName; cfn != nil {
		name = *cfn
	}

	if cln := m.msg.Chat.LastName; cln != nil {
		name += " " + *cln
	}

	return name
}

func (m *telegramMessage) ChatURL() string {
	if m.IsPrivateMessage() {
		return ""
	}

	if cu := m.msg.Chat.Username; cu != nil {
		return "https://t.me/" + *cu
	}

	return ""
}

func (m *telegramMessage) Author() string {
	if m.msg.From == nil {
		return ""
	}

	name := m.msg.From.FirstName
	if fln := m.msg.From.LastName; fln != nil {
		name += " " + *fln
	}

	return name
}

func (m *telegramMessage) AuthorURL() string {
	if m.msg.From == nil {
		return ""
	}

	if fu := m.msg.From.Username; fu != nil {
		return "https://t.me/" + *fu
	}

	return ""
}

func (m *telegramMessage) IsForwarded() bool {
	return m.msg.ForwardFrom != nil ||
		m.msg.ForwardFromChat != nil ||
		m.msg.ForwardSenderName != nil ||
		m.msg.ForwardFromMessageId != nil
}

func (m *telegramMessage) OriginalMessageURL() string {
	chatURL := m.OriginalChatURL()
	if len(chatURL) == 0 {
		return ""
	}

	if ffmi := m.msg.ForwardFromMessageId; ffmi != nil {
		return chatURL + "/" + strconv.FormatInt(int64(*ffmi), 10)
	}

	return ""
}

func (m *telegramMessage) OriginalChatName() string {
	if fc := m.msg.ForwardFromChat; fc != nil {
		var name string
		if fc.FirstName != nil {
			name += *fc.FirstName
		}

		if fc.LastName != nil {
			name += " " + *fc.LastName
		}

		return name
	}

	return ""
}

func (m *telegramMessage) OriginalChatURL() string {
	if fc := m.msg.ForwardFromChat; fc != nil && fc.Username != nil {
		return "https://t.me/" + *fc.Username
	}

	return ""
}

func (m *telegramMessage) OriginalAuthor() string {
	if ff := m.msg.ForwardFrom; ff != nil {
		name := ff.FirstName
		if ff.LastName != nil {
			name += " " + *ff.LastName
		}

		return name
	}

	return ""
}

func (m *telegramMessage) OriginalAuthorURL() string {
	if ff := m.msg.ForwardFrom; ff != nil && ff.Username != nil {
		return "https://t.me/" + *m.msg.ForwardFrom.Username
	}

	return ""
}

func (m *telegramMessage) IsPrivateMessage() bool {
	return m.msg.Chat.Type == telegram.ChatTypePrivate
}

func (m *telegramMessage) IsReply() bool {
	return m.msg.ReplyToMessage != nil
}

func (m *telegramMessage) ReplyToMessageID() string {
	if m.msg.ReplyToMessage != nil {
		return formatTelegramMessageID(m.msg.ReplyToMessage.MessageId)
	}

	return ""
}

func (m *telegramMessage) Entities() []generator.MessageEntity {
	return m.entities
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
		entities := m.msg.Entities

		if entities == nil {
			m.entities = []generator.MessageEntity{{
				Kind: generator.KindText,
				Text: *m.msg.Text,
			}}
			m.markReady()
			return
		}

		// url -> index of the related entry in `m.entities`
		urlsToArchive := make(map[string][]int)

		text := utf16.Encode([]rune(*m.msg.Text))

		textIndex := 0
		for _, e := range *entities {
			if e.Offset > textIndex {
				// append previous unhandled plain text
				m.entities = append(m.entities, generator.MessageEntity{
					Kind:   generator.KindText,
					Text:   string(utf16.Decode(text[textIndex:e.Offset])),
					Params: nil,
				})
			}

			data := string(utf16.Decode(text[e.Offset : e.Offset+e.Length]))
			textIndex = e.Offset + e.Length

			kind, ok := map[telegram.MessageEntityType]generator.FormatKind{
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
			}[e.Type]

			if ok {
				m.entities = append(m.entities, generator.MessageEntity{
					Kind:   kind,
					Text:   data,
					Params: nil,
				})

				continue
			}

			switch e.Type {
			case telegram.MessageEntityTypeUrl:
				m.entities = append(m.entities, generator.MessageEntity{
					Kind: generator.KindURL,
					Text: data,
					Params: map[generator.EntityParamKind]string{
						generator.EntityParamURL:                     data,
						generator.EntityParamWebArchiveURL:           "",
						generator.EntityParamWebArchiveScreenshotURL: "",
					},
				})

				urlsToArchive[data] = append(urlsToArchive[data], len(m.entities)-1)
			case telegram.MessageEntityTypeMention, telegram.MessageEntityTypeTextMention:
				url := "https://t.me/" + strings.TrimPrefix(data, "@")
				m.entities = append(m.entities, generator.MessageEntity{
					Kind: generator.KindURL,
					Text: data,
					Params: map[generator.EntityParamKind]string{
						generator.EntityParamURL:                     url,
						generator.EntityParamWebArchiveURL:           "",
						generator.EntityParamWebArchiveScreenshotURL: "",
					},
				})

				// TODO: do we really need to archive user page? (while reasonable to me)
				urlsToArchive[data] = append(urlsToArchive[data], len(m.entities)-1)
			default:
				client.logger.E("message entity unhandled", log.String("type", string(e.Type)))
			}
		}

		if textIndex < len(text)-1 {
			m.entities = append(m.entities, generator.MessageEntity{
				Kind:   generator.KindText,
				Text:   string(utf16.Decode(text[textIndex:])),
				Params: nil,
			})
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

			// url -> archive url
			archiveURLs := make(map[string]string)
			// url -> screen shot url
			screenshotURLs := make(map[string]string)
			for url, indexes := range urlsToArchive {
				if indexes == nil {
					urlsToArchive[url] = make([]int, 0, 1)
				}

				archiveURL, screenshot, ext, err := w.Archive(url)
				if err != nil {
					select {
					case errCh <- fmt.Errorf("Internal bot error: unable to archive web page %s: %w", url, err):
					case <-client.ctx.Done():
					}

					continue
				}

				archiveURLs[url] = archiveURL

				if len(screenshot) == 0 {
					// no screenshot
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
				for url, idxes := range urlsToArchive {
					for _, idx := range idxes {
						m.entities[idx].Params[generator.EntityParamWebArchiveURL] = archiveURLs[url]
						m.entities[idx].Params[generator.EntityParamWebArchiveScreenshotURL] = screenshotURLs[url]
					}
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

		m.entities = append(m.entities, generator.MessageEntity{
			Kind: generator.KindAudio,
			Text: "",
			Params: map[generator.EntityParamKind]string{
				generator.EntityParamURL:     "",
				generator.EntityParamCaption: "",
			},
		})
	case m.msg.Document != nil:
		doc := m.msg.Document

		requestFileID = doc.FileId
		if doc.FileName != nil {
			requestFileName = *doc.FileName
		}

		m.entities = append(m.entities, generator.MessageEntity{
			Kind: generator.KindDocument,
			Text: "",
			Params: map[generator.EntityParamKind]string{
				generator.EntityParamURL:     "",
				generator.EntityParamCaption: "",
			},
		})
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

		m.entities = append(m.entities, generator.MessageEntity{
			Kind: generator.KindImage,
			Text: "",
			Params: map[generator.EntityParamKind]string{
				generator.EntityParamURL:     "",
				generator.EntityParamCaption: "",
			},
		})
	case m.msg.Video != nil:
		video := m.msg.Video

		requestFileID = video.FileId
		if video.FileName != nil {
			requestFileName = *video.FileName
		}

		m.entities = append(m.entities, generator.MessageEntity{
			Kind: generator.KindVideo,
			Text: "",
			Params: map[generator.EntityParamKind]string{
				generator.EntityParamURL:     "",
				generator.EntityParamCaption: "",
			},
		})
	case m.msg.Voice != nil:
		// TODO: sound to text
		voice := m.msg.Voice
		requestFileID = voice.FileId

		m.entities = append(m.entities, generator.MessageEntity{
			Kind: generator.KindVideo,
			Text: "",
			Params: map[generator.EntityParamKind]string{
				generator.EntityParamURL:     "",
				generator.EntityParamCaption: "",
			},
		})
	case m.msg.VideoNote != nil:
		// TODO
		m.markReady()
		return nil, nil
	case m.msg.Animation != nil:
		m.markReady()
		return nil, nil
	case m.msg.Sticker != nil,
		m.msg.Dice != nil,
		m.msg.Game != nil:
		// TODO: shall we just ignore them
		m.markReady()
		return nil, nil
	case m.msg.Poll != nil:
		// TODO
		m.markReady()
		return nil, nil
	case m.msg.Venue != nil:
		// TODO
		m.markReady()
		return nil, nil
	case m.msg.Location != nil:
		// TODO
		m.markReady()
		return nil, nil
	default:
		client.logger.E("unhandled telegram message", log.Any("msg", m.msg))
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
			m.entities[0].Params[generator.EntityParamURL] = fileURL
			m.entities[0].Params[generator.EntityParamFilename] = requestFileName
		})
	}()

	return errCh, nil
}
