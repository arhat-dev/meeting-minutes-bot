package rt

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"runtime"
	"time"

	"arhat.dev/pkg/fshelper"
)

type Cache interface {
	// NewWriter creates a new cache writer for caching
	NewWriter() (CacheWriter, error)

	// Open cached data by id
	Open(CacheID) (CacheReader, error)
}

// CacheWriter is the writer for data cache writing
type CacheWriter interface {
	ID() CacheID

	io.Writer
	io.WriterAt
	io.Closer
}

// CacheReader is the reader for data cache reading
type CacheReader interface {
	// Name is the sha256 sum of the content
	ID() CacheID
	Size() (int64, error)

	io.Reader
	io.Seeker
	io.Closer
}

func NewCache(cacheDir string) (_ Cache, err error) {
	osfs := fshelper.NewOSFS(false, func(fshelper.Op, string) (string, error) {
		return cacheDir, nil
	})

	err = osfs.MkdirAll(".", 0755)
	if err != nil && !errors.Is(err, fs.ErrExist) {
		return
	}

	return cacheImpl{
		fs: osfs,
	}, nil
}

type cacheImpl struct{ fs *fshelper.OSFS }

// New creates a new cache entry for writing
func (c cacheImpl) NewWriter() (_ CacheWriter, err error) {
	const MAX_TRIES = 5
	var (
		name string
		file *os.File
		now  uint64
		n    int
	)

	for {
		now = uint64(time.Now().UnixNano())
		name = CacheID(now).String()

		f, err := c.fs.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0600)
		if err != nil {
			if n > MAX_TRIES {
				return nil, err
			}

			runtime.Gosched()
			n++
			continue
		}

		file = f.(*os.File)
		break
	}

	return &cacheFile{
		id:   CacheID(now),
		File: *file,
	}, nil
}

func (c cacheImpl) Open(id CacheID) (_ CacheReader, err error) {
	file, err := c.fs.Open(id.String())
	if err != nil {
		return nil, err
	}

	return &cacheFile{
		id:   id,
		File: *file.(*os.File),
	}, nil
}

var _ io.WriteCloser = (*cacheFile)(nil)

type cacheFile struct {
	id CacheID
	os.File
}

// ID implements CacheReader.ID
func (f *cacheFile) ID() CacheID { return f.id }

// Size implements CacheReader.Size
func (f *cacheFile) Size() (_ int64, err error) {
	info, err := f.File.Stat()
	if err != nil {
		return
	}

	return info.Size(), nil
}
