// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathhelper

// copied from go1.17.3 filepath/path.go

// A lazybuf is a lazily constructed path buffer.
// It supports append, reading previously appended bytes,
// and retrieving the final string. It does not allocate a buffer
// to hold the output until that output diverges from s.
type lazybuf struct {
	path       string
	buf        []byte
	w          int
	volAndPath string
	volLen     int
}

func (b *lazybuf) index(i int) byte {
	if b.buf != nil {
		return b.buf[i]
	}

	return b.path[i]
}

func (b *lazybuf) append(c byte) {
	if b.buf == nil {
		if b.w < len(b.path) && b.path[b.w] == c {
			b.w++
			return
		}
		b.buf = make([]byte, len(b.path))
		copy(b.buf, b.path[:b.w])
	}
	b.buf[b.w] = c
	b.w++
}

func (b *lazybuf) string() string {
	if b.buf == nil {
		return b.volAndPath[:b.volLen+b.w]
	}
	return b.volAndPath[:b.volLen] + string(b.buf[:b.w])
}

// isUNC reports whether path is a UNC path.
func isUNC(path string) bool {
	return volumeNameLen(path) > 2
}

// volumeNameLen returns length of the leading volume name on Windows.
// It returns 0 elsewhere.
func volumeNameLen(path string) int {
	if len(path) < 2 {
		return 0
	}
	// with drive letter
	if path[1] == ':' && IsWindowsDiskDesignatorChar(path[0]) {
		return 2
	}
	// is it UNC? https://msdn.microsoft.com/en-us/library/windows/desktop/aa365247(v=vs.85).aspx
	if l := len(path); l >= 5 && IsWindowsSlash(path[0]) && IsWindowsSlash(path[1]) &&
		!IsWindowsSlash(path[2]) && path[2] != '.' {
		// first, leading `\\` and next shouldn't be `\`. its server name.
		for n := 3; n < l-1; n++ {
			// second, next '\' shouldn't be repeated.
			if IsWindowsSlash(path[n]) {
				n++
				// third, following something characters. its share name.
				if !IsWindowsSlash(path[n]) {
					if path[n] == '.' {
						break
					}
					for ; n < l; n++ {
						if IsWindowsSlash(path[n]) {
							break
						}
					}
					return n
				}
				break
			}
		}
	}
	return 0
}
