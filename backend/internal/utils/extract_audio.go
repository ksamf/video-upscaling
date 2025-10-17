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

func ExtractAudio(inputPath, fileName string, s3 *storage.Storage) error {
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
