package broker

import (
	"context"
	"encoding/json"
	"time"
)

func (c *Client) ReadProcessor() (*VideoProcessor, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msg, err := c.ReaderProcessor.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}
	var job VideoProcessor
	if err := json.Unmarshal(msg.Value, &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (c *Client) ReadUpscale() (*VideoUpscale, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	msg, err := c.ReaderUpscale.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var job VideoUpscale
	if err := json.Unmarshal(msg.Value, &job); err != nil {
		return nil, err
	}
	return &job, nil
}
func (c *Client) ReadSubtitle() (*VideoSubtitle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	msg, err := c.ReaderSubtitle.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var job VideoSubtitle
	if err := json.Unmarshal(msg.Value, &job); err != nil {
		return nil, err
	}
	return &job, nil
}
