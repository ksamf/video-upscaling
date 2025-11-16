package handler

import (
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
)

type Handlers struct {
	Video  *VideoHandler
	User   *UserHandler
	Action *ActionHandler
}

func NewUserHandlers(user interfaces.User, s3 interfaces.ObjectStorage) *UserHandler {
	return &UserHandler{
		user: user,
		s3:   s3,
	}
}
func NewVideoHandlers(video interfaces.Video, action interfaces.Action, conf *config.Config, s3 interfaces.ObjectStorage, redis interfaces.Cache, kafka interfaces.MessageBroker) *VideoHandler {
	return &VideoHandler{
		video:  video,
		action: action,
		config: conf,
		s3:     s3,
		redis:  redis,
		kafka:  kafka,
	}
}
func NewActionHandlers(action interfaces.Action, redis interfaces.Cache) *ActionHandler {
	return &ActionHandler{
		action: action,
		redis:  redis,
	}
}

func NewHandlers(models database.Models, conf *config.Config, s3 interfaces.ObjectStorage, redis interfaces.Cache, kafka interfaces.MessageBroker) *Handlers {
	userHandler := NewUserHandlers(&models.Users, s3)
	videoHandler := NewVideoHandlers(&models.Videos, &models.Actions, conf, s3, redis, kafka)
	actionHandler := NewActionHandlers(&models.Actions, redis)
	return &Handlers{
		videoHandler,
		userHandler,
		actionHandler,
	}
}
