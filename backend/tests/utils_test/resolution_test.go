package utils_tests

import (
	"os"
	"testing"

	"github.com/ksamf/video-upscaling/backend/internal/testutil"
	"github.com/ksamf/video-upscaling/backend/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetResolution(t *testing.T) {
	t.Parallel()

	tmpFile := "test.mp4"
	err := testutil.CreateTestMP4(tmpFile)
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpFile) }()
	tests := []struct {
		name    string
		path    string
		wantW   int
		wantH   int
		wantErr bool
	}{
		{"valid mp4", "test.mp4", 480, 360, false},
		{"nonexistent file", "none.mp4", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height, err := utils.GetResolution(tt.path)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantW, width)
				assert.Equal(t, tt.wantH, height)
			}
		})
	}
}
