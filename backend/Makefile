include ../.env
export $(shell sed 's/=.*//' ../.env)

DB_URL="postgres://$(DB_USER):$(DB_PASS)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable"
.PHONY: all run create_migration migrate_up migrate_down migrate_force install_all install_air install_migrate


all:
	make migrate_up 
	air 
	
add_all:
	make install_air 
	make install_migrate

add_air:
	go install github.com/air-verse/air@latest

add_migrate:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

create_migration:
	migrate create -ext sql -dir migrations $(name)

migrate_up:
	migrate -path migrations -database "$(DB_URL)" up

migrate_down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate_force:
	migrate -path migrations -database "$(DB_URL)" force $(version)