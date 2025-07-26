package main

import (
	"backend/cmd/internal/database"
	"backend/cmd/internal/env"
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

type application struct {
	port   int
	models database.Models
}

func main() {
	conf := env.New()
	pool, err := pgxpool.New(context.Background(), fmt.Sprintf("postgres://%s:%s@%s:%s/%s", conf.Postgres.User, conf.Postgres.Pass, conf.Postgres.Host, conf.Postgres.Port, conf.Postgres.Name))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()
	models := database.NewModels(pool)
	app := &application{
		port:   8080,
		models: models,
	}

	if err := app.serve(); err != nil {
		panic(err)
	}
}
