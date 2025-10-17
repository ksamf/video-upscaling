package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ksamf/video-upscaling/backend/internal/config"
)

func New(conf *config.Config) *pgxpool.Pool {
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
	return pool
}
