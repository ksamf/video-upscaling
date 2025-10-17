package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
)

type Storage struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Client          *minio.Client
}

func (s3 *Storage) PutObject(object string, reader io.Reader) error {
	contentType := "application/octet-stream"
	info, err := s3.Client.PutObject(context.Background(), s3.BucketName, object, reader, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return fmt.Errorf("failed to put object %s: %w", object, err)
	}
	log.Printf("Successfully uploaded %s of size %d\n", object, info.Size)
	return nil
}

func (s3 *Storage) GetObject(object, tmpPath string) error {
	ctx := context.Background()

	reader, err := s3.Client.GetObject(ctx, s3.BucketName, object, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to get object %s: %w", object, err)
	}
	defer reader.Close()

	localFile, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to create tmp file: %w", err)
	}
	defer localFile.Close()

	if _, err := io.Copy(localFile, reader); err != nil {
		return fmt.Errorf("failed to copy reader to local file: %w", err)
	}

	if err := localFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync tmp file: %w", err)
	}

	return nil
}

func (s3 *Storage) DeleteObject(object string) error {
	err := s3.Client.RemoveObject(context.Background(), s3.BucketName, object, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object %s: %w", object, err)
	}
	return nil
}

func (s3 *Storage) ExitsObjects(object string) bool {
	_, err := s3.Client.StatObject(context.Background(), s3.BucketName, object, minio.StatObjectOptions{})
	if err != nil {
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false
		} else {
			return false
		}
	}
	return true
}

func (s3 *Storage) GetURL(id uuid.UUID) string {
	return fmt.Sprintf("https://%s/%s/%s", s3.Endpoint, s3.BucketName, id.String())
}
