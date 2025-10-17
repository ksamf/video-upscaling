package broker

import (
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/segmentio/kafka-go"
)

type KafkaClients struct {
	Writer *kafka.Writer
	Reader *kafka.Reader
}
type VideoJob struct {
	VideoID        uuid.UUID `json:"video_id"`
	FileName       string    `json:"file_name"`
	FileExt        string    `json:"file_ext"`
	BaseURL        string    `json:"base_url"`
	Upscale        bool      `json:"upscale"`
	RealisticVideo bool      `json:"realistic_video"`
}

func New(conf *config.Config) *KafkaClients {
	topic := "video-job"
	group := "video-processor"
	broker := fmt.Sprintf("%s:%d", conf.Kafka.Host, conf.Kafka.Port)

	conn, err := kafka.Dial("tcp", broker)
	if err != nil {
		log.Fatalf("failed to connect to Kafka broker: %v", err)
	}
	defer conn.Close()
	partitions, err := conn.ReadPartitions()
	if err != nil {
		log.Fatalf("failed to read partitions: %v", err)
	}
	topicExists := false
	for _, p := range partitions {
		if p.Topic == topic {
			topicExists = true
			break
		}
	}
	if !topicExists {
		err := conn.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		})
		if err != nil {
			log.Fatalf("failed to create topic %s: %v", topic, err)
		}
		log.Println("Created Kafka topic:", topic)
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{broker},
		GroupID:  group,
		Topic:    topic,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	log.Println("Kafka connected to:", broker, "topic:", topic)
	return &KafkaClients{
		Writer: writer,
		Reader: reader,
	}
}
