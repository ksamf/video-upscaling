package testutil

import (
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/database"
)

type MockUser struct {
	InsertFunc                 func(userId uuid.UUID, email, password string) error
	GetByEmailFunc             func(email string) (*database.User, error)
	GetByIdFunc                func(id uuid.UUID) (*database.User, error)
	DeleteFunc                 func(id uuid.UUID) error
	GetRecommendationFunc      func(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error)
	GetMixedRecommendationFunc func(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error)
	VideosFunc                 func(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error)
}

func (m *MockUser) Insert(userId uuid.UUID, email, password string) error {
	return m.InsertFunc(userId, email, password)
}
func (m *MockUser) GetByEmail(email string) (*database.User, error) {
	return m.GetByEmailFunc(email)
}
func (m *MockUser) GetById(id uuid.UUID) (*database.User, error) {
	return m.GetByIdFunc(id)
}
func (m *MockUser) Delete(id uuid.UUID) error {
	return m.DeleteFunc(id)
}
func (m *MockUser) GetRecommendation(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error) {
	return m.GetRecommendationFunc(userId, getURL)
}
func (m *MockUser) GetMixedRecommendation(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error) {
	return m.GetMixedRecommendationFunc(userId, getURL)
}
func (m *MockUser) Videos(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error) {
	return m.VideosFunc(userId, getURL)
}
