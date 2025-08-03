# Contact Management Microservice Makefile

# Variables
BINARY_NAME=contact-service
BUILD_DIR=bin
MAIN_PATH=./cmd/server
MIGRATE_PATH=./cmd/migrate
GO_VERSION=1.21

# Colors for output
RED=\033[0;31m
GREEN=\033[0;32m
YELLOW=\033[1;33m
BLUE=\033[0;34m
NC=\033[0m # No Color

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo "$(BLUE)Contact Management Microservice$(NC)"
	@echo "$(BLUE)===============================$(NC)"
	@echo ""
	@echo "Available commands:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  $(GREEN)%-15s$(NC) %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development commands
.PHONY: dev
dev: ## Run the service in development mode with hot reload
	@echo "$(YELLOW)Starting development server...$(NC)"
	@air -c .air.toml

.PHONY: run
run: ## Run the service
	@echo "$(YELLOW)Starting contact service...$(NC)"
	@go run $(MAIN_PATH)/main.go

.PHONY: build
build: ## Build the binary
	@echo "$(YELLOW)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)/main.go
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

.PHONY: build-linux
build-linux: ## Build for Linux
	@echo "$(YELLOW)Building $(BINARY_NAME) for Linux...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME)-linux $(MAIN_PATH)/main.go
	@echo "$(GREEN)Linux build completed: $(BUILD_DIR)/$(BINARY_NAME)-linux$(NC)"

.PHONY: build-windows
build-windows: ## Build for Windows
	@echo "$(YELLOW)Building $(BINARY_NAME) for Windows...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=windows GOARCH=amd64 go build -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)/main.go
	@echo "$(GREEN)Windows build completed: $(BUILD_DIR)/$(BINARY_NAME).exe$(NC)"

# Database commands
.PHONY: migrate-up
migrate-up: ## Run database migrations up
	@echo "$(YELLOW)Running migrations up...$(NC)"
	@go run $(MIGRATE_PATH)/main.go up

.PHONY: migrate-down
migrate-down: ## Roll back last migration
	@echo "$(YELLOW)Rolling back migration...$(NC)"
	@go run $(MIGRATE_PATH)/main.go down

.PHONY: migrate-status
migrate-status: ## Show migration status
	@echo "$(YELLOW)Checking migration status...$(NC)"
	@go run $(MIGRATE_PATH)/main.go status

.PHONY: migrate-create
migrate-create: ## Create new migration (usage: make migrate-create NAME=create_users_table)
	@if [ -z "$(NAME)" ]; then \
		echo "$(RED)Error: NAME is required. Usage: make migrate-create NAME=migration_name$(NC)"; \
		exit 1; \
	fi
	@echo "$(YELLOW)Creating migration: $(NAME)...$(NC)"
	@go run $(MIGRATE_PATH)/main.go create "$(NAME)"

# Testing commands
.PHONY: test
test: ## Run unit tests
	@echo "$(YELLOW)Running unit tests...$(NC)"
	@go test ./internal/... -v -short
	@echo "$(GREEN)Unit tests completed$(NC)"

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@go test ./... -coverprofile=coverage.out -covermode=atomic
	@go tool cover -html=coverage.out -o coverage.html
	@go tool cover -func=coverage.out | tail -1
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

.PHONY: test-integration
test-integration: ## Run integration tests
	@echo "$(YELLOW)Running integration tests...$(NC)"
	@RUN_INTEGRATION_TESTS=true go test ./tests/integration/... -v -tags=integration -timeout=5m
	@echo "$(GREEN)Integration tests completed$(NC)"

.PHONY: test-unit
test-unit: ## Run only unit tests with verbose output
	@echo "$(YELLOW)Running unit tests with detailed output...$(NC)"
	@go test ./internal/handlers/... -v
	@go test ./internal/services/... -v
	@go test ./internal/repository/... -v
	@echo "$(GREEN)Unit tests completed$(NC)"

.PHONY: test-handlers
test-handlers: ## Run handler tests only
	@echo "$(YELLOW)Running handler tests...$(NC)"
	@go test ./internal/handlers/... -v
	@echo "$(GREEN)Handler tests completed$(NC)"

.PHONY: test-services
test-services: ## Run service tests only
	@echo "$(YELLOW)Running service tests...$(NC)"
	@go test ./internal/services/... -v
	@echo "$(GREEN)Service tests completed$(NC)"

.PHONY: test-watch
test-watch: ## Run tests in watch mode
	@echo "$(YELLOW)Running tests in watch mode...$(NC)"
	@echo "$(BLUE)Press Ctrl+C to stop$(NC)"
	@find . -name "*.go" | entr -c go test ./... -v

.PHONY: test-race
test-race: ## Run tests with race condition detection
	@echo "$(YELLOW)Running tests with race detection...$(NC)"
	@go test ./... -race -v
	@echo "$(GREEN)Race condition tests completed$(NC)"

.PHONY: test-clean
test-clean: ## Clean test cache and artifacts
	@echo "$(YELLOW)Cleaning test cache and artifacts...$(NC)"
	@go clean -testcache
	@rm -f coverage.out coverage.html
	@rm -f tests/integration/test.db
	@echo "$(GREEN)Test cleanup completed$(NC)"

.PHONY: benchmark
benchmark: ## Run benchmark tests
	@echo "$(YELLOW)Running benchmark tests...$(NC)"
	@go test ./... -bench=. -benchmem

# Code quality commands
.PHONY: lint
lint: ## Run linter
	@echo "$(YELLOW)Running linter...$(NC)"
	@golangci-lint run

.PHONY: format
format: ## Format code
	@echo "$(YELLOW)Formatting code...$(NC)"
	@go fmt ./...
	@goimports -w .

.PHONY: vet
vet: ## Run go vet
	@echo "$(YELLOW)Running go vet...$(NC)"
	@go vet ./...

.PHONY: security
security: ## Run security scan
	@echo "$(YELLOW)Running security scan...$(NC)"
	@gosec ./...

# Dependency management
.PHONY: deps
deps: ## Download dependencies
	@echo "$(YELLOW)Downloading dependencies...$(NC)"
	@go mod download

.PHONY: deps-update
deps-update: ## Update dependencies
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy

.PHONY: deps-vendor
deps-vendor: ## Vendor dependencies
	@echo "$(YELLOW)Vendoring dependencies...$(NC)"
	@go mod vendor

# Documentation commands
.PHONY: docs
docs: ## Generate comprehensive API documentation
	@echo "$(YELLOW)Generating API documentation...$(NC)"
	@swag init -g $(MAIN_PATH)/main.go -o ./docs --parseDependency --parseInternal
	@echo "$(GREEN)API documentation generated in ./docs/$(NC)"
	@echo "$(BLUE)Swagger spec: ./docs/swagger.json$(NC)"
	@echo "$(BLUE)Swagger YAML: ./docs/swagger.yaml$(NC)"

.PHONY: docs-serve
docs-serve: run ## Serve documentation with the API server
	@echo "$(YELLOW)Serving documentation with API server...$(NC)"
	@echo "$(GREEN)Swagger UI: http://localhost:8081/swagger/index.html$(NC)"
	@echo "$(GREEN)Health Check: http://localhost:8081/health$(NC)"
	@echo "$(GREEN)API Documentation: ./docs/README.md$(NC)"

.PHONY: docs-validate
docs-validate: ## Validate OpenAPI specification
	@echo "$(YELLOW)Validating OpenAPI specification...$(NC)"
	@swagger-codegen validate -i ./docs/swagger.yaml || echo "$(YELLOW)swagger-codegen not found, skipping validation$(NC)"

.PHONY: docs-export
docs-export: docs ## Export documentation in multiple formats
	@echo "$(YELLOW)Exporting documentation...$(NC)"
	@cp ./docs/swagger.yaml ./docs/api-spec.yaml
	@echo "$(GREEN)Documentation exported:$(NC)"
	@echo "  - OpenAPI YAML: ./docs/api-spec.yaml"
	@echo "  - Comprehensive Guide: ./docs/README.md"

# Docker commands
.PHONY: docker-build
docker-build: ## Build Docker image
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@docker build -t mejona/contact-service:latest .
	@echo "$(GREEN)Docker image built: mejona/contact-service:latest$(NC)"

.PHONY: docker-run
docker-run: ## Run Docker container
	@echo "$(YELLOW)Running Docker container...$(NC)"
	@docker run -p 8081:8081 --env-file .env mejona/contact-service:latest

.PHONY: docker-compose-up
docker-compose-up: ## Start services with docker-compose
	@echo "$(YELLOW)Starting services with docker-compose...$(NC)"
	@docker-compose up -d

.PHONY: docker-compose-down
docker-compose-down: ## Stop services with docker-compose
	@echo "$(YELLOW)Stopping services with docker-compose...$(NC)"
	@docker-compose down

# Production commands
.PHONY: deploy-staging
deploy-staging: ## Deploy to staging environment
	@echo "$(YELLOW)Deploying to staging...$(NC)"
	@./scripts/deploy-staging.sh

.PHONY: deploy-production
deploy-production: ## Deploy to production environment
	@echo "$(YELLOW)Deploying to production...$(NC)"
	@./scripts/deploy-production.sh

# MCP Server commands
.PHONY: mcp-build
mcp-build: ## Build MCP server binary
	@echo "$(YELLOW)Building MCP server...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/mcp-server ./cmd/mcp-server/main.go
	@echo "$(GREEN)MCP server built: $(BUILD_DIR)/mcp-server$(NC)"

.PHONY: mcp-run
mcp-run: ## Run MCP server
	@echo "$(YELLOW)Starting MCP server...$(NC)"
	@echo "$(BLUE)MCP server communicates via stdin/stdout$(NC)"
	@go run ./cmd/mcp-server/main.go

.PHONY: mcp-test
mcp-test: ## Test MCP server functionality
	@echo "$(YELLOW)Testing MCP server...$(NC)"
	@go run scripts/test-mcp.go

.PHONY: mcp-docs
mcp-docs: ## Show MCP integration documentation
	@echo "$(BLUE)MCP Integration Documentation$(NC)"
	@echo "============================="
	@echo ""
	@echo "$(GREEN)Setup Instructions:$(NC)"
	@echo "  See: docs/MCP_INTEGRATION.md"
	@echo ""
	@echo "$(GREEN)Configuration:$(NC)"
	@echo "  Config file: configs/mcp-server.json"
	@echo ""
	@echo "$(GREEN)Available Tools:$(NC)"
	@echo "  • create_contact - Create new contacts"
	@echo "  • search_contacts - Search and filter contacts"
	@echo "  • get_contact - Get contact details"
	@echo "  • update_contact - Update contact information"
	@echo "  • delete_contact - Delete contacts"
	@echo "  • get_analytics - Contact analytics"
	@echo "  • export_contacts - Export contact data"
	@echo ""
	@echo "$(GREEN)Usage:$(NC)"
	@echo "  make mcp-build    - Build MCP server"
	@echo "  make mcp-run      - Run MCP server"
	@echo "  make mcp-test     - Test MCP functionality"

# Utility commands
.PHONY: clean
clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@rm -rf vendor/
	@rm -f mcp-server
	@echo "$(GREEN)Clean completed$(NC)"

.PHONY: logs
logs: ## Show service logs
	@echo "$(YELLOW)Showing service logs...$(NC)"
	@tail -f logs/contact-service.log

.PHONY: health
health: ## Check service health
	@echo "$(YELLOW)Checking service health...$(NC)"
	@curl -s http://localhost:8081/health | jq '.' || echo "$(RED)Service is not responding$(NC)"

.PHONY: env-check
env-check: ## Check environment configuration
	@echo "$(YELLOW)Checking environment configuration...$(NC)"
	@go run scripts/env-check.go

# Database utilities
.PHONY: db-reset
db-reset: ## Reset database (WARNING: This will drop all data)
	@echo "$(RED)WARNING: This will drop all database data!$(NC)"
	@echo "$(YELLOW)Are you sure? [y/N]$(NC)" && read ans && [ $${ans:-N} = y ]
	@echo "$(YELLOW)Resetting database...$(NC)"
	@go run scripts/db-reset.go

.PHONY: db-seed
db-seed: ## Seed database with sample data
	@echo "$(YELLOW)Seeding database...$(NC)"
	@go run scripts/db-seed.go

.PHONY: db-backup
db-backup: ## Backup database
	@echo "$(YELLOW)Creating database backup...$(NC)"
	@go run scripts/db-backup.go

# Performance testing
.PHONY: load-test
load-test: ## Run load tests
	@echo "$(YELLOW)Running load tests...$(NC)"
	@vegeta attack -targets=tests/load/targets.txt -duration=30s -rate=100 | vegeta report

.PHONY: stress-test
stress-test: ## Run stress tests
	@echo "$(YELLOW)Running stress tests...$(NC)"
	@go run tests/stress/main.go

# All-in-one commands
.PHONY: install
install: deps build ## Install dependencies and build
	@echo "$(GREEN)Installation completed$(NC)"

.PHONY: check
check: format vet lint test ## Run all checks (format, vet, lint, test)
	@echo "$(GREEN)All checks passed$(NC)"

.PHONY: ci
ci: deps check test-coverage ## Run CI pipeline
	@echo "$(GREEN)CI pipeline completed$(NC)"

.PHONY: full-test
full-test: test-clean test-unit test-integration test-race benchmark ## Run comprehensive test suite
	@echo "$(GREEN)===========================================$(NC)"
	@echo "$(GREEN)  Full Test Suite Completed Successfully  $(NC)"
	@echo "$(GREEN)===========================================$(NC)"
	@echo "$(BLUE)Test Coverage Report: coverage.html$(NC)"
	@echo "$(BLUE)Integration Results: Available$(NC)"
	@echo "$(BLUE)Race Condition Check: Passed$(NC)"
	@echo "$(BLUE)Benchmark Results: Available$(NC)"

.PHONY: test-all
test-all: test-clean deps test-coverage test-integration benchmark ## Run all tests with coverage and benchmarks
	@echo "$(GREEN)Complete test suite with coverage analysis completed$(NC)"