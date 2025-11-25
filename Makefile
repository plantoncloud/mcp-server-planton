.PHONY: build install test lint docker-build docker-run clean help

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME := mcp-server-planton
BINARY_PATH := bin/$(BINARY_NAME)

# Docker image
DOCKER_IMAGE := mcp-server-planton:local
GHCR_IMAGE := ghcr.io/plantoncloud-inc/mcp-server-planton

## build: Build the binary for local architecture
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) ./cmd/mcp-server-planton
	@echo "Binary built: $(BINARY_PATH)"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install ./cmd/mcp-server-planton
	@echo "Binary installed to GOPATH/bin"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

## docker-run: Run Docker image with environment variables
docker-run:
	@echo "Running Docker container..."
	@docker run -i --rm \
		-e USER_JWT_TOKEN=$(USER_JWT_TOKEN) \
		-e PLANTON_APIS_GRPC_ENDPOINT=$(PLANTON_APIS_GRPC_ENDPOINT) \
		$(DOCKER_IMAGE)

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf dist/
	@echo "Clean complete"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

