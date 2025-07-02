# Makefile for E2E Testing

.PHONY: all test test-all test-byoa test-mock test-integration test-quick clean help

# Default test timeout
TIMEOUT := 30m

# Test binary name
TEST_BINARY := apidirect

# Colors for output
GREEN := \033[0;32m
RED := \033[0;31m
BLUE := \033[0;34m
NC := \033[0m

all: test

## test: Run all E2E tests
test: test-all

## test-all: Run complete E2E test suite
test-all:
	@echo "$(BLUE)Running all E2E tests...$(NC)"
	@./run_tests.sh -m all -v

## test-byoa: Run BYOA deployment tests
test-byoa:
	@echo "$(BLUE)Running BYOA deployment tests...$(NC)"
	@./run_tests.sh -m byoa -v

## test-hosted: Run hosted deployment tests
test-hosted:
	@echo "$(BLUE)Running hosted deployment tests...$(NC)"
	@./run_tests.sh -m hosted -v

## test-modes: Run deployment mode comparison tests
test-modes:
	@echo "$(BLUE)Running deployment mode comparison tests...$(NC)"
	@./run_tests.sh -m modes -v

## test-mock: Run tests with mock AWS services
test-mock:
	@echo "$(BLUE)Running mock AWS tests...$(NC)"
	@./run_tests.sh -m mock -v

## test-integration: Run integration tests
test-integration:
	@echo "$(BLUE)Running integration tests...$(NC)"
	@RUN_INTEGRATION_TESTS=true ./run_tests.sh -m integration -v

## test-quick: Run quick tests without AWS
test-quick:
	@echo "$(BLUE)Running quick tests...$(NC)"
	@./run_tests.sh -m quick -v

## test-specific: Run specific test by name (use TEST=TestName)
test-specific:
	@echo "$(BLUE)Running test: $(TEST)$(NC)"
	@go test -v -timeout $(TIMEOUT) -run $(TEST) .

## build-cli: Build the CLI binary
build-cli:
	@echo "$(BLUE)Building CLI binary...$(NC)"
	@cd ../.. && go build -o $(TEST_BINARY)
	@echo "$(GREEN)✓ CLI built successfully$(NC)"

## setup: Install test dependencies
setup:
	@echo "$(BLUE)Installing test dependencies...$(NC)"
	@go mod download
	@go install github.com/stretchr/testify/assert@latest
	@echo "$(GREEN)✓ Dependencies installed$(NC)"

## coverage: Run tests with coverage
coverage:
	@echo "$(BLUE)Running tests with coverage...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report generated: coverage.html$(NC)"

## benchmark: Run benchmark tests
benchmark:
	@echo "$(BLUE)Running benchmarks...$(NC)"
	@go test -bench=. -benchmem ./...

## clean: Clean test artifacts
clean:
	@echo "$(BLUE)Cleaning test artifacts...$(NC)"
	@rm -f coverage.out coverage.html
	@rm -rf test-reports/
	@rm -f $(TEST_BINARY)
	@echo "$(GREEN)✓ Cleaned$(NC)"

## check-aws: Check AWS credentials and configuration
check-aws:
	@echo "$(BLUE)Checking AWS configuration...$(NC)"
	@aws sts get-caller-identity || (echo "$(RED)✗ AWS credentials not configured$(NC)" && exit 1)
	@echo "$(GREEN)✓ AWS credentials valid$(NC)"

## check-deps: Check required dependencies
check-deps:
	@echo "$(BLUE)Checking dependencies...$(NC)"
	@command -v aws >/dev/null 2>&1 || (echo "$(RED)✗ AWS CLI not installed$(NC)" && exit 1)
	@command -v terraform >/dev/null 2>&1 || (echo "$(RED)✗ Terraform not installed$(NC)" && exit 1)
	@command -v go >/dev/null 2>&1 || (echo "$(RED)✗ Go not installed$(NC)" && exit 1)
	@echo "$(GREEN)✓ All dependencies satisfied$(NC)"

## lint: Run linters
lint:
	@echo "$(BLUE)Running linters...$(NC)"
	@golangci-lint run ./... || go fmt ./...
	@echo "$(GREEN)✓ Linting complete$(NC)"

## help: Show this help message
help:
	@echo "API-Direct E2E Testing"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' Makefile | sed 's/## /  /' | column -t -s ':'

# Watch for file changes and run tests
watch:
	@echo "$(BLUE)Watching for changes...$(NC)"
	@while true; do \
		inotifywait -r -e modify,create,delete *.go 2>/dev/null || fswatch -r *.go; \
		clear; \
		make test-quick; \
	done

# Run tests in Docker
docker-test:
	@echo "$(BLUE)Running tests in Docker...$(NC)"
	@docker build -t apidirect-test -f ../../Dockerfile.test ../..
	@docker run --rm \
		-e AWS_ACCESS_KEY_ID \
		-e AWS_SECRET_ACCESS_KEY \
		-e AWS_REGION \
		apidirect-test

# Generate test report
report:
	@echo "$(BLUE)Generating test report...$(NC)"
	@mkdir -p test-reports
	@go test -v -json ./... > test-reports/results.json
	@echo "$(GREEN)✓ Test report generated: test-reports/results.json$(NC)"

# CI/CD friendly test command
ci-test:
	@echo "$(BLUE)Running CI tests...$(NC)"
	@go test -v -timeout $(TIMEOUT) -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

.DEFAULT_GOAL := help