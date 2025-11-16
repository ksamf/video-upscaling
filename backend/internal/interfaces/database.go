package interfaces

import (
	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/database"
)

type User interface {
	Insert(userId uuid.UUID, email, hash string) error
	GetByEmail(email string) (*database.User, error)
	GetById(id uuid.UUID) (*database.User, error)
	Delete(id uuid.UUID) error
	GetRecommendation(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error)
	GetMixedRecommendation(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error)
	Videos(userId uuid.UUID, getURL func(uuid.UUID) string) ([]*database.Video, error)
}
type Video interface {
	Insert(video *database.Video) error
	GetAll(limit, offset string, getURL func(uuid.UUID) string) ([]*database.Video, error)
	GetByID(id uuid.UUID, getURL func(uuid.UUID) string) (*database.Video, error)
	Delete(id uuid.UUID) error
	GetLanguage(id uuid.UUID) (string, error)
	UpdateLanguage(videoId uuid.UUID, lang string) error
	UpdateUpscaled(videoId uuid.UUID) error
	UpdateQuality(videoId uuid.UUID, height int) error
	UpdateViews(videoId uuid.UUID, views int) error
	UpdateLikes(videoId uuid.UUID, likes int64) error
	UpdateDislikes(videoId uuid.UUID, dislikes int64) error
}
type Action interface {
	Insert(userId, videoId uuid.UUID) error
	Get(userId, videoId uuid.UUID) (*database.Action, error)
	Update(userId, videoId uuid.UUID, reaction string) error
	ClearReaction(userId, videoId uuid.UUID) error
}
