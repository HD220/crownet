# Makefile for the CrowNet project

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOLINT=golangci-lint # Assumes golangci-lint is installed and in PATH

# Target binary name
BINARY_NAME=crownet

# Default target executed when no arguments are given to make.
.PHONY: all
all: lint test build

# Build the application binary.
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	$(GOBUILD) -o $(BINARY_NAME) main.go
	@echo "$(BINARY_NAME) built successfully."

# Run tests.
# Note: Test execution may currently be hindered by tool issues in the development environment.
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) ./... -v

# Run linter.
# Note: golangci-lint needs to be installed (e.g., via `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`).
# Linting may currently be hindered by tool issues in the development environment.
.PHONY: lint
lint:
	@echo "Running linter..."
	$(GOLINT) run ./...

# Clean build artifacts and test cache.
.PHONY: clean
clean:
	@echo "Cleaning up..."
	$(GOCLEAN) -testcache
	rm -f $(BINARY_NAME)
	@echo "Cleanup complete."

# Help target to display available commands.
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make all          : Run lint, test, and build (default)"
	@echo "  make build        : Build the application binary ($(BINARY_NAME))"
	@echo "  make test         : Run all tests verbosely"
	@echo "  make lint         : Run golangci-lint"
	@echo "  make clean        : Remove build artifacts and test cache"
	@echo "  make help         : Show this help message"
