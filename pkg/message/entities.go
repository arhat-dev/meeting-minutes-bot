package message

import (
	"context"
	"encoding/hex"
	"fmt"
	"sync"

	"arhat.dev/pkg/sha256helper"
	"github.com/h2non/filetype"
	"go.uber.org/multierr"

	"arhat.dev/meeting-minutes-bot/pkg/storage"
	"arhat.dev/meeting-minutes-bot/pkg/webarchiver"
)

func NewMessageEntities(entities []Entity, urlsToArchive map[string][]int) *Entities {
	if urlsToArchive == nil {
		urlsToArchive = make(map[string][]int)
	}

	return &Entities{
		urlsToArchive: urlsToArchive,
		entities:      entities,

		mu: &sync.Mutex{},
	}
}

type Entities struct {
	// url -> index of the related entry in `m.entities`
	urlsToArchive map[string][]int
	entities      []Entity

	mu *sync.Mutex
}

func (me *Entities) Get() []Entity {
	me.mu.Lock()
	defer me.mu.Unlock()

	return me.entities
}

func (me *Entities) Append(e Entity) {
	me.mu.Lock()
	defer me.mu.Unlock()

	me.entities = append(me.entities, e)
	if e.IsURL() && e.Params != nil {
		urlVal := e.Params[EntityParamURL]
		if url, ok := urlVal.(string); ok {
			me.urlsToArchive[url] = append(me.urlsToArchive[url], len(me.entities)-1)
		}
	}
}

func (me *Entities) NeedPreProcess() bool {
	me.mu.Lock()
	defer me.mu.Unlock()

	return len(me.urlsToArchive) != 0
}

func (me *Entities) PreProcess(ctx context.Context, w webarchiver.Interface, u storage.Interface) error {
	me.mu.Lock()
	defer me.mu.Unlock()

	// url -> archive url
	archiveURLs := make(map[string]string)
	// url -> screen shot url
	screenshotURLs := make(map[string]string)

	var err error
	for url, indexes := range me.urlsToArchive {
		if indexes == nil {
			me.urlsToArchive[url] = make([]int, 0, 1)
		}

		archiveURL, screenshot, err2 := w.Archive(ctx, url)
		if err2 != nil {
			err = multierr.Append(err, fmt.Errorf("unable to archive web page %s: %w", url, err2))
			continue
		}

		archiveURLs[url] = archiveURL

		if len(screenshot) == 0 {
			// no screenshot
			continue
		}

		var (
			contentType string
			fileExt     string
		)
		t, err2 := filetype.Match(screenshot)
		if err2 == nil {
			if len(t.Extension) != 0 {
				fileExt = "." + t.Extension
			}

			contentType = t.MIME.Value
		}

		filename := hex.EncodeToString(sha256helper.Sum(screenshot)) + fileExt
		screenshotURL, err2 := u.Upload(ctx, filename, contentType, screenshot)
		if err2 != nil {
			err = multierr.Append(err, fmt.Errorf("unable to upload web page screenshot: %w", err2))
			continue
		}

		screenshotURLs[url] = screenshotURL
	}

	for url, idxes := range me.urlsToArchive {
		for _, idx := range idxes {
			me.entities[idx].Params[EntityParamWebArchiveURL] = archiveURLs[url]
			me.entities[idx].Params[EntityParamWebArchiveScreenshotURL] = screenshotURLs[url]
		}
	}

	return nil
}
