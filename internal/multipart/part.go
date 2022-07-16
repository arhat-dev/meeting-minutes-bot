package multipart

import "io"

type Part struct {
	offset int // offset in header
	header string
	data   io.Reader
}

func (t *Part) Read(p []byte) (sum int, err error) {
	var n int

	sz := len(p)
	szHeader := len(t.header)
	maxOffset := szHeader + 2

	for t.offset < maxOffset {
		switch {
		default:
			n = copy(p[sum:], t.header[t.offset:])
		case t.offset == szHeader:
			n = copy(p[sum:], "\r\n")
		case t.offset == szHeader+1:
			n = copy(p[sum:], "\n")
		}

		sum += n
		t.offset += n
		if sum == sz {
			return
		}
	}

	n, err = t.data.Read(p[sum:])
	sum += n
	return
}
