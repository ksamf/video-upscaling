package aws

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

var (
	sess       *session.Session
	s3svc      *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
)

func init() {
	var err error
	sess, err = session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	if err != nil {
		log.Fatalf("Ошибка создания сессии: %v", err)
	}
	s3svc = s3.New(sess)
	uploader = s3manager.NewUploader(sess)
	downloader = s3manager.NewDownloader(sess)

	log.Println("Клиент S3 готов к использованию")
}

// UploadFileToS3 загружает файл в указанный бакет.
func UploadFileToS3(bucket, key string, content []byte) error {
	upParams := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(content),
	}
	result, err := uploader.Upload(upParams)
	if err != nil {
		return fmt.Errorf("Ошибка загрузки файла в S3: %w", err)
	}
	log.Printf("Файл загружен: %s\n", result.Location)
	return nil
}

// DownloadFileFromS3 скачивает объект из S3 и сохраняет в локальный файл.
func DownloadFileFromS3(bucket, key, localPath string) error {
	file, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("Не удалось создать локальный файл: %w", err)
	}
	defer file.Close()

	dlParams := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	// Downloader записывает данные через интерфейс WriterAt,
	// но os.File реализует WriterAt, так что можно передать file.
	_, err = downloader.Download(file, dlParams)
	if err != nil {
		return fmt.Errorf("Ошибка скачивания файла из S3: %w", err)
	}
	log.Println("Файл успешно скачан в", localPath)
	return nil
}

// DeleteObjectFromS3 удаляет объект из S3-бакета.
func DeleteObjectFromS3(bucket, key string) error {
	_, err := s3svc.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("Ошибка удаления объекта из S3: %w", err)
	}
	log.Println("Объект успешно удален из S3:", key)
	return nil
}
