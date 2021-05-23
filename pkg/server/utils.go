package server

import (
	"encoding/binary"
	"encoding/hex"

	"arhat.dev/meeting-minutes-bot/pkg/generator"
)

func generatePoweredByContent(fm generator.Formatter) string {
	return fm.Format(
		generator.KindParagraph,
		fm.Format(
			generator.KindItalic,
			"powered by",
		)+" "+fm.Format(
			generator.KindURL,
			"meeting-minutes-bot",
			"https://github.com/arhat-dev/meeting-minutes-bot",
		),
	)
}

func encodeUint64Hex(n uint64) string {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, n)
	return hex.EncodeToString(buf)
}

func decodeUint64Hex(s string) (uint64, error) {
	data, err := hex.DecodeString(s)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(data), nil
}
