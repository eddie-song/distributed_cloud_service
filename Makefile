.PHONY: build test clean run docker-build docker-up docker-down help

# Build binaries
build:
	@echo "Building node..."
	@go build -o bin/node ./cmd/node
	@echo "Building cloudctl..."
	@go build -o bin/cloudctl ./cmd/cloudctl
	@echo "Build complete!"

# Run all tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf data/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

# Run node1 (local)
run-node1:
	@go run ./cmd/node -config configs/node1.yaml

# Run node2 (local)
run-node2:
	@go run ./cmd/node -config configs/node2.yaml

# Run node3 (local)
run-node3:
	@go run ./cmd/node -config configs/node3.yaml

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Format complete!"

# Lint code
lint:
	@echo "Linting code..."
	@golangci-lint run ./... || echo "Install golangci-lint: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# Docker build
docker-build:
	@echo "Building Docker image..."
	@docker compose build

# Docker up
docker-up:
	@echo "Starting cluster..."
	@docker compose up -d

# Docker down
docker-down:
	@echo "Stopping cluster..."
	@docker compose down -v

# Docker logs
docker-logs:
	@docker compose logs -f

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build node and cloudctl binaries"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Remove build artifacts and data"
	@echo "  run-node1      - Run node1 locally"
	@echo "  run-node2      - Run node2 locally"
	@echo "  run-node3      - Run node3 locally"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code (requires golangci-lint)"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-up      - Start cluster with Docker Compose"
	@echo "  docker-down    - Stop cluster"
	@echo "  docker-logs    - View cluster logs"
	@echo "  help           - Show this help message"

