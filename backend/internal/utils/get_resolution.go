package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

type VideoInfo struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
}

func GetResolution(path string) (int, int, error) {
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
