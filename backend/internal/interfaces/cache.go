package interfaces

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Set(key string, value any, exp time.Duration)
	Get(key string) string
	Del(key string)
	Incr(key string)
	Decr(key string)
	SAdd(key string, members ...any)
	SRem(key string, members ...any)
	SCard(key string) int64
	Scan(cursor uint64, match string, count int64) *redis.ScanCmd
}
