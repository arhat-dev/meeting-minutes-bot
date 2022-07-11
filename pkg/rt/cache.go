package rt

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"hash"
	"io"
	"io/fs"
	"os"
	"strconv"
	"time"

	"arhat.dev/pkg/fshelper"
)

type Cache interface {
	// NewWriter creates a new cache writer for caching
	NewWriter() (CacheWriter, error)
}

// CacheWriter is the writer for data cache writing
type CacheWriter interface {
	io.Writer
	io.WriterAt

	// CloseWrite called when content completely written to this cache writer
	CloseWrite() error

	// Open cached data for reading
	//
	// can ONLY be called after Close() been called
	OpenReader() (CacheReader, error)
}

// CacheReader is the reader for data cache reading
type CacheReader interface {
	// Name is the sha256 sum of the content
	Name() string
	Size() (int64, error)

	io.Reader
	io.Seeker
	io.Closer
}

func NewCache(cacheDir string) (_ Cache, err error) {
	fs := fshelper.NewOSFS(false, func() (string, error) {
		return cacheDir, nil
	})

	err = fs.MkdirAll(".", 0755)
	if err != nil {
		return
	}

	return cacheImpl{
		fs: fs,
	}, nil
}

type cacheImpl struct{ fs *fshelper.OSFS }

// New creates a new cache entry for writing
func (c cacheImpl) NewWriter() (_ CacheWriter, err error) {
	var (
		f cacheFile
	)

	err = f.open(c.fs)
	if err != nil {
		return
	}

	return &f, err
}

var _ io.WriteCloser = (*cacheFile)(nil)

type cacheFile struct {
	name string
	hash hash.Hash
	file *os.File
	fs   *fshelper.OSFS
}

// Open a new cache file
func (f *cacheFile) open(fs *fshelper.OSFS) error {
	var (
		name string
		file *os.File
		n    int
	)

	for ; ; n++ {
		now := time.Now().UnixNano()
		name = strconv.FormatUint(uint64(now), 10)

		f, err := fs.OpenFile(name, os.O_CREATE|os.O_RDWR|os.O_EXCL, 0600)
		if err != nil {
			if n > 5 {
				return err
			}

			continue
		}

		file = f.(*os.File)
		break
	}

	f.name = name
	f.hash = sha256.New()
	f.file = file
	f.fs = fs

	return nil
}

// Name implements CacheReader
func (f *cacheFile) Name() string { return f.name }

// Close implements CacheReader
func (f *cacheFile) Close() (err error) {
	return f.file.Close()
}

// Seek implements CacheReader
func (f *cacheFile) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

// Read implements CacheReader
func (f *cacheFile) Read(p []byte) (int, error) {
	return f.file.Read(p)
}

func (f *cacheFile) Size() (_ int64, err error) {
	info, err := f.file.Stat()
	if err != nil {
		return
	}

	return info.Size(), nil
}

// Write implements CacheWriter
func (f *cacheFile) Write(p []byte) (n int, err error) {
	n, err = f.file.Write(p)
	if err != nil {
		return
	}

	_, err = f.hash.Write(p)
	return
}

func (f *cacheFile) WriteAt(p []byte, off int64) (n int, err error) {
	n, err = f.file.WriteAt(p, off)
	if err != nil {
		return
	}

	// TODO: write hash
	return
}

// Close implements CacheWriter
//
// MUST be called after data been written into the cache
func (f *cacheFile) CloseWrite() (err error) {
	err = f.file.Close()
	if err != nil {
		return
	}

	name := hex.EncodeToString(f.hash.Sum(nil))
	err = f.fs.Rename(f.name, name)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			return err
		}

		err = f.fs.Remove(f.name)
	}

	f.name = name

	return
}

// OpenReader implements CacheWriter
func (f *cacheFile) OpenReader() (_ CacheReader, err error) {
	file, err := f.fs.Open(f.name)
	if err != nil {
		return
	}

	f.file = file.(*os.File)
	return f, nil
}
