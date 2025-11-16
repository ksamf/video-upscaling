package worker

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
)

func SyncLikesMetrics(db interfaces.Video, redis interfaces.Cache) {
	ctx := context.Background()
	iter := redis.Scan(0, "likes_count:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		videoID, err := uuid.Parse(strings.TrimPrefix(key, "likes_count:"))
		if err != nil {
			log.Println("failed parse video id")
			continue
		}
		count := redis.SCard(key)
		err = db.UpdateLikes(videoID, count)
		if err != nil {
			log.Printf("failed update like for %s: %v", videoID, err)
			continue
		}
		redis.Del(key)
	}
	if err := iter.Err(); err != nil {
		log.Println("redis scan error:", err)
	}
}
