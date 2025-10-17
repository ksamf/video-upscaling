package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/ksamf/video-upscaling/backend/internal/database"
	broker "github.com/ksamf/video-upscaling/backend/internal/kafka"
	"github.com/ksamf/video-upscaling/backend/internal/rest"
	"github.com/ksamf/video-upscaling/backend/internal/storage"
)

var StandardHeights = []int{144, 240, 360, 480, 720, 1080, 1440, 2160, 4320}

func processVideoJob(job broker.VideoJob, db database.VideoModel, s3 *storage.Storage) error {
	videoIDStr := job.VideoID.String()
	s3Path := fmt.Sprintf("%s/tmp%s", videoIDStr, job.FileExt)
	tmpInputPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_input.%s", videoIDStr, job.FileExt))
	defer os.Remove(tmpInputPath)
	defer s3.DeleteObject(s3Path)
	fmt.Println(s3Path)
	if err := s3.GetObject(s3Path, tmpInputPath); err != nil {
		return fmt.Errorf("failed to download from S3: %w", err)
	}

	_, height, err := GetResolution(tmpInputPath)
	if err != nil {
		return fmt.Errorf("failed get height:%w", err)
	}
	if !slices.Contains(StandardHeights, height) {
		height = ClosestStandardHeight(height)
		crf := 26 - 2*height
		TranscodeVideo(tmpInputPath, height, crf, videoIDStr, s3, 30*time.Minute)

	}
	origKey := fmt.Sprintf("%s/%d.mp4", videoIDStr, height)
	origFile, err := os.Open(tmpInputPath)
	if err != nil {
		return fmt.Errorf("failed to open original temp file: %w", err)
	}
	if err := s3.PutObject(origKey, origFile); err != nil {
		_ = origFile.Close()
		return fmt.Errorf("failed to upload original: %w", err)
	}
	_ = origFile.Close()
	log.Printf("Uploaded original as %s", origKey)

	qualities := LowerStandardRes(height)

	errCh := make(chan error, len(qualities)+5)
	var collected []error
	var mu sync.Mutex
	doneErr := make(chan struct{})

	go func() {
		for e := range errCh {
			if e == nil {
				continue
			}
			mu.Lock()
			collected = append(collected, e)
			mu.Unlock()
			log.Printf("worker error: %v", e)
		}
		close(doneErr)
	}()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := ExtractAudio(tmpInputPath, videoIDStr, s3); err != nil {
			errCh <- fmt.Errorf("audio extract failed: %w", err)
		}

		lang, err := rest.CreateSubtitles(job.VideoID, job.BaseURL)
		if err != nil {
			errCh <- fmt.Errorf("create subtitles request failed: %w", err)
			return
		}
		langId, err := db.GetLanguageId(lang)
		if langId == 0 || err != nil {
			errCh <- fmt.Errorf("db get language failed: %w", err)
		}
		if err := db.Insert(&database.Video{
			VideoId:    job.VideoID,
			Name:       job.FileName,
			LanguageId: langId,
			Quality:    height,
		}); err != nil {
			errCh <- fmt.Errorf("db insert failed: %w", err)
			return
		}
	}()

	for i, q := range qualities {
		crf := 26 - 2*i
		if crf < 8 {
			crf = 8
		}
		if crf > 26 {
			crf = 26
		}

		wg.Add(1)
		go func(targetHeight, crf int) {
			defer wg.Done()
			if err := TranscodeVideo(tmpInputPath, targetHeight, crf, videoIDStr, s3, 30*time.Minute); err != nil {
				errCh <- fmt.Errorf("transcode %dp failed: %w", targetHeight, err)
			}
		}(q, crf)
	}

	if height > 1440 {
		job.Upscale = false
	}
	if job.Upscale {

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rest.Upscale(job.VideoID, job.BaseURL, height, job.RealisticVideo); err != nil {
				errCh <- fmt.Errorf("upscale failed: %w", err)
			}
		}()
	}

	wg.Wait()
	close(errCh)
	<-doneErr

	if len(collected) > 0 {
		for _, e := range collected {
			log.Printf("processing error: %v", e)
		}
		return fmt.Errorf("some processing tasks failed")
	}

	return nil
}
