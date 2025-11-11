# Makefile for wink CLI
# Constitution: Build and test automation

.PHONY: build test lint install clean coverage integration-test bench help

# Binary name
BINARY_NAME=wink
VERSION?=dev
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) cmd/wink/main.go

# Run all tests
test:
	@echo "Running tests..."
	go test -v -race -cover ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -race -cover ./tests/unit/...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v -race -cover -tags=integration ./tests/integration/...

# Test coverage (Constitution: ≥90% target)
coverage:
	@echo "Generating coverage report..."
	go test -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"
	@go tool cover -func=coverage.out | grep total

# Run linter (Constitution: Code Quality First)
lint:
	@echo "Running golangci-lint..."
	golangci-lint run ./...

# Format code
fmt:
	@echo "Formatting code..."
	gofmt -s -w .
	goimports -w .

# Install the binary
install: build
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) ./cmd/wink

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/

# Run benchmarks (Constitution: Performance Requirements)
bench:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Cross-compile for multiple platforms
build-all:
	@echo "Building for multiple platforms..."
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 cmd/wink/main.go
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 cmd/wink/main.go
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 cmd/wink/main.go
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe cmd/wink/main.go

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build the binary"
	@echo "  test               - Run all tests"
	@echo "  test-unit          - Run unit tests only"
	@echo "  test-integration   - Run integration tests only"
	@echo "  coverage           - Generate test coverage report (target: ≥90%)"
	@echo "  lint               - Run golangci-lint"
	@echo "  fmt                - Format code"
	@echo "  install            - Install binary to GOPATH/bin"
	@echo "  clean              - Remove build artifacts"
	@echo "  bench              - Run benchmarks"
	@echo "  build-all          - Cross-compile for Linux/macOS/Windows"
	@echo "  help               - Show this help"
