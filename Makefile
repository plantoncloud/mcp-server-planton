.PHONY: build install test lint fmt fmt-check release docker-build docker-run clean help

# Default target
.DEFAULT_GOAL := help

# Binary name
BINARY_NAME := mcp-server-planton
BINARY_PATH := bin/$(BINARY_NAME)

# Docker image
DOCKER_IMAGE := mcp-server-planton:local
GHCR_IMAGE := ghcr.io/plantoncloud-inc/mcp-server-planton

## build: Build the binary for local architecture
build: fmt-check
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	@go build -o $(BINARY_PATH) ./cmd/mcp-server-planton
	@echo "Binary built: $(BINARY_PATH)"

## install: Install the binary to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@go install ./cmd/mcp-server-planton
	@echo "Binary installed to GOPATH/bin"

## install-local: Build and install binary to /usr/local/bin (requires sudo)
install-local: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@sudo cp $(BINARY_PATH) /usr/local/bin/$(BINARY_NAME)
	@sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Binary installed to /usr/local/bin/$(BINARY_NAME)"
	@echo "You can now use '$(BINARY_NAME)' command globally"

## test: Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

## lint: Run linter (requires golangci-lint)
lint:
	@echo "Running linter..."
	@golangci-lint run

## fmt: Format Go code
fmt:
	@echo "Formatting Go code..."
	@gofmt -w .
	@echo "Code formatted"

## fmt-check: Check if Go code is formatted
fmt-check:
	@echo "Checking Go code formatting..."
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Go code is not formatted:"; \
		gofmt -l .; \
		echo "Run 'make fmt' to fix formatting"; \
		exit 1; \
	fi
	@echo "All Go code is properly formatted"

## docker-build: Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE) .
	@echo "Docker image built: $(DOCKER_IMAGE)"

## docker-run: Run Docker image with environment variables
docker-run:
	@echo "Running Docker container..."
	@docker run -i --rm \
		-e PLANTON_API_KEY=$(PLANTON_API_KEY) \
		-e PLANTON_APIS_GRPC_ENDPOINT=$(PLANTON_APIS_GRPC_ENDPOINT) \
		$(DOCKER_IMAGE)

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@rm -rf dist/
	@echo "Clean complete"

## release: Create and push a release version (usage: make release version=v1.0.0 [force=true])
release: build
ifndef version
	@echo "Error: version is required. Usage: make release version=v1.0.0"
	@exit 1
endif
	@echo "Creating release version $(version)..."
	@if ! echo "$(version)" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+'; then \
		echo "Error: version must follow semantic versioning (e.g., v1.0.0, v2.1.3)"; \
		exit 1; \
	fi
	@# Check if tag exists locally
	@if git rev-parse $(version) >/dev/null 2>&1; then \
		if [ "$(force)" = "true" ]; then \
			echo "Tag $(version) exists locally. Deleting due to force=true..."; \
			git tag -d $(version); \
		else \
			echo "Error: Tag $(version) already exists locally."; \
			echo "Use 'make release version=$(version) force=true' to force delete and recreate."; \
			exit 1; \
		fi \
	fi
	@# Check if tag exists remotely
	@if git ls-remote --tags origin | grep -q "refs/tags/$(version)$$"; then \
		if [ "$(force)" = "true" ]; then \
			echo "Tag $(version) exists remotely. Deleting due to force=true..."; \
			git push origin :refs/tags/$(version); \
		else \
			echo "Error: Tag $(version) already exists remotely."; \
			echo "Use 'make release version=$(version) force=true' to force delete and recreate."; \
			exit 1; \
		fi \
	fi
	@git tag -a $(version) -m "Release $(version)"
	@git push origin $(version)
	@echo "Release version $(version) created and pushed"
	@echo "GitHub Actions will now build and publish the release"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

