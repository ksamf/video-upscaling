package utils

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/ksamf/video-upscaling/backend/internal/database"
	broker "github.com/ksamf/video-upscaling/backend/internal/kafka"
	"github.com/ksamf/video-upscaling/backend/internal/storage"
	"github.com/segmentio/kafka-go"
)

func StartVideoWorker(
	reader *kafka.Reader,
	db database.VideoModel,
	s3 *storage.Storage,
) {
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("Kafka read error: %v", err)
			time.Sleep(time.Second)
			continue
		}

		var job broker.VideoJob
		if err := json.Unmarshal(msg.Value, &job); err != nil {
			log.Printf("Invalid job: %v", err)
			continue
		}

		log.Printf("Processing job %s (%s)", job.VideoID, job.FileName)

		if err := processVideoJob(job, db, s3); err != nil {
			log.Printf("Job %s failed: %v", job.VideoID, err)
		} else {
			log.Printf("Job %s completed successfully", job.VideoID)
		}
	}
}
