package main

import (
	"log"

	_ "github.com/lib/pq"

	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	broker "github.com/ksamf/video-upscaling/backend/internal/kafka"
	cache "github.com/ksamf/video-upscaling/backend/internal/redis"
	"github.com/ksamf/video-upscaling/backend/internal/storage"
	"github.com/ksamf/video-upscaling/backend/internal/utils"
	"github.com/redis/go-redis/v9"
)

type application struct {
	host   string
	port   int
	models database.Models
	config *config.Config
	s3     *storage.Storage
	redis  *redis.Client
	kafka  *broker.KafkaClients
}

func main() {

	conf := config.New()
	s3 := storage.New(conf)
	pool := database.New(conf)
	redis := cache.New(conf)
	kafka := broker.New(conf)

	defer pool.Close()
	defer redis.Close()
	defer s3.CredContext().Client.CloseIdleConnections()
	defer func() {
		if kafka.Writer != nil {
			kafka.Writer.Close()
		}
		if kafka.Reader != nil {
			kafka.Reader.Close()
		}
	}()
	models := database.NewModel(pool)
	buckets := storage.NewBucket(s3, conf)

	app := &application{
		host:   conf.App.Host,
		port:   conf.App.Port,
		models: models,
		config: conf,
		s3:     buckets,
		redis:  redis,
		kafka:  kafka,
	}
	log.Println("Kafka worker started and waiting for messages...")
	go utils.StartVideoWorker(app.kafka.Reader, app.models.Videos, app.s3)
	if err := app.serve(); err != nil {
		panic(err)
	}

}
