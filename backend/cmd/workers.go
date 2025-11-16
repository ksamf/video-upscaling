package main

import (
	"log"
	"time"

	"github.com/ksamf/video-upscaling/backend/internal/worker"
)

func (app *application) StartWorkers() {
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			worker.SyncViewsMetrics(&app.models.Videos, app.redis)
			worker.SyncLikesMetrics(&app.models.Videos, app.redis)
			worker.SyncDislikesMetrics(&app.models.Videos, app.redis)
		}
	}()
	log.Println("Kafka worker started and waiting for messages...")
	worker.StartVideo(app.kafka, &app.models.Videos, app.s3)
}
