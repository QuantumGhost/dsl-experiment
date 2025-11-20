.PHONY: all build test lint format clean help

# Binary name
APP_NAME := dsl-run
# Main package path
CMD_PATH := ./cmd/dsl-run

# Default target
all: format lint test build

# Build the application
build:
	@echo "Building $(APP_NAME)..."
	go build -o $(APP_NAME) $(CMD_PATH)

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Lint the code
lint:
	@echo "Linting..."
	go vet ./...
	@echo "Checking sealed interfaces..."
	go-sumtype ./...

# Format the code
format:
	@echo "Formatting..."
	go fmt ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	go clean

# Show help
help:
	@echo "Makefile commands:"
	@echo "  make build    - Build the application"
	@echo "  make test     - Run tests with race detection and coverage"
	@echo "  make lint     - Run go vet and go-sumtype checks"
	@echo "  make format   - Format code using go fmt"
	@echo "  make clean    - Remove build artifacts"
	@echo "  make all      - Run format, lint, test, and build"
# Frontend commands
web-install:
	@echo "Installing frontend dependencies..."
	cd web && npm install

web-dev:
	@echo "Starting frontend dev server..."
	cd web && npm run dev

web-build:
	@echo "Building frontend..."
	cd web && npm install && npm run build
