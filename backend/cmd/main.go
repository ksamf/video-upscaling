package main

import (
	_ "github.com/lib/pq"

	"github.com/ksamf/video-upscaling/backend/internal/auth"
	"github.com/ksamf/video-upscaling/backend/internal/broker"
	"github.com/ksamf/video-upscaling/backend/internal/cache"
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/ksamf/video-upscaling/backend/internal/handler"
	"github.com/ksamf/video-upscaling/backend/internal/interfaces"
	"github.com/ksamf/video-upscaling/backend/internal/storage"
)

type application struct {
	host     string
	port     int
	handlers *handler.Handlers
	auth     interfaces.Auth
	models   *database.Models
	config   *config.Config
	s3       interfaces.ObjectStorage
	redis    interfaces.Cache
	kafka    interfaces.MessageBroker
}

func main() {

	conf := config.New()
	s3 := storage.New(conf)
	pool := database.New(conf)
	redis := cache.New(conf)
	kafka := broker.New(conf)

	defer pool.Close()
	defer s3.CredContext().Client.CloseIdleConnections()
	models := database.NewModel(pool)
	buckets := storage.NewBucket(s3, conf)
	handlers := handler.NewHandlers(*models, conf, buckets, redis, kafka)
	auth := auth.New(&models.Users, conf)

	app := &application{
		host:     conf.App.Host,
		port:     conf.App.Port,
		handlers: handlers,
		auth:     auth,
		models:   models,
		config:   conf,
		s3:       buckets,
		redis:    redis,
		kafka:    kafka,
	}
	go app.StartWorkers()
	if err := app.serve(); err != nil {
		panic(err)
	}

}
