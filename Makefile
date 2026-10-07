.PHONY: up down migrate-up migrate-down test test-unit test-integration lint fmt vet ci build run dev

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

test-unit:
	go test -race -short ./...

test-integration:
	go test -race -v ./internal/handler ./internal/ws ./internal/testutil

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

ci:
	go test -race ./...
	go vet ./...
	go build ./cmd/server

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server
