.PHONY: up down migrate-up migrate-down test lint build run dev

up:
	docker compose up --build

down:
	docker compose down -v

dev:
	docker compose -f docker-compose.yml -f docker-compose.override.yml up --build

migrate-up:
	migrate -path ./migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path ./migrations -database "$(DATABASE_URL)" down 1

test:
	go test -race -cover ./...

lint:
	golangci-lint run ./...

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server
