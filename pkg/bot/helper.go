package bot

import (
	"time"

	"arhat.dev/mbot/internal/mime"
	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
)

// NeedPreProcess returns true when any of following is true:
// - there is url to be archived
func NeedPreProcess(spans []rt.Span) bool {
	for i := range spans {
		if len(spans[i].URL) != 0 {
			return true
		}
	}

	return false
}

// PreprocessText message entities, doing following jobs
// - archive web pages
func PreprocessText(con rt.Conversation, wf *Workflow, spans []rt.Span) error {
	var input rt.Input

	for i := range spans {
		if len(spans[i].URL) == 0 {
			continue
		}

		if wf.WebArchiver != nil {
			pageArchive, err := wf.WebArchiver.Archive(con, spans[i].URL)
			if err != nil {
				continue
			}

			if pageArchive.SizeWARC > 0 {
				input = rt.NewInput(pageArchive.SizeWARC, pageArchive.WARC)
				url, err2 := wf.Storage.Upload(con, "", mime.New("application/warc"), &input)
				if err2 == nil {
					spans[i].WebArchiveURL = url
				}
			}

			if pageArchive.SizeScreenshot > 0 {
				input = rt.NewInput(pageArchive.SizeScreenshot, pageArchive.Screenshot)
				url, err2 := wf.Storage.Upload(con, "", mime.New("image/png"), &input)
				if err2 == nil {
					spans[i].WebArchiveScreenshotURL = url
				}
			}
		}
	}

	return nil
}

func Download(cache rt.Cache, doDownload func(rt.CacheWriter) error) (cacheRD rt.CacheReader, sz int64, err error) {
	cacheWR, err := cache.NewWriter()
	if err != nil {
		return
	}

	err = doDownload(cacheWR)
	if err != nil {
		_ = cacheWR.Close()
		return
	}

	err = cacheWR.Close()
	if err != nil {
		return
	}

	cacheRD, err = cache.Open(cacheWR.ID())
	if err == nil {
		sz, err = cacheRD.Size()
	}

	return
}

func GenerateContent(con rt.Conversation, gen generator.Interface, msgs []*rt.Message) (result string, err error) {
	for _, m := range msgs {
		for !m.Ready() {
			// TODO: m.Wait()
			time.Sleep(time.Second)
		}
	}

	result, err = gen.RenderBody(con, msgs)
	return
}
