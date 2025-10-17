package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ksamf/video-upscaling/backend/internal/storage"
)

func TranscodeVideo(inputPath string, targetHeight, crf int, fileName string, s3 *storage.Storage, timeout time.Duration) error {
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
