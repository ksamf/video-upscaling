package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/aws"
	"github.com/ksamf/video-upscaling/backend/internal/database"
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

func Handler(file io.Reader, fileName uuid.UUID, db database.VideoModel) error {
	s3 := aws.New()
	var wg sync.WaitGroup
	fileNameString := fileName.String()

	tmpPath := filepath.Join(os.TempDir(), fmt.Sprintf("%s_input.mp4", fileNameString))
	out, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed create temp file: %w", err)
	}
	if _, err := io.Copy(out, file); err != nil {
		return fmt.Errorf("failed write temp file: %w", err)
	}
	out.Close()
	defer os.Remove(tmpPath)

	width, height, err := getResolutionFromFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed probe resolution: %w", err)
	}
	log.Printf("Video resolution: %dx%d", width, height)

	origKey := fmt.Sprintf("%s/%d.mp4", fileNameString, height)
	origFile, _ := os.Open(tmpPath)
	defer origFile.Close()
	if err := s3.PutObject(origKey, origFile); err != nil {
		return fmt.Errorf("failed upload original: %w", err)
	}
	log.Printf("Uploaded original as %s", origKey)

	wg.Add(1)
	go extractAudio(tmpPath, fileNameString, s3, &wg)

	q := qualities[height]
	max_q := q[len(q)-2]
	wg.Add(1)
	go func() {
		defer wg.Done()
		db.UpdateQualities(fileName, q)
	}()

	for _, vs := range videoSettings {
		if vs["height"] < max_q {
			wg.Add(1)
			go transcodeVideo(tmpPath, vs["height"], vs["crf"], fileNameString, s3, &wg)
		}
	}

	wg.Wait()
	// go http.Get(fmt.Sprintf("http://localhost:8080/api/upscale?id=%s&file=%s&real=false", fileName.String(), strconv.Itoa(height)+".mp4"))
	// go http.Get("http://localhost:8080/api/subtitles?id=" + fileName.String())
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

func transcodeVideo(inputPath string, targetHeight, crf int, fileName string, s3 *aws.S3Storage, wg *sync.WaitGroup) {
	defer wg.Done()
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
		log.Printf("transcode %dp failed: %v", targetHeight, err)
		return
	}

	outFile, _ := os.Open(tmpOut)
	defer outFile.Close()
	key := fmt.Sprintf("%s/%d.mp4", fileName, targetHeight)
	if err := s3.PutObject(key, outFile); err != nil {
		log.Printf("upload %dp failed: %v", targetHeight, err)
	}

	os.Remove(tmpOut)
}

func extractAudio(inputPath, fileName string, s3 *aws.S3Storage, wg *sync.WaitGroup) {
	defer wg.Done()
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
		log.Printf("failed to get stdout pipe: %v", err)
		return
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Printf("extractAudio start failed: %v", err)
		return
	}

	if err := s3.PutObject(fmt.Sprintf("%s/audio.mp3", fileName), stdout); err != nil {
		log.Printf("upload audio failed: %v", err)
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("extractAudio wait failed: %v", err)
	}

}
