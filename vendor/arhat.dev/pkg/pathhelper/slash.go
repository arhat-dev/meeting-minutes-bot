package pathhelper

import (
	"strings"

	"arhat.dev/pkg/numhelper"
	"arhat.dev/pkg/stringhelper"
)

// ToSlash replaces all backslashes to slashs in s as ret, unlike path/filepath package from stdlib,
// this function always does the replace no matter what platform it is
//
// newBuf should be used for following calls when len(s) > len(buf) and s contains backslash
func ToSlash(buf []byte, s string) (ret string, newBuf []byte) {
	var (
		offset int
		i      int
		sz     int
	)

	for {
		i = strings.IndexByte(s[offset:], '\\')
		if i == -1 {
			if offset == 0 {
				return s, buf
			}

			return stringhelper.Convert[string, byte](buf[:sz]), buf
		}

		if offset == 0 {
			// first time found, replace with buffer
			sz = len(s)
			k := cap(buf)
			if sz > k {
				// align to multiple of 256 bytes
				buf = make([]byte, numhelper.SizeAlign(uint64(sz), 256))
			} else {
				buf = buf[:k]
			}

			copy(buf, s)
		}

		offset += i
		buf[offset] = '/'
		offset++
	}
}
