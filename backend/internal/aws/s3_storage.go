package aws

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Storage struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Client          *minio.Client
}

func New() *S3Storage {
	conf := config.New()
	minioClient, err := minio.New(conf.S3.EndpointURL, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.S3.AccessKeyID, conf.S3.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
	storage := &S3Storage{
		Endpoint:        conf.S3.EndpointURL,
		AccessKeyID:     conf.S3.AccessKeyID,
		SecretAccessKey: conf.S3.SecretAccessKey,
		BucketName:      conf.S3.BucketName,
		Client:          minioClient,
	}

	err = minioClient.MakeBucket(context.Background(), storage.BucketName, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := minioClient.BucketExists(context.Background(), storage.BucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", storage.BucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", storage.BucketName)
	}
	return storage
}
func (s3 *S3Storage) PutObject(object string, reader io.Reader) error {

	contentType := "application/octet-stream"
	info, err := s3.Client.PutObject(context.Background(), s3.BucketName, object, reader, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return err
	}
	log.Printf("Successfully uploaded %s of size %d\n", object, info.Size)
	return nil
}

func (s3 *S3Storage) GetObject(object string) (io.ReadCloser, error) {
	obj, err := s3.Client.GetObject(
		context.Background(),
		s3.BucketName,
		object,
		minio.GetObjectOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get object %s: %w", object, err)
	}
	return obj, nil
}

func (s3 *S3Storage) DeleteObject(object string) error {
	err := s3.Client.RemoveObject(context.Background(), s3.BucketName, object, minio.RemoveObjectOptions{})
	if err != nil {
		return err
	}
	return nil
}
