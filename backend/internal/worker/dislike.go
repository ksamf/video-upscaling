package worker

import (
	"context"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
)

func SyncDislikesMetrics(db interfaces.Video, redis interfaces.Cache) {
	ctx := context.Background()
	iter := redis.Scan(0, "dislikes_count:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		videoID, err := uuid.Parse(strings.TrimPrefix(key, "dislikes_count:"))
		if err != nil {
			log.Println("failed parse video id")
		}
		count := redis.SCard(key)

		err = db.UpdateDislikes(videoID, count)
		if err != nil {
			log.Printf("failed update dislikes for %s: %v", videoID, err)
		}
		redis.Del(key)
	}
	if err := iter.Err(); err != nil {
		log.Println("redis scan error:", err)
	}
}
