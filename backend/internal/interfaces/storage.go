package interfaces

import (
	"io"

	"github.com/google/uuid"
)

type ObjectStorage interface {
	PutObject(object string, reader io.Reader) error
	GetObject(object, tmpPath string) error
	DeleteObject(object string) error
	ExitsObjects(object string) bool
	GetURL(id uuid.UUID) string
}
