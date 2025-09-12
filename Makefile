# Go Discord Bot Makefile

.PHONY: help build run test clean lint format check deps register

# Default target
help:
	@echo "Available targets:"
	@echo "  build       - Build the bot binary"
	@echo "  run         - Run the bot"
	@echo "  register    - Register bot commands with Discord"
	@echo "  test        - Run tests"
	@echo "  test-cover  - Run tests with coverage report"
	@echo "  lint        - Run linter"
	@echo "  format      - Format code with goimports"
	@echo "  check       - Run format and lint checks"
	@echo "  deps        - Download and tidy dependencies"
	@echo "  clean       - Clean build artifacts"

# Build the bot
build:
	go build -o pxnx-discord-bot main.go

# Run the bot
run:
	go run main.go

# Register commands with Discord
register:
	go run main.go --register-commands

# Run tests
test:
	go test -v

# Run tests with coverage
test-cover:
	go test -cover -v

# Run linter
lint:
	golangci-lint run --timeout=2m

# Format code with goimports
format:
	$(shell go env GOPATH)/bin/goimports -w *.go

# Check formatting and linting
check: format lint
	@echo "Code formatting and linting complete!"

# Download and organize dependencies
deps:
	go mod tidy

# Clean build artifacts
clean:
	rm -f pxnx-discord-bot
	go clean

# Development workflow - run before committing
dev-check: format lint test
	@echo "All development checks passed!"