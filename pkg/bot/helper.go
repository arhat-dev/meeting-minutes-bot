package bot

import (
	"sync"
	"time"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
	"arhat.dev/meeting-minutes-bot/pkg/message"
	"arhat.dev/meeting-minutes-bot/pkg/rt"
)

// PreprocessText message entities, doing following jobs
// - archive web pages
func PreprocessText(ctx *rt.RTContext, wf *Workflow, entities message.Entities) error {
	for i := range entities {
		if entities[i].URL.IsNil() {
			continue
		}

		pageArchive, err := wf.WebArchiver.Archive(ctx.Context(), entities[i].URL.Get())
		if err != nil {
			continue
		}

		data, sz := pageArchive.WARC()
		if sz > 0 {
			url, err2 := wf.Storage.Upload(ctx.Context(), "", "application/warc", sz, data)
			if err2 == nil {
				entities[i].WebArchiveURL.Set(url)
			}
		}

		data, sz = pageArchive.Screenshot()
		if sz > 0 {
			url, err2 := wf.Storage.Upload(ctx.Context(), "", "image/png", sz, data)
			if err2 == nil {
				entities[i].WebArchiveScreenshotURL.Set(url)
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
		_ = cacheWR.CloseWrite()
		return
	}

	err2 := cacheWR.CloseWrite()
	if err2 != nil {
		err = err2
		return
	}

	cacheRD, err = cacheWR.OpenReader()
	if err == nil {
		sz, err = cacheRD.Size()
	}

	return
}

var msgIfacePool = &sync.Pool{
	New: func() any {
		return make([]message.Interface, 0, 16)
	},
}

func getMsgIface(sz int) (ret []message.Interface) {
	ret = msgIfacePool.Get().([]message.Interface)
	if cap(ret) < sz {
		msgIfacePool.Put(ret)
		ret = make([]message.Interface, sz)
		return
	} else {
		ret = ret[:sz]
		return
	}
}

func putMsgIface(s []message.Interface) {
	msgIfacePool.Put(s)
}

func GenerateContent[M message.Interface](gen generator.Interface, msgs []M) (result []byte, err error) {
	data := getMsgIface(len(msgs))
	for i, m := range msgs {
		for !m.Ready() {
			time.Sleep(time.Second)
		}

		data[i] = msgs[i]
	}

	result, err = gen.RenderPageBody(data)
	putMsgIface(data)
	return
}
