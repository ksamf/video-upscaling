package utils_tests

import (
	"os"
	"testing"
	"time"

	"github.com/ksamf/video-upscaling/backend/internal/testutil"
	"github.com/ksamf/video-upscaling/backend/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestTranscode(t *testing.T) {
	t.Parallel()

	tmpFile := "test.mp4"
	err := testutil.CreateTestMP4(tmpFile)
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile) }()
	tests := []struct {
		name         string
		path         string
		fileName     string
		targetHeight int
		crf          int
		timeout      time.Duration
		wantErr      bool
	}{
		{"valid mp4", "test.mp4", "123412341234", 240, 24, time.Hour, false},
		{"nonexistent file", "none.mp4", "341234123", 240, 24, time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := utils.Transcode(tt.path, tt.fileName, tt.targetHeight, tt.crf, tt.timeout)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NotEmpty(t, tmp)
				_, err = os.Stat(tmp)
				assert.NoError(t, err)
				os.Remove(tmp)
			}
		})
	}
}
