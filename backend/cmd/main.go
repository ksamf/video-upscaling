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
)

type application struct {
	port   int
	models database.Models
	config *config.Config
	s3     *aws.S3Storage
}

func main() {

	conf := config.New()
	s3 := aws.New()
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
	app := &application{
		port:   8000,
		models: models,
		config: conf,
		s3:     s3,
	}
	if err := app.serve(); err != nil {
		panic(err)
	}

}
