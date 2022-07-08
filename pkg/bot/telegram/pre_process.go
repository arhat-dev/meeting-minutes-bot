package telegram

import (
	"fmt"
	"net/http"
	"path"
	"sync"

	"arhat.dev/pkg/log"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	api "arhat.dev/meeting-minutes-bot/pkg/botapis/telegram"
	"arhat.dev/meeting-minutes-bot/pkg/message"
)

// errCh can be nil when there is no background pre-process worker
// nolint:gocyclo
func (c *telegramBot) preProcess(
	wf *bot.Workflow,
	m *Message,
) (errCh chan error, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		requestFileID          string
		requestFileName        string // can be empty
		requestFileContentType string
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

			err2 := me.PreProcess(c.Context(), wf.WebArchiver, wf.Storage)
			if err2 != nil {
				select {
				case errCh <- fmt.Errorf("Message pre-process error: %w", err2):
				case <-c.Context().Done():
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

		if audio.MimeType != nil {
			requestFileContentType = *audio.MimeType
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

		if doc.MimeType != nil {
			requestFileContentType = *doc.MimeType
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

		if video.MimeType != nil {
			requestFileContentType = *video.MimeType
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

		if voice.MimeType != nil {
			requestFileContentType = *voice.MimeType
		}

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
		// TODO: currently we just ignore them
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
		c.Logger().E("unhandled telegram message", log.Any("msg", m.msg))
		m.markReady()
		return nil, nil
	}

	errCh = make(chan error, 1)
	var wg sync.WaitGroup

	if m.msg.Caption != nil {
		cme := parseTelegramEntities(*m.msg.Caption, m.msg.CaptionEntities)

		wg.Add(1)
		if cme.NeedPreProcess() {
			go func() {
				defer wg.Done()

				err2 := cme.PreProcess(c.Context(), wf.WebArchiver, wf.Storage)
				if err2 != nil {
					select {
					case errCh <- fmt.Errorf("Caption pre-process error: %w", err2):
					case <-c.Context().Done():
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

		logger := c.Logger().WithFields(
			log.String("filename", requestFileName),
			log.String("id", requestFileID),
		)

		logger.V("requesting file info")
		resp, err := c.client.PostGetFile(c.Context(), api.PostGetFileJSONRequestBody{
			FileId: requestFileID,
		})
		if err != nil {
			logger.I("failed to request get file", log.Error(err))

			select {
			case errCh <- fmt.Errorf("failed to request file: %w", err):
			case <-c.Context().Done():
			}

			return
		}

		f, err := api.ParsePostGetFileResponse(resp)
		_ = resp.Body.Close()
		if err != nil {
			logger.I("failed to parse get file response", log.Error(err))

			select {
			case errCh <- fmt.Errorf("failed to parse file response: %w", err):
			case <-c.Context().Done():
			}

			return
		}

		if f.JSON200 == nil || !f.JSON200.Ok {
			logger.I("telegram get file error", log.String("reason", f.JSONDefault.Description))

			select {
			case errCh <- fmt.Errorf("failed to request file: telegram: %s", f.JSONDefault.Description):
			case <-c.Context().Done():
			}

			return
		}

		if f.JSON200.Result.FilePath == nil {
			logger.I("telegram file path not found")

			select {
			case errCh <- fmt.Errorf("Invalid empty path for telegram file downloading"):
			case <-c.Context().Done():
			}

			return
		}

		downloadURL := fmt.Sprintf(
			"https://api.telegram.org/file/bot%s/%s",
			c.botToken, *f.JSON200.Result.FilePath,
		)

		req, err := http.NewRequest(http.MethodGet, downloadURL, nil)
		if err != nil {
			logger.I("failed to create bot file download request", log.Error(err))

			select {
			case errCh <- fmt.Errorf("Internal bot error: %w", err):
			case <-c.Context().Done():
			}

			return
		}

		resp, err = c.client.Client.Do(req)
		if err != nil {
			logger.I("failed to request file download", log.Error(err))

			select {
			case errCh <- fmt.Errorf("Internal bot error: %w", err):
			case <-c.Context().Done():
			}

			return
		}
		defer func() { _ = resp.Body.Close() }()

		// TODO: add cache layer
		// 		var cacheFile *os.File
		// 		cacheFile, err = os.OpenFile("", os.O_WRONLY, 0400)
		// 		if err != nil {
		//
		// 		}

		// fileContent, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.I("failed to download file", log.Error(err))

			select {
			case errCh <- fmt.Errorf("Internal bot error: failed to read file body: %w", err):
			case <-c.Context().Done():
			}

			return
		}

		var (
			fileExt     string
			contentType string
		)

		// filename := hex.EncodeToString(sha256helper.Sum(fileContent))

		// get file extension name
		if len(requestFileName) != 0 {
			fileExt = path.Ext(requestFileName)
		}

		if len(fileExt) == 0 {
			// 			t, err2 := filetype.Match(fileContent)
			// 			if err2 == nil {
			// 				if len(t.Extension) != 0 {
			// 					fileExt = "." + t.Extension
			// 				}
			//
			// 				contentType = t.MIME.Value
			// 			}
		}

		// get content type
		if len(contentType) == 0 && len(requestFileContentType) != 0 {
			contentType = requestFileContentType
		}

		// filename += fileExt
		logger.V("uploading file",
			// log.String("upload_name", filename),
			// log.Int("size", len(fileContent)),
			log.String("content_type", contentType),
		)

		// fileURL, err := c.storage.Upload(c.Context(), filename, contentType, resp.Body)
		if err != nil {
			select {
			case errCh <- fmt.Errorf("failed to upload file: %w", err):
			case <-c.Context().Done():
			}

			return
		}

		// logger.V("file uploaded", log.String("url", fileURL))

		m.update(func() {
			// m.entities[0].Params[message.EntityParamURL] = fileURL
			m.entities[0].Params[message.EntityParamFilename] = requestFileName
			// m.entities[0].Params[message.EntityParamData] = fileContent
		})
	}()

	return errCh, nil
}
