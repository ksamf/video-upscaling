package utils_tests

import (
	"testing"

	"github.com/ksamf/video-upscaling/backend/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestClosestStandardHeight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  int
		output int
	}{
		{"correct height", 360, 360},
		{"less number", 140, 144},
		{"negative number value", -5, 144},
		{"more value", 5000, 4320},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := utils.ClosestStandardHeight(tt.input)
			assert.Equal(t, res, tt.output)
		})
	}

}

func TestAbs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  int
		output int
	}{
		{"positive value", 5, 5},
		{"negative value", -5, 5},
		{"zero value", 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := utils.Abs(tt.input)
			assert.Equal(t, res, tt.output)
		})
	}
}

func TestLowerStandardRes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  int
		output []int
	}{
		{"360 height", 360, []int{144, 240}},
		{"144 height", 144, []int{144}},
		{"4320 height", 4320, []int{144, 240, 360, 480, 720, 1080, 1440, 2160}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := utils.LowerStandardRes(tt.input)
			assert.Equal(t, res, tt.output)
		})
	}
}
