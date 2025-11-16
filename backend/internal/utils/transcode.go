package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func Transcode(inputPath, fileName string, targetHeight, crf int, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("input file does not exist: %w", err)
	}

	tmpOut := filepath.Join(os.TempDir(), fmt.Sprintf("%s_%d.mp4", fileName, targetHeight))

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
		return "", fmt.Errorf("ffmpeg transcode failed: %w", err)
	}
	return tmpOut, nil
}
