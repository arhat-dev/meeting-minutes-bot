package telegram

import (
	"fmt"
	"io"
	"path"
	"time"

	"arhat.dev/pkg/log"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/tg"
	"github.com/h2non/filetype"
	"github.com/h2non/filetype/types"

	"arhat.dev/mbot/pkg/bot"
	"arhat.dev/mbot/pkg/rt"
)

type donwlodFunc = func() (_ rt.CacheReader, contentType, ext string, sz int64, err error)

// nolint:gocyclo
func (c *tgBot) fillMessageSpans(mc *messageContext, wf *bot.Workflow, m *rt.Message) (err error) {
	if len(m.Text) != 0 {
		m.Spans = parseTextEntities(m.Text, mc.msg.Entities)
	}

	media, ok := mc.msg.GetMedia()
	if !ok {
		return
	}

	var (
		nonText    rt.Span
		doDownload donwlodFunc
	)

	switch t := media.(type) {
	case *tg.MessageMediaPhoto:
		photo, ok := t.GetPhoto()
		if !ok {
			return nil
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
				mc.logger.D("download photo", log.Int64("size", int64(maxSize)))
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
				mc.logger.D("download file", log.Int64("size", sz), log.String("content_type", ct))
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
		return nil
	}

	m.Spans = append(m.Spans, nonText)

	if !wf.DownloadMedia() {
		return nil
	}

	mediaSpan := &m.Spans[len(m.Spans)-1]
	peer := mc.src.Chat.InputPeer()
	msgID := mc.msg.GetID()

	m.AddWorker(func(cancel rt.Signal, _ *rt.Message) {
		cacheRD, contentType, ext, sz, err := doDownload()
		if err != nil {
			mc.logger.I("failed to download file", log.Error(err))
			c.sendErrorf(peer, msgID, "unable to download: %v", err)
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
				mc.logger.E("failed to seek to start", log.Error(err))
				c.sendErrorf(peer, msgID, "bad cache reuse")
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

		mc.logger.D("upload file",
			log.String("filename", filename),
			rt.LogCacheID(cacheRD.ID()),
			log.Int64("size", sz),
		)

		input := rt.NewStorageInput(filename, sz, cacheRD, contentType)
		sout, err := wf.Storage.Upload(&mc.con, &input)
		if err != nil {
			mc.logger.I("failed to upload file", log.Error(err))
			c.sendErrorf(peer, msgID, "unable to upload file: %v", err)
			return
		}

		// seek to start to reuse this cache file
		//
		// NOTE: here we do not close the cache reader to keep it available (avoid unexpected file deletion)
		_, err = cacheRD.Seek(0, io.SeekStart)
		if err != nil {
			mc.logger.E("failed to reuse cached data", log.Error(err))
			c.sendErrorf(peer, msgID, "bad cache reuse")
			return
		}

		mediaSpan.URL = sout.URL
		mediaSpan.Data = cacheRD
	})

	return nil
}

func (c *tgBot) sendErrorf(peer tg.InputPeerClass, msgID int, format string, args ...any) {
	_, _ = c.sendTextMessage(
		c.sender.To(peer).NoWebpage().Silent().Reply(int(msgID)),
		styling.Plain("Internal bot error: "),
		styling.Bold(fmt.Sprintf(format, args...)),
	)
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
