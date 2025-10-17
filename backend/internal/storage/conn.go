package storage

import (
	"log"

	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func New(conf *config.Config) *minio.Client {
	conn, err := minio.New(conf.S3.EndpointURL, &minio.Options{
		Creds:  credentials.NewStaticV4(conf.S3.AccessKeyID, conf.S3.SecretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}

	return conn
}
