package testutil

import (
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/database"
)

type MockAction struct {
	InsertFunc        func(userId, videoId uuid.UUID) error
	GetFunc           func(userId, videoId uuid.UUID) (*database.Action, error)
	UpdateFunc        func(userId, videoId uuid.UUID, reaction string) error
	ClearReactionFunc func(userId, videoId uuid.UUID) error
}

func (m *MockAction) Insert(userId, videoId uuid.UUID) error {
	return m.InsertFunc(userId, videoId)
}
func (m *MockAction) Get(userId, videoId uuid.UUID) (*database.Action, error) {
	return m.GetFunc(userId, videoId)
}
func (m *MockAction) Update(userId, videoId uuid.UUID, reaction string) error {
	return m.UpdateFunc(userId, videoId, reaction)
}
func (m *MockAction) ClearReaction(userId, videoId uuid.UUID) error {
	return m.ClearReactionFunc(userId, videoId)
}
