package worker

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
	"github.com/ksamf/video-upscaling/backend/internal/utils"
)

func StartVideo(broker interfaces.MessageBroker, db interfaces.Video, s3 interfaces.ObjectStorage) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go startProcessorWorker(ctx, broker, db, s3)
	go startSubtitleWorker(ctx, broker, db)
	go startUpscaleWorker(ctx, broker, db)

	select {}
}

func startProcessorWorker(ctx context.Context, broker interfaces.MessageBroker, db interfaces.Video, s3 interfaces.ObjectStorage) {
	for {
		vp, err := broker.ReadProcessor()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			log.Printf("[ProcessorWorker]Kafka read error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		log.Printf("[ProcessorWorker] Processing job %s", vp.VideoID)

		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[ProcessorWorker] Panic recovered: %v", r)
				}
			}()

			if err := utils.ProcessVideo(vp, db, s3); err != nil {
				log.Printf("[ProcessorWorker] Job %s failed: %v", vp.VideoID, err)
			} else {
				log.Printf("[ProcessorWorker] Job %s completed successfully", vp.VideoID)
			}
		}()
	}
}

func startSubtitleWorker(ctx context.Context, broker interfaces.MessageBroker, db interfaces.Video) {
	for {
		vs, err := broker.ReadSubtitle()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			log.Printf("[SubtitleWorker]Kafka read error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		log.Printf("[SubtitleWorker] Processing job %s (%s)", vs.VideoID, vs.Language)
		if vs.Done {
			if err := db.UpdateLanguage(vs.VideoID, vs.Language); err != nil {
				log.Printf("[SubtitleWorker] Failed to update language: %v", err)
			}
		}
	}
}

func startUpscaleWorker(ctx context.Context, broker interfaces.MessageBroker, db interfaces.Video) {
	for {
		vu, err := broker.ReadUpscale()
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				continue
			}
			log.Printf("[UpscaleWorker] Kafka read error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		log.Printf("[UpscaleWorker] Processing job %s", vu.VideoID)
		if vu.Done {
			if err := db.UpdateUpscaled(vu.VideoID); err != nil {
				log.Printf("[UpscaleWorker] Failed to update upscaled flag: %v", err)
			}
		}
	}
}
