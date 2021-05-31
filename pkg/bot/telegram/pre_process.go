package telegram

import (
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	"sync"

	"arhat.dev/pkg/hashhelper"
	"arhat.dev/pkg/log"
	"github.com/h2non/filetype"

	"arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

// errCh can be nil when there is no background pre-process worker
// nolint:gocyclo
func (c *telegramBot) preProcess(
	m *telegramMessage,
	w webarchiver.Interface,
	u storage.Interface,
) (errCh chan error, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		requestFileID   string
		requestFileName string // can be empty
	)

	switch {
	case m.msg.Text != nil:
		me := parseTelegramEntities(*m.msg.Text, m.msg.Entities)

		if !me.NeedPreProcess() {
			m.entities = me.Get()

			m.markReady()

			return nil, nil
		}

		errCh = make(chan error, 1)
		go func() {
			defer func() {
				close(errCh)

				m.markReady()
			}()

			err2 := me.PreProcess(c.ctx, w, u)
			if err2 != nil {
				select {
				case errCh <- fmt.Errorf("Message pre-process error: %w", err2):
				case <-c.ctx.Done():
				}
			}

			m.update(func() {
				m.entities = me.Get()
			})
		}()

		return
	case m.msg.Audio != nil:
		audio := m.msg.Audio

		requestFileID = audio.FileId
		if audio.FileName != nil {
			requestFileName = *audio.FileName
		}

		m.entities = append(m.entities, message.Entity{
			Kind: message.KindAudio,
			Text: "",
			Params: map[message.EntityParamKey]interface{}{
				message.EntityParamURL:     "",
				message.EntityParamCaption: nil,
			},
		})
	case m.msg.Document != nil:
		doc := m.msg.Document

		requestFileID = doc.FileId
		if doc.FileName != nil {
			requestFileName = *doc.FileName
		}

		m.entities = append(m.entities, message.Entity{
			Kind: message.KindFile,
			Text: "",
			Params: map[message.EntityParamKey]interface{}{
				message.EntityParamURL:     "",
				message.EntityParamCaption: nil,
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

		m.entities = append(m.entities, message.Entity{
			Kind: message.KindImage,
			Text: "",
			Params: map[message.EntityParamKey]interface{}{
				message.EntityParamURL:     "",
				message.EntityParamCaption: nil,
			},
		})
	case m.msg.Video != nil:
		video := m.msg.Video

		requestFileID = video.FileId
		if video.FileName != nil {
			requestFileName = *video.FileName
		}

		m.entities = append(m.entities, message.Entity{
			Kind: message.KindVideo,
			Text: "",
			Params: map[message.EntityParamKey]interface{}{
				message.EntityParamURL:     "",
				message.EntityParamCaption: nil,
			},
		})
	case m.msg.Voice != nil:
		// TODO: sound to text
		voice := m.msg.Voice
		requestFileID = voice.FileId

		m.entities = append(m.entities, message.Entity{
			Kind: message.KindVideo,
			Text: "",
			Params: map[message.EntityParamKey]interface{}{
				message.EntityParamURL:     "",
				message.EntityParamCaption: nil,
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
		c.logger.E("unhandled telegram message", log.Any("msg", m.msg))
		m.markReady()
		return nil, nil
	}

	errCh = make(chan error, 1)
	wg := &sync.WaitGroup{}

	if m.msg.Caption != nil {
		cme := parseTelegramEntities(*m.msg.Caption, m.msg.CaptionEntities)

		wg.Add(1)
		if cme.NeedPreProcess() {
			go func() {
				defer wg.Done()

				err2 := cme.PreProcess(c.ctx, w, u)
				if err2 != nil {
					select {
					case errCh <- fmt.Errorf("Caption pre-process error: %w", err2):
					case <-c.ctx.Done():
					}
				}

				m.update(func() {
					m.entities[0].Params[message.EntityParamCaption] = cme.Get()
				})
			}()
		}
	}

	wg.Add(1)

	go func() {
		wg.Wait()

		close(errCh)

		m.markReady()
	}()

	go func() {
		defer wg.Done()

		logger := c.logger.WithFields(
			log.String("filename", requestFileName),
			log.String("id", requestFileID),
		)

		logger.V("requesting file info")
		resp, err := c.client.PostGetFile(c.ctx, telegram.PostGetFileJSONRequestBody{
			FileId: requestFileID,
		})
		if err != nil {
			logger.I("failed to request get file", log.Error(err))

			select {
			case errCh <- fmt.Errorf("failed to request file: %w", err):
			case <-c.ctx.Done():
			}

			return
		}

		f, err := telegram.ParsePostGetFileResponse(resp)
		_ = resp.Body.Close()
		if err != nil {
			logger.I("failed to parse get file response", log.Error(err))

			select {
			case errCh <- fmt.Errorf("failed to parse file response: %w", err):
			case <-c.ctx.Done():
			}

			return
		}

		if f.JSON200 == nil || !f.JSON200.Ok {
			logger.I("telegram get file error", log.String("reason", f.JSONDefault.Description))

			select {
			case errCh <- fmt.Errorf("failed to request file: telegram: %s", f.JSONDefault.Description):
			case <-c.ctx.Done():
			}

			return
		}

		pathPtr := f.JSON200.Result.FilePath
		if pathPtr == nil {
			logger.I("telegram file path not found")

			select {
			case errCh <- fmt.Errorf("Invalid empty path for telegram file downloading"):
			case <-c.ctx.Done():
			}

			return
		}

		downloadURL := fmt.Sprintf(
			"https://api.telegram.org/file/bot%s/%s",
			c.opts.BotToken, *pathPtr,
		)

		req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
		if err != nil {
			logger.I("failed to create bot file download request", log.Error(err))

			select {
			case errCh <- fmt.Errorf("Internal bot error: %w", err):
			case <-c.ctx.Done():
			}

			return
		}

		resp, err = c.client.Client.Do(req)
		if err != nil {
			logger.I("failed to request file download", log.Error(err))

			select {
			case errCh <- fmt.Errorf("Internal bot error: %w", err):
			case <-c.ctx.Done():
			}

			return
		}
		defer func() { _ = resp.Body.Close() }()

		fileContent, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logger.I("failed to download file", log.Error(err))

			select {
			case errCh <- fmt.Errorf("Internal bot error: failed to read file body: %w", err):
			case <-c.ctx.Done():
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
		fileURL, err := u.Upload(c.ctx, filename, fileContent)
		if err != nil {
			select {
			case errCh <- fmt.Errorf("failed to upload file: %w", err):
			case <-c.ctx.Done():
			}

			return
		}

		logger.V("file uploaded", log.String("url", fileURL))

		m.update(func() {
			m.entities[0].Params[message.EntityParamURL] = fileURL
			m.entities[0].Params[message.EntityParamFilename] = requestFileName
			m.entities[0].Params[message.EntityParamData] = fileContent
		})
	}()

	return errCh, nil
}
