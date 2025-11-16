package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type KafkaMessage interface {
	GetVideoID() uuid.UUID
	Topic() string
}

func (c *Client) Write(msg KafkaMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	value, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	kmsg := kafka.Message{
		Topic: msg.Topic(),
		Key:   []byte(msg.GetVideoID().String()),
		Value: value,
	}
	if err := c.Writer.WriteMessages(ctx, kmsg); err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}
	return nil
}
