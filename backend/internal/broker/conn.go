package broker

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/segmentio/kafka-go"
)

type Client struct {
	Writer          *kafka.Writer
	ReaderProcessor *kafka.Reader
	ReaderSubtitle  *kafka.Reader
	ReaderUpscale   *kafka.Reader
}
type VideoProcessor struct {
	VideoID uuid.UUID `json:"video_id"`
	FileExt string    `json:"file_ext"`
	Done    bool      `json:"done"`
}

func (v *VideoProcessor) GetVideoID() uuid.UUID {
	return v.VideoID
}
func (v *VideoProcessor) Topic() string {
	return "video_processor"
}

type VideoUpscale struct {
	VideoID   uuid.UUID `json:"video_id"`
	Quality   int       `json:"quality"`
	Realistic bool      `json:"realistic"`
	Done      bool      `json:"done"`
}

func (v *VideoUpscale) GetVideoID() uuid.UUID {
	return v.VideoID
}
func (v *VideoUpscale) Topic() string {
	return "video_upscale"
}

type VideoSubtitle struct {
	VideoID  uuid.UUID `json:"video_id"`
	Language string    `json:"language"`
	Done     bool      `json:"done"`
}

func (v *VideoSubtitle) GetVideoID() uuid.UUID {
	return v.VideoID
}
func (v *VideoSubtitle) Topic() string {
	return "video_subtitle"
}

func New(conf *config.Config) *Client {
	group := "video"
	broker := fmt.Sprintf("%s:%d", conf.Kafka.Host, conf.Kafka.Port)

	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		log.Fatalf("failed to connect to Kafka broker: %v", err)
	}
	defer conn.Close()
	err = conn.CreateTopics(
		kafka.TopicConfig{
			Topic:             "video_processor",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		kafka.TopicConfig{
			Topic:             "video_subtitle",
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
		kafka.TopicConfig{
			Topic:             "video_upscale",
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
	if err != nil {
		log.Fatalf("failed to create topic  %v", err)
	}
	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Balancer: &kafka.LeastBytes{},
	}

	readerProcessor := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		GroupID:  group,
		Topic:    "video_processor",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	readerSubtitle := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		GroupID:  group,
		Topic:    "video_subtitle",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	readerUpscale := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		GroupID:  group,
		Topic:    "video_upscale",
		MinBytes: 1,
		MaxBytes: 10e6,
	})
	return &Client{
		Writer:          writer,
		ReaderProcessor: readerProcessor,
		ReaderSubtitle:  readerSubtitle,
		ReaderUpscale:   readerUpscale,
	}
}
