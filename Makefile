# Go Discord Bot Makefile

.PHONY: help build run test clean lint format check deps register setup-music start-ytdlp stop-ytdlp test-ytdlp

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
	@echo ""
	@echo "Music System:"
	@echo "  setup-music - Setup music dependencies (Python + yt-dlp)"
	@echo "  start-ytdlp - Start yt-dlp service manually"
	@echo "  stop-ytdlp  - Stop yt-dlp service"
	@echo "  test-ytdlp  - Test yt-dlp service functionality"

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

# Music System Targets

# Setup music dependencies
setup-music:
	@echo "Setting up music dependencies..."
	@./scripts/setup-music.sh

# Start yt-dlp service manually (for testing)
start-ytdlp:
	@echo "Starting yt-dlp service on port 8080..."
	@python3 services/ytdlp/server.py --host localhost --port 8080 &
	@echo "yt-dlp service started. PID: $$!"

# Stop yt-dlp service
stop-ytdlp:
	@echo "Stopping yt-dlp service..."
	@pkill -f "services/ytdlp/server.py" || true
	@echo "yt-dlp service stopped"

# Test yt-dlp service functionality
test-ytdlp:
	@echo "Testing yt-dlp service..."
	@python3 -c "import yt_dlp; print('✅ yt-dlp import successful')"
	@python3 services/ytdlp/server.py --host localhost --port 8081 & \
	SERVER_PID=$$!; \
	sleep 3; \
	curl -s http://localhost:8081/health > /dev/null && echo "✅ yt-dlp service health check passed" || echo "❌ yt-dlp service health check failed"; \
	kill $$SERVER_PID 2>/dev/null || true

# Complete setup workflow
setup-complete: deps setup-music build
	@echo "🎉 Complete setup finished! Bot is ready to run."
	@echo "Use 'make register' to register commands, then 'make run' to start the bot."