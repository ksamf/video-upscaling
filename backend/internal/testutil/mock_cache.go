package testutil

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type MockCache struct {
	SetFunc   func(key string, value any, exp time.Duration)
	GetFunc   func(key string) string
	DelFunc   func(key string)
	IncrFunc  func(key string)
	DecrFunc  func(key string)
	SAddFunc  func(key string, members ...any)
	SRemFunc  func(key string, members ...any)
	SCardFunc func(key string) int64
	ScanFunc  func(cursor uint64, match string, count int64) *redis.ScanCmd
}

func (m *MockCache) Set(key string, value any, exp time.Duration) {
	m.SetFunc(key, value, exp)
}
func (m *MockCache) Get(key string) string {
	return m.GetFunc(key)
}
func (m *MockCache) Incr(key string) {
	m.IncrFunc(key)
}
func (m *MockCache) Decr(key string) {
	m.DecrFunc(key)
}
func (m *MockCache) SAdd(key string, members ...any) {
	m.SAddFunc(key, members...)
}
func (m *MockCache) SRem(key string, members ...any) {
	m.SRemFunc(key, members...)
}
func (m *MockCache) Del(key string) {
	m.DelFunc(key)
}
func (m *MockCache) SCard(key string) int64 {
	return m.SCardFunc(key)
}
func (m *MockCache) Scan(cursor uint64, match string, count int64) *redis.ScanCmd {
	return m.Scan(cursor, match, count)
}
