# Makefile for API Direct CLI

# Variables
BINARY_NAME := apidirect
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w \
	-X github.com/api-direct/cli/cmd.Version=$(VERSION) \
	-X github.com/api-direct/cli/cmd.BuildDate=$(BUILD_DATE) \
	-X github.com/api-direct/cli/cmd.GitCommit=$(GIT_COMMIT)

# Go commands
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Directories
BUILD_DIR := build
DIST_DIR := dist

# Colors
GREEN := \033[0;32m
BLUE := \033[0;34m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

.PHONY: all build test clean install fmt vet lint deps help

# Default target
all: deps fmt vet test build

# Build the binary
build:
	@echo -e "$(BLUE)Building $(BINARY_NAME) v$(VERSION)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo -e "$(GREEN)✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

# Build for all platforms
build-all:
	@echo -e "$(BLUE)Building for all platforms...$(NC)"
	@./scripts/build.sh

# Run tests
test:
	@echo -e "$(BLUE)Running tests...$(NC)"
	@$(GOTEST) -v -cover ./...

# Run tests with coverage report
test-coverage:
	@echo -e "$(BLUE)Running tests with coverage...$(NC)"
	@$(GOTEST) -v -coverprofile=coverage.out ./...
	@$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo -e "$(GREEN)✅ Coverage report: coverage.html$(NC)"

# Clean build artifacts
clean:
	@echo -e "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html
	@echo -e "$(GREEN)✅ Clean complete$(NC)"

# Install to system
install: build
	@echo -e "$(BLUE)Installing $(BINARY_NAME)...$(NC)"
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo -e "$(GREEN)✅ Installed to /usr/local/bin/$(BINARY_NAME)$(NC)"

# Install locally (to GOPATH/bin)
install-local: build
	@echo -e "$(BLUE)Installing $(BINARY_NAME) locally...$(NC)"
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/
	@echo -e "$(GREEN)✅ Installed to $(GOPATH)/bin/$(BINARY_NAME)$(NC)"

# Format code
fmt:
	@echo -e "$(BLUE)Formatting code...$(NC)"
	@$(GOFMT) ./...

# Vet code
vet:
	@echo -e "$(BLUE)Vetting code...$(NC)"
	@$(GOVET) ./...

# Run linter (requires golangci-lint)
lint:
	@echo -e "$(BLUE)Running linter...$(NC)"
	@golangci-lint run

# Update dependencies
deps:
	@echo -e "$(BLUE)Updating dependencies...$(NC)"
	@$(GOMOD) download
	@$(GOMOD) tidy
	@echo -e "$(GREEN)✅ Dependencies updated$(NC)"

# Run the CLI locally
run:
	@$(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

# Development mode with hot reload (requires air)
dev:
	@echo -e "$(BLUE)Starting development mode...$(NC)"
	@air

# Docker build
docker-build:
	@echo -e "$(BLUE)Building Docker image...$(NC)"
	@docker build -t apidirect-cli:$(VERSION) .
	@echo -e "$(GREEN)✅ Docker image built: apidirect-cli:$(VERSION)$(NC)"

# Generate mocks (requires mockgen)
mocks:
	@echo -e "$(BLUE)Generating mocks...$(NC)"
	@go generate ./...

# Check for security vulnerabilities
security:
	@echo -e "$(BLUE)Checking for security vulnerabilities...$(NC)"
	@gosec -quiet ./...
	@echo -e "$(GREEN)✅ Security check complete$(NC)"

# Show version
version:
	@echo $(VERSION)

# Help
help:
	@echo "API Direct CLI Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  all            - Run deps, fmt, vet, test, and build"
	@echo "  build          - Build the binary for current platform"
	@echo "  build-all      - Build binaries for all platforms"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install to /usr/local/bin (requires sudo)"
	@echo "  install-local  - Install to GOPATH/bin"
	@echo "  fmt            - Format code"
	@echo "  vet            - Vet code"
	@echo "  lint           - Run linter (requires golangci-lint)"
	@echo "  deps           - Update dependencies"
	@echo "  run            - Run the CLI with ARGS"
	@echo "  dev            - Run in development mode (requires air)"
	@echo "  docker-build   - Build Docker image"
	@echo "  mocks          - Generate mocks"
	@echo "  security       - Check for security vulnerabilities"
	@echo "  version        - Show version"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  make run ARGS='--help'"