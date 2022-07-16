package bot

import (
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
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
func PreprocessText(ctx *rt.RTContext, wf *Workflow, spans []rt.Span) error {
	var input rt.Input

	for i := range spans {
		if len(spans[i].URL) == 0 {
			continue
		}

		if wf.WebArchiver != nil {
			pageArchive, err := wf.WebArchiver.Archive(ctx.Context(), spans[i].URL)
			if err != nil {
				continue
			}

			data, sz := pageArchive.WARC()
			if sz > 0 {
				input = rt.NewInput(sz, data)
				url, err2 := wf.Storage.Upload(ctx.Context(), "", rt.NewMIME("application/warc"), &input)
				if err2 == nil {
					spans[i].WebArchiveURL = url
				}
			}

			data, sz = pageArchive.Screenshot()
			if sz > 0 {
				input = rt.NewInput(sz, data)
				url, err2 := wf.Storage.Upload(ctx.Context(), "", rt.NewMIME("image/png"), &input)
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

func GenerateContent(gen generator.Interface, msgs []*rt.Message) (result []byte, err error) {
	for _, m := range msgs {
		for !m.Ready() {
			// TODO: m.Wait()
			time.Sleep(time.Second)
		}
	}

	result, err = gen.RenderPageBody(msgs)
	return
}
