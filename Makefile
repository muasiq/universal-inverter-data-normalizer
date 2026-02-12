.PHONY: build run test clean docker docker-run lint fmt tidy

BINARY_NAME := normalizer
BUILD_DIR := ./bin
CMD_DIR := ./cmd/normalizer
DOCKER_IMAGE := universal-inverter-normalizer

# Build
build: tidy
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_DIR)
	@echo "Built: $(BUILD_DIR)/$(BINARY_NAME)"

# Run locally
run: build
	$(BUILD_DIR)/$(BINARY_NAME) --config config.yaml

# Run tests
test:
	go test -v -race -cover ./...

# Clean build artifacts
clean:
	@rm -rf $(BUILD_DIR)
	@echo "Cleaned"

# Go module tidy
tidy:
	go mod tidy

# Format code
fmt:
	gofmt -w .

# Lint (requires golangci-lint)
lint:
	golangci-lint run ./...

# Docker build
docker:
	docker build -t $(DOCKER_IMAGE):latest .

# Docker run
docker-run: docker
	docker run -p 8080:8080 \
		-v $(PWD)/config.yaml:/app/config.yaml:ro \
		$(DOCKER_IMAGE):latest

# Development: watch and rebuild (requires air)
dev:
	air -c .air.toml

# Show help
help:
	@echo "Available targets:"
	@echo "  build       Build the binary"
	@echo "  run         Build and run"
	@echo "  test        Run tests with race detector"
	@echo "  clean       Remove build artifacts"
	@echo "  tidy        Run go mod tidy"
	@echo "  fmt         Format Go code"
	@echo "  lint        Run golangci-lint"
	@echo "  docker      Build Docker image"
	@echo "  docker-run  Build and run Docker container"
	@echo "  dev         Run with hot reload (requires air)"
