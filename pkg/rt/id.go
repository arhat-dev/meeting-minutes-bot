package rt

import (
	"strconv"

	"arhat.dev/pkg/log"
	"go.uber.org/zap/zapcore"
)

type (
	ChatID    uint64
	UserID    uint64
	MessageID uint64

	CacheID uint64
)

func (id CacheID) String() string {
	return strconv.FormatUint(uint64(id), 10)
}

func LogCacheID(id CacheID) zapcore.Field { return logUint64("cache_id", id) }

func LogChatID(id ChatID) zapcore.Field       { return logUint64("chat_id", id) }
func LogSenderID(id UserID) zapcore.Field     { return logUint64("sender_id", id) }
func LogMessageID(id MessageID) zapcore.Field { return logUint64("msg_id", id) }

func LogOrigChatID(id ChatID) zapcore.Field       { return logUint64("orig_chat_id", id) }
func LogOrigSenderID(id UserID) zapcore.Field     { return logUint64("orig_sender_id", id) }
func LogOrigMessageID(id MessageID) zapcore.Field { return logUint64("orig_msg_id", id) }

func logUint64[T ~uint64](key string, v T) zapcore.Field {
	return log.Uint64(key, uint64(v))
}
