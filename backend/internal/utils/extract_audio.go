package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func ExtractAudio(inputPath, fileName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	if _, err := os.Stat(inputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("input file does not exist: %w", err)
	}
	tmpAudio := filepath.Join(os.TempDir(), fmt.Sprintf("%s_audio.mp3", fileName))

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
		return "", fmt.Errorf("ffmpeg audio extract failed: %w", err)
	}
	return tmpAudio, nil
}
