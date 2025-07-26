package main

import (
	"backend/cmd/internal/env"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Print("No .env file found")
	}
}
func main() {
	conf := env.New()

	if len(os.Args) < 2 {
		log.Fatal("Please provide a migration direction: 'up' or 'down'")
	}
	direction := os.Args[1]
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		conf.Postgres.User,
		conf.Postgres.Pass,
		conf.Postgres.Host,
		strconv.Itoa(conf.Postgres.Port),
		conf.Postgres.Name,
	)

	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v", err)
	}
	defer pool.Close()

	instance, err := postgres.WithInstance(pool, &postgres.Config{})
	if err != nil {
		log.Fatalf("Could not create migrate instance: %v", err)
	}

	fSrc, err := (&file.File{}).Open("./cmd/migrate/migrations")
	if err != nil {
		log.Fatalf("Could not open migrations folder: %v", err)
	}

	m, err := migrate.NewWithInstance("file", fSrc, "pgx", instance)
	if err != nil {
		log.Fatalf("Migration init failed: %v", err)
	}

	switch direction {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
	default:
		log.Fatal("Invalid direction. Use 'up' or 'down'")
	}
}
