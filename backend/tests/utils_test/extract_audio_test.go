package utils_tests

import (
	"os"
	"testing"

	"github.com/ksamf/video-upscaling/backend/internal/testutil"
	"github.com/ksamf/video-upscaling/backend/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestExtractAudio(t *testing.T) {
	t.Parallel()
	tmpFile := "test.mp4"
	err := testutil.CreateTestMP4(tmpFile)
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile) }()
	tests := []struct {
		name     string
		path     string
		fileName string
		wantErr  bool
	}{
		{"valid mp4", "test.mp4", "123412341234", false},
		{"nonexistent file", "none.mp4", "341234123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpAudio, err := utils.ExtractAudio(tt.path, tt.fileName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NotEmpty(t, tmpAudio)
				_, err = os.Stat(tmpAudio)
				assert.NoError(t, err)
				os.Remove(tmpAudio)
			}
		})
	}
}
