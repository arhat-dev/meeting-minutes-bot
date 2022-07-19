package authorized

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

/*
MIT License

Copyright (c)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/
// Copied from https://github.com/yitsushi/totp-cli/blob/main/internal/security/otp.go
// with modification to codeLength
func generateTOTPCode(token string, t time.Time, codeLength int) (string, error) {
	const (
		mask1              = 0xf
		mask2              = 0x7f
		mask3              = 0xff
		timeSplitInSeconds = 30
		shift24            = 24
		shift16            = 16
		shift8             = 8
		sumByteLength      = 8
		passwordHashLength = 32
	)

	if codeLength <= 0 {
		codeLength = 6
	}

	timer := uint64(math.Floor(float64(t.Unix()) / float64(timeSplitInSeconds)))
	// Remove spaces, some providers are giving us in a readable format
	// so they add spaces in there. If it's not removed while pasting in,
	// remove it now.
	token = strings.ReplaceAll(token, " ", "")

	// It should be uppercase always
	token = strings.ToUpper(token)

	secretBytes, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(token)
	if err != nil {
		return "", fmt.Errorf("invalid totp token: %w", err)
	}

	buf := make([]byte, sumByteLength)
	mac := hmac.New(sha1.New, secretBytes)

	binary.BigEndian.PutUint64(buf, timer)
	_, _ = mac.Write(buf)
	sum := mac.Sum(nil)

	// http://tools.ietf.org/html/rfc4226#section-5.4
	offset := sum[len(sum)-1] & mask1
	value := int64(((int(sum[offset]) & mask2) << shift24) |
		((int(sum[offset+1] & mask3)) << shift16) |
		((int(sum[offset+2] & mask3)) << shift8) |
		(int(sum[offset+3]) & mask3))

	modulo := int32(value % int64(math.Pow10(codeLength)))

	format := fmt.Sprintf("%%0%dd", codeLength)

	return fmt.Sprintf(format, modulo), nil
}
