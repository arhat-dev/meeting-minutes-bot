package telegram

import (
	"fmt"
	"io"
	"sync"
	"time"

	"arhat.dev/pkg/log"
	"github.com/gotd/td/telegram/downloader"
	"github.com/gotd/td/tg"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"

	"arhat.dev/meeting-minutes-bot/pkg/bot"
	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

// errCh can be nil when there is no background pre-process worker
// nolint:gocyclo
func (c *tgBot) preprocessMessage(wf *bot.Workflow, m *Message, msg *tg.Message) (errCh chan error, err error) {
	var (
		nonText    message.Span
		doDownload func() (rt.CacheReader, string, int64, error)
	)

	media, ok := msg.GetMedia()
	if !ok {
		return c.preprocessText(wf, m, msg.GetMessage(), msg.Entities, nil)
	}

	switch t := media.(type) {
	case *tg.MessageMediaPhoto:
		photo, ok := t.GetPhoto()
		if !ok {
			return nil, nil
		}

		switch p := photo.(type) {
		case *tg.Photo:
			var (
				maxSizeName string
				maxSize     int
			)
			psizes := p.GetSizes()

			for _, psz := range psizes {
				switch s := psz.(type) {
				case *tg.PhotoSizeEmpty:
					if maxSize == 0 {
						maxSizeName = s.GetType()
					}
				case *tg.PhotoSize:
					if sz := s.GetSize(); sz > maxSize {
						maxSize = s.GetSize()
						maxSizeName = s.GetType()
					}
				case *tg.PhotoCachedSize:
					if sz := len(s.GetBytes()); sz > maxSize {
						maxSize = sz
						maxSizeName = s.GetType()
					}
				case *tg.PhotoStrippedSize:
					if sz := len(s.GetBytes()); sz > maxSize {
						maxSize = sz
						maxSizeName = s.GetType()
					}
				case *tg.PhotoSizeProgressive:
					msizes := s.GetSizes()
					for _, ms := range msizes {
						if ms > maxSize {
							maxSize = ms
							maxSizeName = s.GetType()
						}
					}
				case *tg.PhotoPathSize:
					if sz := len(s.GetBytes()); sz > maxSize {
						maxSize = sz
						maxSizeName = s.GetType()
					}
				}
			}

			fileLoc := &tg.InputPhotoFileLocation{
				ID:            p.GetID(),
				AccessHash:    p.GetAccessHash(),
				FileReference: p.GetFileReference(),
				ThumbSize:     maxSizeName,
			}

			doDownload = func() (rt.CacheReader, string, int64, error) {
				return c.download(
					c.downloader.Download(c.client.API(), fileLoc).WithVerify(true),
					maxSize,
					"",
				)
			}
		default:
			return c.preprocessText(wf, m, msg.GetMessage(), msg.Entities, nil)
		}
	case *tg.MessageMediaDocument:
		doc, ok := t.GetDocument()
		if !ok {
			return c.preprocessText(wf, m, msg.GetMessage(), msg.Entities, nil)
		}

		switch d := doc.(type) {
		case *tg.Document:
			sz := d.GetSize()
			fileLoc := d.AsInputDocumentFileLocation()
			ct := d.GetMimeType()

			attrs := d.GetAttributes()
			for _, attr := range attrs {
				switch a := attr.(type) {
				case *tg.DocumentAttributeImageSize:
				case *tg.DocumentAttributeAnimated:
				case *tg.DocumentAttributeSticker:
				case *tg.DocumentAttributeVideo:
					nonText.Duration.Set(time.Duration(a.GetDuration()) * time.Second)
				case *tg.DocumentAttributeAudio:
					nonText.Duration.Set(time.Duration(a.GetDuration()) * time.Second)
				case *tg.DocumentAttributeFilename:
					nonText.Filename.Set(a.GetFileName())
				case *tg.DocumentAttributeHasStickers:
				}
			}

			doDownload = func() (rt.CacheReader, string, int64, error) {
				return c.download(
					c.downloader.Download(c.client.API(), fileLoc).WithVerify(true),
					sz,
					ct,
				)
			}
		default:
			return c.preprocessText(wf, m, msg.GetMessage(), msg.Entities, nil)
		}

	// case *tg.MessageMediaGeo:
	// case *tg.MessageMediaContact:
	// case *tg.MessageMediaUnsupported:
	// case *tg.MessageMediaWebPage:
	// case *tg.MessageMediaVenue:
	// case *tg.MessageMediaGame:
	// case *tg.MessageMediaInvoice:
	// case *tg.MessageMediaGeoLive:
	// case *tg.MessageMediaPoll:
	// case *tg.MessageMediaDice:
	default:
		c.Logger().I("unhandled telegram message", log.Any("msg", msg))
		m.markReady()
		return nil, nil
	}

	var wg sync.WaitGroup

	if len(msg.GetMessage()) != 0 {
		errCh, err = c.preprocessText(wf, m, msg.GetMessage(), msg.Entities, &wg)
		if err != nil {
			return
		}
	}

	wg.Add(1) // add for the goroutine spawned at last

	noBgTextProcessing := errCh == nil
	if noBgTextProcessing { // no background job for text processing
		errCh = make(chan error, 1)
	} else {
		// goroutine to wait file downloading and caption handling
		go func() {
			wg.Wait()
			close(errCh)
			m.markReady()
		}()
	}

	go func() {
		defer func() {
			wg.Done()

			if noBgTextProcessing {
				close(errCh)
				m.markReady()
			}
		}()

		cacheRD, contentType, sz, err := doDownload()
		if err != nil {
			c.Logger().I("failed to download file", log.Error(err))
			c.sendErrorf(errCh, "unable to download: %w", err)
			return
		}

		storageURL, err := wf.Storage.Upload(c.Context(), cacheRD.Name(), contentType, sz, cacheRD)
		if err != nil {
			c.Logger().I("failed to upload file", log.Error(err))
			c.sendErrorf(errCh, "unable to upload file: %w", err)
			return
		}

		// seek to start to reuse this cache file
		//
		// NOTE: here we do not close the cache reader to keep it available (avoid unexpected file deletion)
		_, err = cacheRD.Seek(0, io.SeekStart)
		if err != nil {
			c.Logger().E("failed to reuse cached data", log.Error(err))
			c.sendErrorf(errCh, "bad cache reuse")
			return
		}

		nonText.URL.Set(storageURL)
		nonText.Data = cacheRD
		if len(contentType) == 0 {
			var (
				buf [32]byte
				ft  types.Type
			)

			n, _ := cacheRD.Read(buf[:])
			_, err = cacheRD.Seek(0, io.SeekStart)
			if err != nil {
				c.Logger().E("failed to seek to start", log.Error(err))
				c.sendErrorf(errCh, "bad cache reuse")
				return
			}

			ft, err = filetype.Match(buf[:n])
			if err == nil {
				contentType = ft.MIME.Value
			}
		}

		if len(contentType) != 0 {
			nonText.ContentType = contentType
		} else {
			// provide default mime type for storage driver
			nonText.ContentType = "application/octet-stream"
		}

		m.nonTextEntity.Set(nonText)
	}()

	return errCh, nil
}

func (c *tgBot) sendErrorf(errCh chan<- error, format string, args ...any) {
	select {
	case errCh <- fmt.Errorf(format, args...):
	case <-c.Context().Done():
	}
}

// preprocessText
//
// when wg is not nil, there is other worker working on the same message
func (c *tgBot) preprocessText(
	wf *bot.Workflow,
	m *Message,
	text string,
	entities []tg.MessageEntityClass,
	wg *sync.WaitGroup,
) (errCh chan error, err error) {
	mEntities := parseTextEntities(text, entities)

	if !mEntities.NeedPreProcess() {
		m.entities = mEntities

		if wg == nil {
			m.markReady()
		}

		return nil, nil
	}

	errCh = make(chan error, 1)
	if wg != nil {
		wg.Add(1)
	}

	go func() {
		defer func() {
			if wg == nil {
				close(errCh)
				m.markReady()
			} else {
				wg.Done()
			}
		}()

		err2 := bot.PreprocessText(&c.RTContext, wf, mEntities)
		if err2 != nil {
			c.sendErrorf(errCh, "Message pre-process error: %w", err2)
		}

		m.textEntities = mEntities
	}()

	return
}

func (c *tgBot) download(
	b *downloader.Builder, sz int, suggestContentType string,
) (cacheRD rt.CacheReader, contentType string, actualSize int64, err error) {
	cacheRD, actualSize, err = bot.Download(c.Cache(), func(cacheWR rt.CacheWriter) (err2 error) {
		var resp tg.StorageFileTypeClass
		switch {
		case sz < 5*1024*1024: // < 5MB
			resp, err2 = b.Stream(c.Context(), cacheWR)
		default:
			// 3 threads is optimal for downloading
			resp, err2 = b.WithThreads(3).Parallel(c.Context(), cacheWR)
		}

		if err2 != nil {
			return
		}

		switch resp.TypeID() {
		case tg.StorageFileJpegTypeID:
			contentType = "image/jpeg"
		case tg.StorageFileGifTypeID:
			contentType = "image/gif"
		case tg.StorageFilePngTypeID:
			contentType = "image/png"
		case tg.StorageFilePdfTypeID:
			contentType = "application/pdf"
		case tg.StorageFileMp3TypeID:
			contentType = "audio/mpeg"
		case tg.StorageFileMovTypeID:
			contentType = "video/quicktime"
		case tg.StorageFileMp4TypeID:
			contentType = "video/mp4"
		case tg.StorageFileWebpTypeID:
			contentType = "image/webp"
		case tg.StorageFileUnknownTypeID:
			fallthrough
		case tg.StorageFilePartialTypeID:
			fallthrough
		default:
			contentType = suggestContentType
		}

		return
	})

	return
}
