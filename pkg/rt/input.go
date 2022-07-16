package rt

import (
	"bytes"
	"io"

	"arhat.dev/pkg/stringhelper"
)

func NewInput(size int64, data io.Reader) Input {
	return Input{
		size: size,
		data: data,
	}
}

type Input struct {
	size int64
	data io.Reader
}

func (in *Input) Size() int64       { return in.size }
func (in *Input) Reader() io.Reader { return in.data }

func (in *Input) Bytes() (ret []byte, err error) {
	var buf bytes.Buffer
	_, err = buf.ReadFrom(in.data)
	ret = buf.Next(buf.Len())
	return
}

func (in *Input) String() (ret string, err error) {
	data, err := in.Bytes()
	ret = stringhelper.Convert[string, byte](data)
	return
}
