package multipart

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"strings"
	"unsafe"
)

type Builder struct {
	boundary string
	parts    []Part
}

func (b *Builder) CreatePart(header Header, data io.Reader) *Builder {
	b.parts = append(b.parts, Part{
		header: string(header),
		data:   data,
	})

	return b
}

const (
	errInvalidBoundaryLength errString = "invalid boundary length"
	errInvalidBoundaryChar   errString = "invalid boundary character"
)

func (b *Builder) SetBoundary(boundary string) error {
	// rfc2046#section-5.1.1
	if len(boundary) < 1 || len(boundary) > 70 {
		return errInvalidBoundaryLength
	}
	end := len(boundary) - 1
	for i, b := range boundary {
		if 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' || '0' <= b && b <= '9' {
			continue
		}
		switch b {
		case '\'', '(', ')', '+', '_', ',', '-', '.', '/', ':', '=', '?':
			continue
		case ' ':
			if i != end {
				continue
			}
		}
		return errInvalidBoundaryChar
	}

	b.boundary = boundary
	return nil
}

func (b *Builder) Build() (formDataContentType string, r Reader) {
	const (
		PREFIX = "multipart/form-data; boundary="
	)

	boundary := b.boundary
	if len(boundary) == 0 {
		boundary = RandomBoundary()
	}

	// We must quote the boundary if it contains any of the
	// tspecials characters defined by RFC 2045, or space.
	if strings.ContainsAny(boundary, `()<>@,;:\"/[]?= `) {
		formDataContentType = PREFIX + `"` + boundary + `"`
	} else {
		formDataContentType = PREFIX + boundary
	}

	return formDataContentType, Reader{
		offset:   -2,
		boundary: boundary,
		parts:    b.parts,
	}
}

func RandomBoundary() string {
	var buf [60]byte
	_, err := io.ReadFull(rand.Reader, buf[30:])
	if err != nil {
		panic(err)
	}

	ret := buf[:]
	_ = hex.Encode(ret, buf[30:])
	return *(*string)(unsafe.Pointer(&ret))
}
