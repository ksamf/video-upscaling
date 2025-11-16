package testutil

import (
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/database"
)

type MockVideo struct {
	InsertFunc         func(video *database.Video) error
	GetAllFunc         func(limit, offset string, getURL func(uuid.UUID) string) ([]*database.Video, error)
	GetByIDFunc        func(id uuid.UUID, getURL func(uuid.UUID) string) (*database.Video, error)
	DeleteFunc         func(id uuid.UUID) error
	GetLanguageFunc    func(id uuid.UUID) (string, error)
	UpdateLanguageFunc func(videoId uuid.UUID, lang string) error
	UpdateUpscaledFunc func(videoId uuid.UUID) error
	UpdateQualityFunc  func(videoId uuid.UUID, height int) error
	UpdateViewsFunc    func(videoId uuid.UUID, views int) error
	UpdateLikesFunc    func(videoId uuid.UUID, likes int64) error
	UpdateDislikesFunc func(videoId uuid.UUID, dislikes int64) error
}

func (m *MockVideo) Insert(video *database.Video) error {
	return m.InsertFunc(video)
}
func (m *MockVideo) GetAll(limit, offset string, getURL func(uuid.UUID) string) ([]*database.Video, error) {
	return m.GetAllFunc(limit, offset, getURL)
}
func (m *MockVideo) GetByID(id uuid.UUID, getURL func(uuid.UUID) string) (*database.Video, error) {
	return m.GetByIDFunc(id, getURL)
}
func (m *MockVideo) Delete(id uuid.UUID) error {
	return m.DeleteFunc(id)
}
func (m *MockVideo) GetLanguage(id uuid.UUID) (string, error) {
	return m.GetLanguageFunc(id)
}
func (m *MockVideo) UpdateLanguage(videoId uuid.UUID, lang string) error {
	return m.UpdateLanguageFunc(videoId, lang)
}
func (m *MockVideo) UpdateUpscaled(videoId uuid.UUID) error {
	return m.UpdateUpscaledFunc(videoId)
}
func (m *MockVideo) UpdateQuality(videoId uuid.UUID, height int) error {
	return m.UpdateQualityFunc(videoId, height)
}
func (m *MockVideo) UpdateViews(videoId uuid.UUID, views int) error {
	return m.UpdateViewsFunc(videoId, views)
}
func (m *MockVideo) UpdateLikes(videoId uuid.UUID, likes int64) error {
	return m.UpdateLikesFunc(videoId, likes)
}
func (m *MockVideo) UpdateDislikes(videoId uuid.UUID, dislikes int64) error {
	return m.UpdateDislikesFunc(videoId, dislikes)
}
