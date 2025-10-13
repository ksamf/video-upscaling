package main

import (
	"context"
	"fmt"
	"os"

	_ "github.com/lib/pq"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamf/video-upscaling/backend/internal/aws"
	"github.com/ksamf/video-upscaling/backend/internal/config"
	"github.com/ksamf/video-upscaling/backend/internal/database"
	"github.com/redis/go-redis/v9"
)

type application struct {
	host   string
	port   int
	models database.Models
	config *config.Config
	s3     *aws.S3Storage
	redis  *redis.Client
}

func main() {

	conf := config.New()
	s3 := aws.New(conf)

	url := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		conf.Postgres.User,
		conf.Postgres.Pass,
		conf.Postgres.Host,
		conf.Postgres.Port,
		conf.Postgres.Name)
	pool, err := pgxpool.New(context.Background(), url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	models := database.NewModel(pool)
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Redis.Host + ":" + fmt.Sprintf("%d", conf.Redis.Port),
		Password: conf.Redis.Pass,
		DB:       0,
	})
	_, err = client.Ping(context.Background()).Result()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to redis: %v\n", err)
		os.Exit(1)
	}
	app := &application{
		host:   conf.App.Host,
		port:   conf.App.Port,
		models: models,
		config: conf,
		s3:     s3,
		redis:  client,
	}
	if err := app.serve(); err != nil {
		panic(err)
	}

}
