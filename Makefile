.PHONY: build run test clean docker docker-run help

# Default target
all: build

# Build the binary
build:
	@echo "Building jwks-mock-api..."
	@go build -o jwks-mock-api ./cmd/jwks-mock-api

# Build with optimizations (smaller binary)
build-optimized:
	@echo "Building optimized jwks-mock-api..."
	@CGO_ENABLED=0 go build -ldflags='-w -s' -o jwks-mock-api ./cmd/jwks-mock-api

# Run the application
run:
	@echo "Running jwks-mock-api..."
	@go run ./cmd/jwks-mock-api

# Run with config file
run-config:
	@echo "Running jwks-mock-api with config file..."
	@go run ./cmd/jwks-mock-api -config config.yaml.example

# Test the application
test:
	@echo "Running tests..."
	@go test ./...

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -cover ./...



# Run Docker-based integration tests
test-integration:
	@echo "Running Docker integration tests..."
	@docker compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from integration-tests
	@docker compose -f docker-compose.test.yml down

# Run integration tests with external Docker setup (for local development)
test-integration-external:
	@echo "Starting JWKS API server for external testing..."
	@docker compose -f docker-compose.test.yml up -d jwks-api
	@echo "Waiting for server to be ready..."
	@sleep 10
	@echo "Running integration tests against external server..."
	@JWKS_API_URL=http://localhost:3001 go test -v ./test/integration/...
	@echo "Cleaning up external test server..."
	@docker compose -f docker-compose.test.yml down

# Run all tests (integration only)
test-all:
	@echo "Running all tests..."
	@$(MAKE) test-integration

# Generate coverage for integration tests
test-integration-coverage:
	@echo "Running Docker integration tests with coverage..."
	@mkdir -p ./test/integration/results
	@docker compose -f docker-compose.test.yml up --build --abort-on-container-exit --exit-code-from integration-tests
	@docker compose -f docker-compose.test.yml down

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	@rm -f jwks-mock-api
	@go clean

# Build Docker image
docker:
	@echo "Building Docker image..."
	@docker build -t jwks-mock-api:latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	@docker run -p 3000:3000 jwks-mock-api:latest

# Run Docker container with environment variables
docker-run-env:
	@echo "Running Docker container with custom environment..."
	@docker run -p 3000:3000 \
		-e JWT_ISSUER=http://localhost:3000 \
		-e JWT_AUDIENCE=dev-api \
		-e PORT=3000 \
		jwks-mock-api:latest

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Vet code
vet:
	@echo "Vetting code..."
	@go vet ./...

# Lint code (requires golangci-lint to be installed)
lint:
	@echo "Linting code..."
	@golangci-lint run

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  build-optimized - Build optimized binary (smaller size)"
	@echo "  run             - Run the application"
	@echo "  run-config      - Run with example config file"
	@echo "  test            - Run basic tests"
	@echo "  test-coverage   - Run tests with coverage"
	@echo "  test-integration- Run Docker-based integration tests"
	@echo "  test-integration-external - Run integration tests against external Docker server"
	@echo "  test-all        - Run all tests (integration only)"
	@echo "  clean           - Clean build artifacts"
	@echo "  docker          - Build Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-run-env  - Run Docker container with custom env vars"
	@echo "  fmt             - Format code"
	@echo "  vet             - Vet code"
	@echo "  lint            - Lint code"
	@echo "  deps            - Download and tidy dependencies"
	@echo "  help            - Show this help"