# Makefile for go-smart-queue

.PHONY: all build test clean install example help

# Default target
all: build

# Build the main CLI application
build:
	@echo "Building go-smart-queue..."
	@go build -o go-smart-queue .

# Run tests
test:
	@echo "Running tests..."
	@go test -v .

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -cover -coverprofile=coverage.out .
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f go-smart-queue
	@rm -f coverage.out coverage.html
	@rm -f examples/example-run

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@go mod download
	@go mod tidy

# Run the example
example:
	@echo "Running example..."
	@cd examples && ./run-example.sh

# Install the CLI globally
install: build
	@echo "Installing go-smart-queue to $(GOPATH)/bin..."
	@cp go-smart-queue $(GOPATH)/bin/

# Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build the CLI application"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Install dependencies"
	@echo "  example       - Run the example program"
	@echo "  install       - Install CLI globally"
	@echo "  help          - Show this help message"
