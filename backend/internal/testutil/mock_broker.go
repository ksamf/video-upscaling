package testutil

import "github.com/ksamf/video-upscaling/backend/internal/broker"

type MockBroker struct {
	WriteFunc         func(msg broker.KafkaMessage) error
	ReadProcessorFunc func() (*broker.VideoProcessor, error)
	ReadUpscaleFunc   func() (*broker.VideoUpscale, error)
	ReadSubtitleFunc  func() (*broker.VideoSubtitle, error)
}

func (m *MockBroker) Write(msg broker.KafkaMessage) error {
	return m.WriteFunc(msg)
}
func (m *MockBroker) ReadProcessor() (*broker.VideoProcessor, error) {
	return m.ReadProcessorFunc()
}
func (m *MockBroker) ReadUpscale() (*broker.VideoUpscale, error) {
	return m.ReadUpscaleFunc()
}
func (m *MockBroker) ReadSubtitle() (*broker.VideoSubtitle, error) {
	return m.ReadSubtitleFunc()
}
