package worker

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
)

func SyncViewsMetrics(db interfaces.Video, redis interfaces.Cache) {
	ctx := context.Background()
	iter := redis.Scan(0, "views:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		videoID, err := uuid.Parse(strings.TrimPrefix(key, "views:"))
		if err != nil {
			log.Println("failed parse video id")
		}
		countStr := redis.Get(key)
		count, _ := strconv.Atoi(countStr)

		err = db.UpdateViews(videoID, count)
		if err != nil {
			log.Printf("failed update views for %s: %v", videoID, err)
		}
		redis.Del(key)
	}
	if err := iter.Err(); err != nil {
		log.Println("redis scan error:", err)
	}
}
