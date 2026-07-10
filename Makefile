GO ?= go
SWAG ?= $(GO) run github.com/swaggo/swag/cmd/swag@v1.16.6
GOLINT ?= $(GO) run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8
AIR ?= $(GO) run github.com/air-verse/air@v1.65.3
APP_PKG ?= ./cmd/server

.PHONY: run dev build test vet lint tidy swag migrate-up migrate-down docker-up docker-down

run:
	$(GO) run $(APP_PKG)

dev:
	$(AIR) -c .air.toml

build:
	$(GO) build ./...

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

lint:
	$(GOLINT) run ./...

tidy:
	$(GO) mod tidy

swag:
	$(SWAG) init -g cmd/server/main.go -o docs --parseInternal --parseDependency --parseDepth 3

migrate-up:
	$(GO) run $(APP_PKG) -migrate up

migrate-down:
	$(GO) run $(APP_PKG) -migrate down

docker-up:
	docker compose -f deployments/docker-compose.yml up -d

docker-down:
	docker compose -f deployments/docker-compose.yml down -v
