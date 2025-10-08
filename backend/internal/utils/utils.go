package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/aws"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/rest"
)

var videoSettings = []map[string]int{
	{"height": 240, "crf": 26},
	{"height": 360, "crf": 24},
	{"height": 480, "crf": 22},
	{"height": 720, "crf": 20},
	{"height": 1080, "crf": 18},
	{"height": 1440, "crf": 17},
	{"height": 2160, "crf": 16},
	{"height": 4320, "crf": 14},
	{"height": 8640, "crf": 12},
	{"height": 17280, "crf": 8},
}

var qualities = map[int][]int{
	240:  {240, 360, 480},
	360:  {240, 360, 480, 720},
	480:  {240, 360, 480, 720, 1080},
	720:  {240, 360, 480, 720, 1080, 1440},
	1080: {240, 360, 480, 720, 1080, 1440, 2160},
	1440: {240, 360, 480, 720, 1080, 1440, 2160, 4320},
	2160: {240, 360, 480, 720, 1080, 1440, 2160, 4320, 8640},
	4320: {240, 360, 480, 720, 1080, 1440, 2160, 4320, 8640, 17280},
}

type VideoInfo struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

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
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	out.Close()
	defer os.Remove(tmpInputPath)

	width, height, err := getResolutionFromFile(tmpInputPath)
	if err != nil {
		return fmt.Errorf("failed to probe resolution: %w", err)
	}
	log.Printf("Video resolution: %dx%d", width, height)

	origKey := fmt.Sprintf("%s/%d.mp4", videoIDStr, height)
	origFile, err := os.Open(tmpInputPath)
	if err != nil {
		return fmt.Errorf("failed to open original temp file: %w", err)
	}
	defer origFile.Close()

	if err := s3.PutObject(origKey, origFile); err != nil {
		return fmt.Errorf("failed to upload original: %w", err)
	}
	log.Printf("Uploaded original as %s", origKey)

	errCh := make(chan error, 10)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := extractAudio(tmpInputPath, videoIDStr, s3); err != nil {
			errCh <- fmt.Errorf("audio extract failed: %w", err)
		}
		if err := rest.CreateSubtitles(videoID, baseUrl); err != nil {
			errCh <- fmt.Errorf("create subtitles request failed: %w", err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := db.Insert(&database.Video{
			VideoId:   videoID,
			Name:      strings.TrimSuffix(name, filepath.Ext(name)),
			VideoPath: s3URL,
		}); err != nil {
			errCh <- fmt.Errorf("db insert failed: %w", err)
		}
		if err := db.UpdateQualities(videoID, qualities[height]); err != nil {
			errCh <- fmt.Errorf("db update qualities failed: %w", err)
		}
	}()

	q := qualities[height]
	maxOutput := q[len(q)-2]

	for _, vs := range videoSettings {
		targetHeight := vs["height"]
		crf := vs["crf"]

		if targetHeight < maxOutput {
			wg.Add(1)
			go func(targetHeight, crf int) {
				defer wg.Done()
				if err := transcodeVideo(tmpInputPath, targetHeight, crf, videoIDStr, s3); err != nil {
					errCh <- fmt.Errorf("transcode %dp failed: %w", targetHeight, err)
				}
			}(targetHeight, crf)
		}
	}

	wg.Wait()
	boolUp, err := strconv.ParseBool(upscale)
	if err != nil {
		// fmt.Errorf("failed parse upscale: %w", err)
	}
	if boolUp {
		// wg.Add(1)
		go func() {
			// defer wg.Done()
			rest.Upscale(videoID, baseUrl, height, realisticVideo)
			// errCh <- fmt.Errorf("upscale %d failed: %w", videoID, err)
		}()
	}
	close(errCh)

	var hasErrors bool
	for err := range errCh {
		if err != nil {
			hasErrors = true
			log.Printf("processing error: %v", err)
		}
	}

	if hasErrors {
		return fmt.Errorf("some processing tasks failed")
	}

	return nil
}

func getResolutionFromFile(path string) (int, int, error) {
	cmd := exec.Command("ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", "-loglevel", "error", path)
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

func transcodeVideo(inputPath string, targetHeight, crf int, fileName string, s3 *aws.S3Storage) error {
	tmpOut := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d.mp4", fileName, targetHeight))
	cmd := exec.Command("ffmpeg",
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
	)

	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg transcode failed: %w", err)
	}

	outFile, err := os.Open(tmpOut)
	if err != nil {
		return fmt.Errorf("failed to open transcoded file: %w", err)
	}
	defer func() {
		outFile.Close()
		os.Remove(tmpOut)
	}()

	key := fmt.Sprintf("%s/%d.mp4", fileName, targetHeight)
	if err := s3.PutObject(key, outFile); err != nil {
		return fmt.Errorf("s3 upload failed: %w", err)
	}

	return nil
}

func extractAudio(inputPath, fileName string, s3 *aws.S3Storage) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputPath,
		"-vn",
		"-acodec", "mp3",
		"-f", "mp3",
		"pipe:1",
		"-loglevel", "error",
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("stdout pipe failed: %w", err)
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start failed: %w", err)
	}

	if err := s3.PutObject(fmt.Sprintf("%s/audio.mp3", fileName), stdout); err != nil {
		return fmt.Errorf("s3 upload failed: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("ffmpeg wait failed: %w", err)
	}

	return nil
}
