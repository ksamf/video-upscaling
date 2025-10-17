package broker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/segmentio/kafka-go"
)

func Publish(ctx context.Context, writer *kafka.Writer, msg VideoJob) error {
	value, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	kmsg := kafka.Message{
		Key:   []byte(msg.VideoID.String()),
		Value: value,
	}
	if err := writer.WriteMessages(ctx, kmsg); err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}
	return nil
}
