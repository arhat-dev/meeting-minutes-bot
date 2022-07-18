package rttest

import (
	"bytes"

	"arhat.dev/mbot/pkg/rt"
)

var _ rt.CacheReader = (*fakeCache)(nil)

func FakeCacheReader(data []byte) rt.CacheReader {
	ret := new(fakeCache)
	ret.buf.Reset(data)
	return ret
}

type fakeCache struct {
	buf bytes.Reader
}

// Read implements rt.CacheReader
func (f *fakeCache) Read(p []byte) (n int, err error) {
	return f.buf.Read(p)
}

// Seek implements rt.CacheReader
func (f *fakeCache) Seek(offset int64, whence int) (int64, error) {
	return f.buf.Seek(offset, whence)
}

// Close implements rt.CacheReader
func (f *fakeCache) Close() error { return nil }

// ID implements rt.CacheReader
func (f *fakeCache) ID() rt.CacheID { return 0 }

// Size implements rt.CacheReader
func (f *fakeCache) Size() (int64, error) { return f.buf.Size(), nil }
