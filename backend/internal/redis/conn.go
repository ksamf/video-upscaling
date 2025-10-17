package cache

import (
	"context"
	"fmt"
	"os"

	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

func New(conf *config.Config) *redis.Client {
	conn := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Host + ":" + fmt.Sprintf("%d", conf.Redis.Port),
		Password: conf.Redis.Pass,
		DB:       0,
	})
	_, err := conn.Ping(context.Background()).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to redis: %v\n", err)
		os.Exit(1)
	}
	return conn
}
