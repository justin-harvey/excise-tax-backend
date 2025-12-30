# =====================================================
# Excise Tax Portal - Go Backend Makefile
# =====================================================

.PHONY: help build run test clean

# Configuration
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_USER ?= postgres
DB_PASSWORD ?= postgres
DB_NAME ?= excise_tax_portal
DATABASE_URL ?= postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

MIGRATIONS_DIR = ./migrations
SCRIPTS_DIR = ./scripts

# Service Binaries
BIN_DIR = ./bin
API_GATEWAY_BIN = $(BIN_DIR)/api-gateway
PAYMENT_SERVICE_BIN = $(BIN_DIR)/payment-service
TAX_SERVICE_BIN = $(BIN_DIR)/tax-service
REPORTING_SERVICE_BIN = $(BIN_DIR)/reporting-service
NOTIFICATION_SERVICE_BIN = $(BIN_DIR)/notification-service
AUTH_SERVICE_BIN = $(BIN_DIR)/auth-service

# Colors
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[1;33m
NC := \033[0m # No Color

help: ## Show this help message
	@echo "$(GREEN)Excise Tax Portal - Go Backend Makefile$(NC)"
	@echo ""
	@echo "$(YELLOW)Available targets:$(NC)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(GREEN)%-25s$(NC) %s\n", $$1, $$2}'

# =====================================================
# Build Commands
# =====================================================

build: ## Build all microservices
	@echo "$(GREEN)Building all services...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(API_GATEWAY_BIN) ./cmd/api-gateway
	@go build -o $(PAYMENT_SERVICE_BIN) ./cmd/payment-service
	@go build -o $(TAX_SERVICE_BIN) ./cmd/tax-service
	@go build -o $(REPORTING_SERVICE_BIN) ./cmd/reporting-service
	@go build -o $(NOTIFICATION_SERVICE_BIN) ./cmd/notification-service
	@go build -o $(AUTH_SERVICE_BIN) ./cmd/auth-service
	@echo "$(GREEN)All services built successfully!$(NC)"

build-api-gateway: ## Build API Gateway
	@echo "$(GREEN)Building API Gateway...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(API_GATEWAY_BIN) ./cmd/api-gateway

build-payment: ## Build Payment Service
	@echo "$(GREEN)Building Payment Service...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(PAYMENT_SERVICE_BIN) ./cmd/payment-service

build-tax: ## Build Tax Service
	@echo "$(GREEN)Building Tax Service...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(TAX_SERVICE_BIN) ./cmd/tax-service

build-reporting: ## Build Reporting Service
	@echo "$(GREEN)Building Reporting Service...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(REPORTING_SERVICE_BIN) ./cmd/reporting-service

build-notification: ## Build Notification Service
	@echo "$(GREEN)Building Notification Service...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(NOTIFICATION_SERVICE_BIN) ./cmd/notification-service

build-auth: ## Build Auth Service
	@echo "$(GREEN)Building Auth Service...$(NC)"
	@mkdir -p $(BIN_DIR)
	@go build -o $(AUTH_SERVICE_BIN) ./cmd/auth-service

# =====================================================
# Run Commands
# =====================================================

run-api-gateway: ## Run API Gateway
	@echo "$(GREEN)Starting API Gateway...$(NC)"
	@go run ./cmd/api-gateway/main.go

run-payment: ## Run Payment Service
	@echo "$(GREEN)Starting Payment Service...$(NC)"
	@go run ./cmd/payment-service/main.go

run-tax: ## Run Tax Service
	@echo "$(GREEN)Starting Tax Service...$(NC)"
	@go run ./cmd/tax-service/main.go

run-reporting: ## Run Reporting Service
	@echo "$(GREEN)Starting Reporting Service...$(NC)"
	@go run ./cmd/reporting-service/main.go

run-notification: ## Run Notification Service
	@echo "$(GREEN)Starting Notification Service...$(NC)"
	@go run ./cmd/notification-service/main.go

run-auth: ## Run Auth Service
	@echo "$(GREEN)Starting Auth Service...$(NC)"
	@go run ./cmd/auth-service/main.go

run-all: ## Run all services (requires tmux or use docker-compose instead)
	@echo "$(YELLOW)To run all services, use 'make docker-up' or run each service in separate terminals$(NC)"

# =====================================================
# Test Commands
# =====================================================

test: ## Run all tests
	@echo "$(GREEN)Running tests...$(NC)"
	@go test -v ./...

test-unit: ## Run unit tests
	@echo "$(GREEN)Running unit tests...$(NC)"
	@go test -v -short ./...

test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(NC)"
	@go test -v -run Integration ./tests/...

test-e2e: ## Run end-to-end tests
	@echo "$(GREEN)Running e2e tests...$(NC)"
	@go test -v ./tests/e2e/...

test-coverage: ## Run tests with coverage
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

# =====================================================
# Code Quality Commands
# =====================================================

lint: ## Run golangci-lint
	@echo "$(GREEN)Running linter...$(NC)"
	@golangci-lint run ./...

fmt: ## Format Go code
	@echo "$(GREEN)Formatting code...$(NC)"
	@go fmt ./...

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@go vet ./...

tidy: ## Tidy Go modules
	@echo "$(GREEN)Tidying modules...$(NC)"
	@go mod tidy
	@go mod verify

# =====================================================
# Migration Commands
# =====================================================

migrate-up: ## Apply all pending migrations
	@echo "$(GREEN)Applying migrations...$(NC)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" up

migrate-down: ## Rollback last migration
	@echo "$(YELLOW)Rolling back migration...$(NC)"
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" down 1

migrate-create: ## Create new migration (usage: make migrate-create name=add_users)
	@if [ -z "$(name)" ]; then \
		echo "$(RED)Error: name parameter required$(NC)"; \
		echo "Usage: make migrate-create name=your_migration_name"; \
		exit 1; \
	fi
	@migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

migrate-version: ## Show current migration version
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" version

migrate-force: ## Force set migration version (usage: make migrate-force version=1)
	@if [ -z "$(version)" ]; then \
		echo "$(RED)Error: version parameter required$(NC)"; \
		exit 1; \
	fi
	@migrate -path $(MIGRATIONS_DIR) -database "$(DATABASE_URL)" force $(version)

# =====================================================
# Database Commands
# =====================================================

db-create: ## Create database
	@echo "$(GREEN)Creating database...$(NC)"
	@createdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) || echo "Database may already exist"

db-drop: ## Drop database
	@echo "$(RED)Dropping database...$(NC)"
	@dropdb -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) $(DB_NAME) || echo "Database may not exist"

db-reset: db-drop db-create migrate-up ## Reset database
	@echo "$(GREEN)Database reset complete!$(NC)"

db-console: ## Open PostgreSQL console
	@psql -h $(DB_HOST) -p $(DB_PORT) -U $(DB_USER) -d $(DB_NAME)

# =====================================================
# Docker Commands
# =====================================================

docker-build: ## Build Docker images for all services
	@echo "$(GREEN)Building Docker images...$(NC)"
	@docker build -t excise-tax/api-gateway -f ./infrastructure/docker/Dockerfile.api-gateway .
	@docker build -t excise-tax/payment-service -f ./infrastructure/docker/Dockerfile.payment .
	@docker build -t excise-tax/tax-service -f ./infrastructure/docker/Dockerfile.tax .
	@docker build -t excise-tax/reporting-service -f ./infrastructure/docker/Dockerfile.reporting .
	@docker build -t excise-tax/notification-service -f ./infrastructure/docker/Dockerfile.notification .
	@docker build -t excise-tax/auth-service -f ./infrastructure/docker/Dockerfile.auth .
	@echo "$(GREEN)Docker images built!$(NC)"

docker-up: ## Start all services with docker-compose
	@echo "$(GREEN)Starting services with docker-compose...$(NC)"
	@docker-compose -f ./infrastructure/docker/docker-compose.yml up -d

docker-down: ## Stop all services with docker-compose
	@echo "$(YELLOW)Stopping services...$(NC)"
	@docker-compose -f ./infrastructure/docker/docker-compose.yml down

docker-logs: ## Show docker-compose logs
	@docker-compose -f ./infrastructure/docker/docker-compose.yml logs -f

docker-ps: ## Show running containers
	@docker-compose -f ./infrastructure/docker/docker-compose.yml ps

# =====================================================
# Development Commands
# =====================================================

dev-setup: ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@make install-tools
	@make tidy
	@make db-create
	@make migrate-up
	@echo "$(GREEN)Development environment ready!$(NC)"

install-tools: ## Install development tools
	@echo "$(GREEN)Installing development tools...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/goimports@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "$(GREEN)Tools installed!$(NC)"

clean: ## Clean build artifacts
	@echo "$(GREEN)Cleaning...$(NC)"
	@rm -rf $(BIN_DIR)
	@rm -f coverage.out coverage.html
	@go clean -cache -testcache
	@echo "$(GREEN)Clean complete!$(NC)"

# =====================================================
# Utility Commands
# =====================================================

swagger: ## Generate Swagger documentation
	@echo "$(GREEN)Generating Swagger docs...$(NC)"
	@swag init -g cmd/api-gateway/main.go -o ./docs

proto: ## Generate protobuf files (if using gRPC)
	@echo "$(GREEN)Generating protobuf files...$(NC)"
	@protoc --go_out=. --go-grpc_out=. ./proto/*.proto

deps: ## Download dependencies
	@echo "$(GREEN)Downloading dependencies...$(NC)"
	@go mod download

verify: ## Verify dependencies
	@echo "$(GREEN)Verifying dependencies...$(NC)"
	@go mod verify

# =====================================================
# CI/CD Commands
# =====================================================

ci: lint test ## Run CI checks (lint + test)
	@echo "$(GREEN)CI checks passed!$(NC)"

pre-commit: fmt lint test ## Run pre-commit checks
	@echo "$(GREEN)Pre-commit checks passed!$(NC)"

# =====================================================
# Default target
# =====================================================

.DEFAULT_GOAL := help
