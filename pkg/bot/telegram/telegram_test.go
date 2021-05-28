package telegram

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
	"testing"
)

func TestDeeplink(t *testing.T) {
	userIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(userIDBytes, math.MaxUint64)
	chatIDBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(chatIDBytes, math.MaxUint64)

	raw := fmt.Sprintf(
		"create:%s:%s",
		hex.EncodeToString(userIDBytes),
		hex.EncodeToString(chatIDBytes),
	)

	encoded := base64.URLEncoding.EncodeToString([]byte(raw))

	println(raw, len(raw))
	println(encoded, len(encoded))
}
