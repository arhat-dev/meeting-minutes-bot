package multipart

import (
	"errors"
	"io"
)

// Reader is like a io.MultiReader but concats all parts with text boundaries
type Reader struct {
	cur    int
	offset int // initial value = -2: "--" + boundary + "\r\n"

	boundary string // size (4, 74]
	parts    []Part
}

func (b *Reader) Read(p []byte) (sum int, err error) {
	var n int

	sz := len(p)
	nParts := len(b.parts)
	szBoundary := len(b.boundary)
	maxOffset := szBoundary + 2

	for ; b.cur < nParts; b.cur++ {
		for b.offset < maxOffset {
			switch {
			default:
				n = copy(p[sum:], b.boundary[b.offset:])
			case b.offset == -2:
				n = copy(p[sum:], "--")
			case b.offset == -1:
				n = copy(p[sum:], "-")
			case b.offset == -4:
				n = copy(p[sum:], "\r\n--")
			case b.offset == -3:
				n = copy(p[sum:], "\n--")
			case b.offset == szBoundary:
				n = copy(p[sum:], "\r\n")
			case b.offset == szBoundary+1:
				n = copy(p[sum:], "\n")
			}

			sum += n
			b.offset += n
			if sum == sz {
				return
			}
		}

		n, err = b.parts[b.cur].Read(p[sum:])
		sum += n
		if err == nil {
			// no error, b.parts[b.cur] may not drained, but this read operation finished
			// just wait for next read
			return
		}

		// error happened
		if !errors.Is(err, io.EOF) {
			return
		}

		err = nil
		// b.parts[b.cur] drained
		b.offset = -4
		if sum == sz {
			return
		}
	}

	if b.cur == nParts {
		maxOffset = szBoundary + 4
		// finish up
		// \r\n--{boundary}--\r\n
		for b.offset < maxOffset {
			switch {
			default:
				n = copy(p[sum:], b.boundary[b.offset:])
			case b.offset == -4:
				n = copy(p[sum:], "\r\n--")
			case b.offset == -3:
				n = copy(p[sum:], "\n--")
			case b.offset == -2:
				n = copy(p[sum:], "--")
			case b.offset == -1:
				n = copy(p[sum:], "-")
			case b.offset == szBoundary:
				n = copy(p[sum:], "--\r\n")
			case b.offset == szBoundary+1:
				n = copy(p[sum:], "-\r\n")
			case b.offset == szBoundary+2:
				n = copy(p[sum:], "\r\n")
			case b.offset == szBoundary+3:
				n = copy(p[sum:], "\n")
			}

			sum += n
			b.offset += n
			if sum == sz {
				return
			}
		}

		b.cur++
	}

	err = io.EOF
	return
}
