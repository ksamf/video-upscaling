package interfaces

import "github.com/ksamf/video-upscaling/backend/internal/broker"

type MessageBroker interface {
	Write(msg broker.KafkaMessage) error
	ReadProcessor() (*broker.VideoProcessor, error)
	ReadUpscale() (*broker.VideoUpscale, error)
	ReadSubtitle() (*broker.VideoSubtitle, error)
}
