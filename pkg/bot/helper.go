package bot

import (
	"time"

	"arhat.dev/mbot/pkg/generator"
	"arhat.dev/mbot/pkg/rt"
)

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

func GenerateContent(
	gen generator.Interface,
	con rt.Conversation,
	cmd, params string,
	msgs []*rt.Message,
) (result rt.GeneratorOutput, err error) {
	for _, m := range msgs {
		for !m.Ready() {
			// TODO: m.Wait()
			time.Sleep(time.Second)
		}
	}

	result, err = gen.Generate(con, cmd, params, msgs)
	return
}
