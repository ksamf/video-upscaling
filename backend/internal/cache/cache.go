package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisModel struct {
	conn *redis.Client
}

func (r *RedisModel) Set(key string, value any, exp time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r.conn.Set(ctx, key, value, exp)
}
func (r *RedisModel) Get(key string) string {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cache, err := r.conn.Get(ctx, key).Result()
	if err != nil {
		return ""
	}
	return cache
}
func (r *RedisModel) Incr(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r.conn.Incr(ctx, key)
}
func (r *RedisModel) Decr(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r.conn.Decr(ctx, key)
}
func (r *RedisModel) SAdd(key string, members ...any) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r.conn.SAdd(ctx, key, members)
}
func (r *RedisModel) SRem(key string, members ...any) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r.conn.SRem(ctx, key, members)
}
func (r *RedisModel) SCard(key string) int64 {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	res, err := r.conn.SCard(ctx, key).Result()
	if err != nil {
		return 0
	}
	return res
}
func (r *RedisModel) Del(key string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r.conn.Del(ctx, key)
}
func (r *RedisModel) Scan(cursor uint64, match string, count int64) *redis.ScanCmd {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	cmd := r.conn.Scan(ctx, cursor, match, count)
	return cmd
}
