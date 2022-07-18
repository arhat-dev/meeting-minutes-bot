package telegram

import (
	"fmt"
	"io"
	"path"
	"time"

	"arhat.dev/pkg/log"
	"github.com/gotd/td/tg"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
)

// errCh can be nil when there is no background pre-process worker
// nolint:gocyclo
func (c *tgBot) preprocessMessage(mc *messageContext, wf *bot.Workflow, m *rt.Message) (errCh chan error, err error) {
	if len(m.Text) != 0 {
		m.Spans = parseTextEntities(m.Text, mc.msg.Entities)
	}

	media, ok := mc.msg.GetMedia()
	if !ok {
		return
	}

	var (
		nonText    rt.Span
		doDownload func() (_ rt.CacheReader, contentType, ext string, sz int64, err error)
	)

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

			nonText.Flags |= rt.SpanFlag_Image

			doDownload = func() (rt.CacheReader, string, string, int64, error) {
				c.Logger().D("download photo", log.Int64("size", int64(maxSize)))
				return c.download(fileLoc, int64(maxSize), "")
			}
		default:
			return
		}
	case *tg.MessageMediaDocument:
		doc, ok := t.GetDocument()
		if !ok {
			return
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
					nonText.Flags |= rt.SpanFlag_Image
				case *tg.DocumentAttributeAnimated:
					nonText.Flags |= rt.SpanFlag_Video
				case *tg.DocumentAttributeSticker:
					if !a.Mask {
						nonText.Flags |= rt.SpanFlag_Video
					}
				case *tg.DocumentAttributeVideo:
					nonText.Flags |= rt.SpanFlag_Video
					nonText.Duration = time.Duration(a.GetDuration()) * time.Second
				case *tg.DocumentAttributeAudio:
					if a.Voice {
						nonText.Flags |= rt.SpanFlag_Voice
					} else {
						nonText.Flags |= rt.SpanFlag_Audio
					}

					nonText.Duration = time.Duration(a.GetDuration()) * time.Second
				case *tg.DocumentAttributeFilename:
					nonText.Filename = a.GetFileName()
				case *tg.DocumentAttributeHasStickers:
				}
			}

			if !nonText.IsMedia() {
				nonText.Flags |= rt.SpanFlag_File
			}

			doDownload = func() (rt.CacheReader, string, string, int64, error) {
				c.Logger().D("download file", log.Int64("size", sz), log.String("content_type", ct))
				return c.download(fileLoc, sz, ct)
			}
		default:
			return
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
		c.Logger().I("unhandled telegram message", log.Any("msg", mc.msg))
		return nil, nil
	}

	m.Spans = append(m.Spans, nonText)
	mediaSpan := &m.Spans[len(m.Spans)-1]

	errCh = make(chan error, 1)
	m.AddWorker(func(cancel rt.Signal) {
		defer close(errCh)

		cacheRD, contentType, ext, sz, err := doDownload()
		if err != nil {
			c.Logger().I("failed to download file", log.Error(err))
			c.sendErrorf(errCh, "unable to download: %w", err)
			return
		}

		if len(contentType) == 0 || len(ext) == 0 {
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
				if len(contentType) == 0 {
					contentType = ft.MIME.Value
				}

				if len(ext) == 0 {
					ext = ft.Extension
				}
			}
		}

		if len(contentType) != 0 {
			mediaSpan.ContentType = contentType
		} else {
			// provide default mime type for storage driver
			mediaSpan.ContentType = "application/octet-stream"
		}

		var filename string
		if len(mediaSpan.Filename) == 0 { // no filename set
			filename = cacheRD.ID().String() + "." + ext
		} else {
			filename = mediaSpan.Filename
			if len(path.Ext(filename)) == 0 {
				filename += "." + ext
			}
		}

		c.Logger().D("upload file",
			log.String("filename", filename),
			rt.LogCacheID(cacheRD.ID()),
			log.Int64("size", sz),
		)

		input := rt.NewStorageInput(filename, sz, cacheRD, contentType)
		sout, err := wf.Storage.Upload(&mc.con, &input)
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

		mediaSpan.URL = sout.URL
		mediaSpan.Data = cacheRD
	})

	return errCh, nil
}

func (c *tgBot) sendErrorf(errCh chan<- error, format string, args ...any) {
	select {
	case errCh <- fmt.Errorf(format, args...):
	case <-c.Context().Done():
	}
}

func (c *tgBot) download(
	loc tg.InputFileLocationClass, sz int64, suggestContentType string,
) (cacheRD rt.CacheReader, contentType, ext string, actualSize int64, err error) {
	cacheRD, actualSize, err = bot.Download(c.Cache(), func(cacheWR rt.CacheWriter) (err2 error) {
		var resp tg.StorageFileTypeClass
		switch {
		case sz < 5*1024*1024: // < 5MB
			resp, err2 = c.downloader.Download(c.client.API(), loc).
				WithVerify(true).
				Stream(c.Context(), cacheWR)
		default:
			// 3 threads is optimal for multi-thread downloading
			resp, err2 = c.downloader.Download(c.client.API(), loc).
				WithThreads(3).
				WithVerify(true).
				Parallel(c.Context(), cacheWR)
		}

		if err2 != nil {
			return
		}

		contentType = suggestContentType
		switch resp.(type) {
		case *tg.StorageFileJpeg:
			contentType, ext = "image/jpeg", "jpg"
		case *tg.StorageFileGif:
			contentType, ext = "image/gif", "gif"
		case *tg.StorageFilePng:
			contentType, ext = "image/png", "png"
		case *tg.StorageFilePdf:
			contentType, ext = "application/pdf", "pdf"
		case *tg.StorageFileMp3:
			contentType, ext = "audio/mpeg", "mp3"
		case *tg.StorageFileMov:
			contentType, ext = "video/quicktime", "mov"
		case *tg.StorageFileMp4:
			contentType, ext = "video/mp4", "mp4"
		case *tg.StorageFileWebp:
			contentType, ext = "image/webp", "webp"
		case *tg.StorageFileUnknown:
		case *tg.StorageFilePartial:
		default:
		}

		return
	})

	return
}
