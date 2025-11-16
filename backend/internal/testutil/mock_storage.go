package testutil

import (
	"io"

	"github.com/google/uuid"
)

type MockStorage struct {
	PutObjectFunc    func(object string, reader io.Reader) error
	GetObjectFunc    func(object, tmpPath string) error
	DeleteObjectFunc func(object string) error
	ExitsObjectsFunc func(object string) bool
	GetURLFunc       func(id uuid.UUID) string
}

func (s3 *MockStorage) PutObject(object string, reader io.Reader) error {
	return s3.PutObjectFunc(object, reader)
}
func (s3 *MockStorage) GetObject(object, tmpPath string) error {
	return s3.GetObjectFunc(object, tmpPath)
}

func (s3 *MockStorage) DeleteObject(object string) error {
	return s3.DeleteObjectFunc(object)
}
func (s3 *MockStorage) ExitsObjects(object string) bool {
	return s3.ExitsObjectsFunc(object)
}
func (s3 *MockStorage) GetURL(id uuid.UUID) string {
	return s3.GetURLFunc(id)
}
