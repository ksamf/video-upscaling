package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/aws"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/rest"
)

type VideoInfo struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

var standardHeights = []int{144, 240, 360, 480, 720, 1080, 1440, 2160, 4320}

func VideoProcessor(file io.Reader,
	name,
	upscale,
	realisticVideo string,
	baseUrl string,
	db database.VideoModel,
	s3 *aws.S3Storage) error {

	videoID := uuid.New()
	videoIDStr := videoID.String()
	s3URL := fmt.Sprintf("https://%s/%s/%s", s3.Endpoint, s3.BucketName, videoIDStr)

	tmpInputPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_input.mp4", videoIDStr))
	out, err := os.Create(tmpInputPath)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	if _, err := io.Copy(out, file); err != nil {
		out.Close()
		os.Remove(tmpInputPath)
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := out.Close(); err != nil {
		os.Remove(tmpInputPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpInputPath) }()

	width, height, err := getResolution(tmpInputPath)
	if !slices.Contains(standardHeights, height) {
		height = closestStandardHeight(height)
		crf := 26 - 2*height
		transcodeVideo(tmpInputPath, height, crf, videoIDStr, s3, 30*time.Minute)

	}

	if err != nil {
		return fmt.Errorf("failed to probe resolution: %w", err)
	}
	log.Printf("Video resolution: %dx%d", width, height)

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

	qualities := lowerStandardRes(height)

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

		if err := extractAudio(tmpInputPath, videoIDStr, s3); err != nil {
			errCh <- fmt.Errorf("audio extract failed: %w", err)
		}

		lang, err := rest.CreateSubtitles(videoID, baseUrl)
		if err != nil {
			errCh <- fmt.Errorf("create subtitles request failed: %w", err)
			return
		}
		langId, err := db.GetLanguageId(lang)
		if langId == 0 || err != nil {
			errCh <- fmt.Errorf("db get language failed: %w", err)
		}

		if err := db.Insert(&database.Video{
			VideoId:    videoID,
			Name:       strings.TrimSuffix(name, filepath.Ext(name)),
			VideoPath:  s3URL,
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
			if err := transcodeVideo(tmpInputPath, targetHeight, crf, videoIDStr, s3, 30*time.Minute); err != nil {
				errCh <- fmt.Errorf("transcode %dp failed: %w", targetHeight, err)
			}
		}(q, crf)
	}

	boolUp, _ := strconv.ParseBool(upscale)
	if height > 1440 {
		boolUp = false
	}
	if boolUp {

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := rest.Upscale(videoID, baseUrl, height, realisticVideo); err != nil {
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
func closestStandardHeight(height int) int {
	standardHeights := []int{144, 240, 360, 480, 720, 1080, 1440, 2160, 4320}
	closest := standardHeights[0]
	minDiff := abs(height - closest)
	for _, h := range standardHeights[1:] {
		diff := abs(height - h)
		if diff < minDiff {
			minDiff = diff
			closest = h
		}
	}
	return closest
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func lowerStandardRes(base int) []int {
	var result []int
	for _, h := range standardHeights {
		if h < base {
			result = append(result, h)
		}
	}
	return result
}

func getResolution(path string) (int, int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", "-loglevel", "error", path)
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, fmt.Errorf("ffprobe failed: %w", err)
	}

	var vInfo VideoInfo
	if err := json.Unmarshal(output, &vInfo); err != nil {
		return 0, 0, fmt.Errorf("unmarshal failed: %w", err)
	}

	for _, s := range vInfo.Streams {
		if s.CodecType == "video" {
			return s.Width, s.Height, nil
		}
	}
	return 0, 0, fmt.Errorf("no video stream found")
}

func transcodeVideo(inputPath string, targetHeight, crf int, fileName string, s3 *aws.S3Storage, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	tmpOut := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d.mp4", fileName, targetHeight))
	defer func() {
		_ = os.Remove(tmpOut)
	}()

	args := []string{
		"-i", inputPath,
		"-map", "0:v:0",
		"-c:v", "libx264",
		"-crf", fmt.Sprintf("%d", crf),
		"-vf", fmt.Sprintf("scale=-2:%d", targetHeight),
		"-map", "0:a?",
		"-c:a", "aac",
		"-b:a", "128k",
		"-fflags", "+genpts",
		"-loglevel", "error",
		tmpOut,
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", args...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg transcode failed: %w", err)
	}

	outFile, err := os.Open(tmpOut)
	if err != nil {
		return fmt.Errorf("failed to open transcoded file: %w", err)
	}
	defer outFile.Close()

	key := fmt.Sprintf("%s/%d.mp4", fileName, targetHeight)
	if err := s3.PutObject(key, outFile); err != nil {
		return fmt.Errorf("s3 upload failed: %w", err)
	}

	return nil
}

func extractAudio(inputPath, fileName string, s3 *aws.S3Storage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	tmpAudio := filepath.Join(os.TempDir(), fmt.Sprintf("%s_audio.mp3", fileName))
	defer func() { _ = os.Remove(tmpAudio) }()

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-y",
		"-i", inputPath,
		"-vn",
		"-acodec", "mp3",
		"-f", "mp3",
		"-loglevel", "error",
		tmpAudio,
	)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg audio extract failed: %w", err)
	}

	f, err := os.Open(tmpAudio)
	if err != nil {
		return fmt.Errorf("open extracted audio failed: %w", err)
	}
	defer f.Close()

	if err := s3.PutObject(fmt.Sprintf("%s/audio.mp3", fileName), f); err != nil {
		return fmt.Errorf("s3 upload failed: %w", err)
	}

	return nil
}
