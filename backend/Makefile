ifneq (,$(wildcard .env))
	include .env
	export $(shell sed -n 's/^\([^#=]*\)=.*/\1/p'.env)
endif

DB_URL=postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

.PHONY: all run create_migration migrate_up migrate_down migrate_force add_all add_air add_migrate

all:
	make migrate_up && air

run:
	air

add_all:
	make add_air
	make add_migrate

add_air:
	go install github.com/air-verse/air@latest

add_migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

create_migration:
	migrate create -ext sql -dir migrations $(name)

migrate_up:
	migrate -path migrations -database $(DB_URL) up

migrate_down:
	migrate -path migrations -database $(DB_URL) down 1

migrate_force:
	migrate -path migrations -database $(DB_URL) force $(version)
