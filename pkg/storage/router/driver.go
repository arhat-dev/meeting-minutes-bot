package router

import (
	"fmt"
	"regexp"

	"arhat.dev/mbot/internal/mime"
	"arhat.dev/mbot/pkg/rt"
	"arhat.dev/mbot/pkg/storage"
)

var _ storage.Interface = (*Driver)(nil)

type Driver struct {
	underlay []impl
}

type impl struct {
	maxSize int64
	exp     *regexp.Regexp

	store storage.Interface
}

func (m *impl) accepts(contentType string, sz int64) bool {
	if m.exp == nil {
		return m.exp.MatchString(contentType) && sz < m.maxSize
	}

	return sz < m.maxSize
}

func (m *Driver) Upload(
	con rt.Conversation, filename string, contentType mime.MIME, in *rt.Input,
) (url string, err error) {
	sz := len(m.underlay)
	if sz == 0 {
		// no driver, storage disabled
		return
	}

	for i := 0; i < sz; i++ {
		if m.underlay[i].accepts(contentType.Value, in.Size()) {
			return m.underlay[i].store.Upload(con, filename, contentType, in)
		}
	}

	return "", fmt.Errorf("not handled by any storage driver")
}