.PHONY: help build run test clean docker-build docker-run install lint fmt

# Variables
APP_NAME=kite
VERSION=4.0.0
GO=go
GOFLAGS=-v
BINARY_DIR=bin
DOCKER_IMAGE=$(APP_NAME):$(VERSION)

# Help target
help:
	@echo "Kite v4 - Makefile Commands"
	@echo ""
	@echo "Available targets:"
	@echo "  make build         - Build the API server binary"
	@echo "  make run           - Run the API server"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo "  make swagger       - Generate Swagger/OpenAPI documentation"
	@echo "  make docker-build  - Build Docker image"
	@echo "  make docker-run    - Run Docker container"
	@echo "  make install       - Install dependencies"
	@echo "  make lint          - Run linters"
	@echo "  make fmt           - Format code"

# Build the API server
build:
	@echo "Building $(APP_NAME) API server..."
	@mkdir -p $(BINARY_DIR)
	$(GO) build $(GOFLAGS) -o $(BINARY_DIR)/kite-api ./cmd/kite-api

# Run the API server
run: build
	@echo "Running $(APP_NAME) API server..."
	./$(BINARY_DIR)/kite-api

# Run tests
test:
	@echo "Running tests..."
	$(GO) test $(GOFLAGS) ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BINARY_DIR)
	rm -f kite.db
	$(GO) clean

# Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE)..."
	docker build -t $(DOCKER_IMAGE) .

# Run Docker container
docker-run: docker-build
	@echo "Running Docker container..."
	docker run -p 8080:8080 -p 9091:9091 $(DOCKER_IMAGE)

# Run with docker-compose
docker-compose-up:
	@echo "Starting services with docker-compose..."
	docker-compose up -d

# Stop docker-compose services
docker-compose-down:
	@echo "Stopping docker-compose services..."
	docker-compose down

# Install dependencies
install:
	@echo "Installing dependencies..."
	$(GO) mod download
	$(GO) mod tidy

# Run linters (requires golangci-lint)
lint:
	@echo "Running linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Skipping..."; \
	fi

# Format code
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

# Development mode with hot reload (requires air)
dev:
	@echo "Starting development server with hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "air not installed. Run: go install github.com/air-verse/air@latest"; \
		make run; \
	fi

# Generate Swagger/OpenAPI documentation
swagger:
	@echo "Generating Swagger documentation..."
	@if command -v swag >/dev/null 2>&1; then \
		swag init -g internal/api/swagger.go -o docs --parseDependency --parseInternal; \
	else \
		echo "swag not installed. Run: go install github.com/swaggo/swag/cmd/swag@latest"; \
	fi

# Generate godoc documentation
docs:
	@echo "Generating documentation..."
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060; \
	else \
		echo "godoc not installed. Run: go install golang.org/x/tools/cmd/godoc@latest"; \
	fi

# Show project status
status:
	@echo "Project: $(APP_NAME) v$(VERSION)"
	@echo "Go version: $$($(GO) version)"
	@echo "Dependencies status:"
	@$(GO) list -m all | head -20
