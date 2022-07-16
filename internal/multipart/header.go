package multipart

import (
	"net/textproto"
	"strings"
)

type Header string

type HeaderBuilder struct {
	buf strings.Builder
}

func (h *HeaderBuilder) Reset(hdr Header) *HeaderBuilder {
	h.buf.Reset()
	h.buf.Grow(len(hdr))
	h.buf.WriteString(string(hdr))
	return h
}

func (h *HeaderBuilder) Add(name, value string) *HeaderBuilder {
	name = textproto.CanonicalMIMEHeaderKey(name)

	h.buf.Grow(len(name) + len(value) + 4)

	h.buf.WriteString(name)
	h.buf.WriteString(": ")
	h.buf.WriteString(value)
	h.buf.WriteString("\r\n")
	return h
}

func (h *HeaderBuilder) Build() Header {
	return Header(h.buf.String())
}

// TODO: implement
// func (hb HeaderBuilder) Set(name, value string) HeaderBuilder {
// 	var (
// 		start, end int
// 		szName     int
// 		data       string
// 	)
//
// 	szName = len(name)
//
// 	for {
// 		data = hb.sb.String()
// 		if start = strings.Index(data, name); start == -1 {
// 			break
// 		}
//
// 		switch {
// 		default:
// 			continue
// 		case start > 2:
// 			if data[start-2] != '\r' || data[start-1] != '\n' {
// 				continue
// 			}
//
// 			fallthrough
// 		case start == 0:
// 			if start+szName < len(data) {
// 				if data[start+szName+1] == ':' {
//
// 				}
// 				break
// 			}
// 		}
//
// 		hb.sb.Erase(start, end)
// 	}
//
// 	return hb.Add(name, value)
// }
