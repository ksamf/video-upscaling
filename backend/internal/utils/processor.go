package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/ksamf/video-upscaling/backend/internal/broker"
	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
)

var StandardHeights = []int{144, 240, 360, 480, 720, 1080, 1440, 2160, 4320}

func ProcessVideo(job *broker.VideoProcessor, db interfaces.Video, s3 interfaces.ObjectStorage) error {
	videoIDStr := job.VideoID.String()
	s3Path := fmt.Sprintf("%s/tmp%s", videoIDStr, job.FileExt)
	tmpInputPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_input.%s", videoIDStr, job.FileExt))
	defer os.Remove(tmpInputPath)
	defer s3.DeleteObject(s3Path)
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
		tmpInputPath, err = Transcode(tmpInputPath, videoIDStr, height, crf, 30*time.Minute)
		if err != nil {
			return fmt.Errorf("failed to transcode to standard height: %w", err)
		}
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
		tmpAudio, err := ExtractAudio(tmpInputPath, videoIDStr)
		if err != nil {
			errCh <- fmt.Errorf("audio extract failed: %w", err)
		}
		f, err := os.Open(tmpAudio)
		if err != nil {
			errCh <- fmt.Errorf("open extracted audio failed: %w", err)
		}
		defer f.Close()
		if err := s3.PutObject(fmt.Sprintf("%s/audio.mp3", videoIDStr), f); err != nil {
			errCh <- fmt.Errorf("s3 upload failed: %w", err)
		}
		defer os.Remove(tmpAudio)
	}()

	if err := db.UpdateQuality(job.VideoID, height); err != nil {
		errCh <- fmt.Errorf("db update quality failed: %w", err)
	}
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
			tmpOut, err := Transcode(tmpInputPath, videoIDStr, targetHeight, crf, 30*time.Minute)
			if err != nil {
				errCh <- fmt.Errorf("transcode %dp failed: %w", targetHeight, err)
				return
			}
			defer os.Remove(tmpOut)
			outFile, err := os.Open(tmpOut)
			if err != nil {
				errCh <- fmt.Errorf("failed to open transcoded file: %w", err)
			}
			defer outFile.Close()
			key := fmt.Sprintf("%s/%d.mp4", videoIDStr, targetHeight)
			if err := s3.PutObject(key, outFile); err != nil {
				errCh <- fmt.Errorf("s3 upload failed: %w", err)
			}
		}(q, crf)
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
