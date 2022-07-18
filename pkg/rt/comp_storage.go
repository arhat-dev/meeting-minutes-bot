package rt

import (
	"bytes"
	"io"

	"arhat.dev/mbot/internal/mime"
	"arhat.dev/pkg/stringhelper"
)

// StorageOutput
type StorageOutput struct {
	// URL to download the uploaded data
	URL string
}

func NewStorageInput(filename string, size int64, data CacheReader, contentType string) StorageInput {
	return StorageInput{
		filename:    filename,
		size:        size,
		contentType: mime.New(contentType),

		data: data,
	}
}

type StorageInput struct {
	filename    string
	size        int64
	contentType mime.MIME

	data CacheReader
}

func (in *StorageInput) Filename() string { return in.filename }

// Type returns the major mime type value
func (in *StorageInput) Type() string        { return in.contentType.Type() }
func (in *StorageInput) Subtype() string     { return in.contentType.Subtype() }
func (in *StorageInput) ContentType() string { return in.contentType.Value }

func (in *StorageInput) Size() int64       { return in.size }
func (in *StorageInput) Reader() io.Reader { return in.data }

func (in *StorageInput) Bytes() (ret []byte, err error) {
	var buf bytes.Buffer
	_, err = buf.ReadFrom(in.data)
	ret = buf.Next(buf.Len())
	return
}

func (in *StorageInput) String() (ret string, err error) {
	data, err := in.Bytes()
	ret = stringhelper.Convert[string, byte](data)
	return
}
