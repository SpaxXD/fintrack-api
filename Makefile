# FinTrack API - Makefile
# Requires: go, golangci-lint, migrate CLI, swag CLI

-include .env
export

BINARY_NAME=fintrack-api
BUILD_DIR=./bin
MIGRATIONS_DIR=./migrations
DATABASE_URL ?= postgres://fintrack:fintrack@localhost:5432/fintrack?sslmode=disable

.PHONY: build run test lint migrate-up migrate-down migrate-create swagger clean help

## build: Compila o binário da aplicação
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/api

## run: Executa a aplicação localmente
run: build
	$(BUILD_DIR)/$(BINARY_NAME)

## test: Executa todos os testes
test:
	go test -v -race -count=1 ./...

## lint: Executa linter (golangci-lint)
lint:
	golangci-lint run ./...

## migrate-up: Aplica todas as migrações pendentes
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

## migrate-down: Reverte a última migração
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

## migrate-create: Cria uma nova migração (uso: make migrate-create name=nome_da_migracao)
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

## swagger: Gera documentação Swagger a partir das anotações
swagger:
	swag init -g ./cmd/api/main.go -o ./docs

## clean: Remove artefatos de build
clean:
	rm -rf $(BUILD_DIR)

## help: Mostra este menu de ajuda
help:
	@echo "Targets disponíveis:"
	@echo "  build           - Compila o binário da aplicação"
	@echo "  run             - Executa a aplicação localmente"
	@echo "  test            - Executa todos os testes"
	@echo "  lint            - Executa linter (golangci-lint)"
	@echo "  migrate-up      - Aplica todas as migrações pendentes"
	@echo "  migrate-down    - Reverte a última migração"
	@echo "  migrate-create  - Cria nova migração (make migrate-create name=xxx)"
	@echo "  swagger         - Gera documentação Swagger"
	@echo "  clean           - Remove artefatos de build"
