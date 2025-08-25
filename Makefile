# Servo Makefile

# Variables
BINARY_NAME=servo
BUILD_DIR=build
DIST_DIR=dist
GO_FILES=$(shell find . -name "*.go" -type f)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build: $(BUILD_DIR)/$(BINARY_NAME)

$(BUILD_DIR)/$(BINARY_NAME): $(GO_FILES)
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/servo

# Build for multiple platforms
.PHONY: build-all
build-all: clean
	@echo "Building for multiple platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/servo
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/servo
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/servo
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/servo
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/servo

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	go test -v ./...

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run tests with race detection
.PHONY: test-race
test-race:
	@echo "Running tests with race detection..."
	go test -v -race ./...

# Run benchmarks
.PHONY: bench
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Run go-critic
.PHONY: critic
critic:
	@echo "Running go-critic..."
	@which gocritic > /dev/null || (echo "Installing go-critic..." && go install github.com/go-critic/go-critic/cmd/gocritic@latest)
	gocritic check ./...

# Format the code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@which goimports > /dev/null || (echo "Installing goimports..." && go install golang.org/x/tools/cmd/goimports@latest)
	goimports -w .

# Vet the code
.PHONY: vet
vet:
	@echo "Vetting code..."
	go vet ./...

# Run all quality checks
.PHONY: check
check: fmt vet critic test

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

# Install the binary to GOPATH/bin
.PHONY: install
install: build
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/servo

# Uninstall the binary
.PHONY: uninstall
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	rm -f $(shell go env GOPATH)/bin/$(BINARY_NAME)

# Run the binary
.PHONY: run
run: build
	@echo "Running $(BINARY_NAME)..."
	./$(BUILD_DIR)/$(BINARY_NAME)

# Development setup
.PHONY: dev-setup
dev-setup: deps
	@echo "Setting up development environment..."
	@which goimports > /dev/null || go install golang.org/x/tools/cmd/goimports@latest
	@echo "Development environment ready!"

# Generate mocks (if using mockgen)
.PHONY: mocks
mocks:
	@echo "Generating mocks..."
	@which mockgen > /dev/null || go install github.com/golang/mock/mockgen@latest
	go generate ./...

# Update dependencies
.PHONY: update-deps
update-deps:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Security scan
.PHONY: security
security:
	@echo "Running security scan..."
	@which gosec > /dev/null || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	gosec ./...

# Docker build
.PHONY: docker-build
docker-build:
	@echo "Building Docker image..."
	docker build -t servo:$(VERSION) .
	docker tag servo:$(VERSION) servo:latest

# Create example .servo files
.PHONY: examples
examples:
	@echo "Creating example .servo files..."
	@mkdir -p examples
	@echo "Example .servo files created in examples/"

# Validate example files
.PHONY: validate-examples
validate-examples: build examples
	@echo "Validating example .servo files..."
	./$(BUILD_DIR)/$(BINARY_NAME) validate examples/graphiti.servo

# Integration test setup
.PHONY: integration-setup
integration-setup:
	@echo "Setting up integration test environment..."
	@command -v docker >/dev/null 2>&1 || { echo "Docker is required for integration tests"; exit 1; }
	docker pull neo4j:5.13
	docker pull redis:7-alpine

# Run integration tests
.PHONY: test-integration
test-integration: integration-setup
	@echo "Running integration tests..."
	go test -v -tags=integration ./test/integration/...

# Performance tests
.PHONY: test-performance
test-performance:
	@echo "Running performance tests..."
	go test -v -tags=performance -timeout=30m ./test/performance/...

# Quick development cycle
.PHONY: dev
dev: fmt vet test build

# Pre-commit checks
.PHONY: pre-commit
pre-commit: check test-race

# Release preparation
.PHONY: release-prep
release-prep: clean check test-race build-all
	@echo "Release preparation complete!"
	@echo "Built binaries:"
	@ls -la $(DIST_DIR)/

# Generate documentation
.PHONY: docs
docs:
	@echo "Generating documentation..."
	@which godoc > /dev/null || go install golang.org/x/tools/cmd/godoc@latest
	@echo "Run 'godoc -http=:6060' to view documentation at http://localhost:6060"

# Show help
.PHONY: help
help:
	@echo "Servo Makefile Commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  build          Build the binary"
	@echo "  build-all      Build for multiple platforms"
	@echo "  install        Install binary to GOPATH/bin"
	@echo "  uninstall      Remove binary from GOPATH/bin"
	@echo ""
	@echo "Development Commands:"
	@echo "  dev-setup      Set up development environment"
	@echo "  dev            Quick development cycle (fmt, vet, test, build)"
	@echo "  run            Build and run the binary"
	@echo ""
	@echo "Quality Commands:"
	@echo "  test           Run tests"
	@echo "  test-coverage  Run tests with coverage report"
	@echo "  test-race      Run tests with race detection"
	@echo "  bench          Run benchmarks"
	@echo "  critic         Run go-critic"
	@echo "  fmt            Format code"
	@echo "  vet            Vet code"
	@echo "  check          Run all quality checks"
	@echo "  security       Run security scan"
	@echo "  pre-commit     Pre-commit checks"
	@echo ""
	@echo "Test Commands:"
	@echo "  test-integration   Run integration tests"
	@echo "  test-performance   Run performance tests"
	@echo "  integration-setup  Set up integration test environment"
	@echo ""
	@echo "Example Commands:"
	@echo "  examples           Create example .servo files"
	@echo "  validate-examples  Validate example files"
	@echo ""
	@echo "Utility Commands:"
	@echo "  clean          Clean build artifacts"
	@echo "  deps           Install dependencies"
	@echo "  update-deps    Update dependencies"
	@echo "  mocks          Generate mocks"
	@echo "  docs           Generate documentation"
	@echo "  release-prep   Prepare for release"
	@echo "  help           Show this help"

# Default help target
.DEFAULT_GOAL := help
