# PXNX Discord Bot

A modern Discord bot written in Go following TDD principles and best practices. Features include interactive commands, real-time weather data, and a production-ready music system with YouTube integration.

## 🚀 Features

### 🎵 Music System
- **`/join`** - Connect bot to voice channel with validation
- **`/leave`** - Disconnect and cleanup resources
- **`/play <song name or URL>`** - YouTube integration with search
  - Search by query: `/play lofi hip hop`
  - Direct URLs: `/play https://youtu.be/VIDEO_ID`
  - Rich embeds with metadata and thumbnails
  - ⚠️ **Current Status**: Infrastructure complete, investigating audio streaming issues

### 🎮 Commands
- **`/ping`** - Bot responsiveness test
- **`/peepee`** - Interactive command with emoji reactions
- **`/8ball`** - Magic 8-ball responses
- **`/coinflip`** - Random coin flip
- **`/roll [max]`** - Dice rolling (1-1000000)
- **`/server`** - Server information display
- **`/user [target]`** - User profile information
- **`/weather <location>`** - Real weather data via OpenWeatherMap

### 🛠️ System Features
- **Event-driven architecture** with Discord gateway events
- **Service-oriented design** with separate yt-dlp HTTP service
- **Thread-safe operations** with comprehensive error handling
- **Production Docker deployment** with multi-architecture support
- **TDD development workflow** with comprehensive test coverage

## 📋 Quick Start

### Prerequisites
- **Go 1.25+**
- **Python 3.10+** (for music functionality)
- **Discord Bot Token** ([create here](https://discord.com/developers/applications))
- **OpenWeatherMap API Key** ([get free](https://openweathermap.org/api))

### Setup
```bash
# Clone repository
git clone https://github.com/Postmodum37/pxnx-discord-bot-go.git
cd pxnx-discord-bot-go

# Setup dependencies
go mod tidy
pip install -r services/ytdlp/requirements.txt

# Configure environment
cp .env.example .env
# Edit .env with your tokens

# Register commands (first time only)
go run main.go --register-commands

# Start bot
go run main.go
```

### Hot Reload Development
```bash
# Install air for hot reload
go install github.com/air-verse/air@latest

# Run with automatic restart on file changes
air
```

## 🧪 Development & Testing

This project follows **Test-Driven Development (TDD)** principles and Go best practices.

### Development Workflow
```bash
# Code quality checks
make format         # Format code with goimports
make lint          # Run golangci-lint
make test          # Run all tests
make dev-check     # Format + lint + test (pre-commit)

# Testing
go test             # Run all tests
go test -v          # Verbose output
go test -cover      # Coverage report
go test -bench=.    # Benchmarks

# Music system testing
make test-ytdlp     # Test yt-dlp service integration
make start-ytdlp    # Start yt-dlp service manually
```

### TDD Structure
```
internal/commands/
├── ping.go
├── ping_test.go          # Test-first development
├── user.go
├── user_test.go
└── testdata/             # Test fixtures
```

### Test Categories
- **Unit Tests**: Individual component testing
- **Integration Tests**: Cross-component interaction testing
- **Music System Tests**: Audio pipeline and service integration
- **Mock Testing**: Comprehensive mocking with testutils

## 🏗️ Architecture

### Project Structure (Pragmatic Go Layout)
```
pxnx-discord-bot-go/
├── main.go               # Application entrypoint
├── bot/                  # Core bot logic and session management
├── commands/             # Discord command handlers
├── music/                # Music system
│   ├── manager/         # Voice connection management
│   ├── player/          # DCA audio player
│   ├── queue/           # Thread-safe queue
│   ├── providers/       # Audio providers (YouTube)
│   └── types/           # Interfaces and types
├── services/             # External integrations
│   ├── ytdlp/           # yt-dlp service integration
│   └── weather.go       # OpenWeatherMap API
├── testutils/            # Test utilities and mocks
├── utils/                # Shared utility functions
├── scripts/              # Build and deployment scripts
├── go.mod               # Go module definition
└── go.sum               # Go module checksums
```

**Note**: This project uses a pragmatic structure suitable for its size. While the [Go Standard Project Layout](https://github.com/golang-standards/project-layout) with `cmd/` and `internal/` is recommended for larger projects, the current structure provides good organization without unnecessary complexity.

### Music System Architecture
```
Discord Command → Go Bot → yt-dlp Service → YouTube → Audio Stream → Discord Voice
                    ↓
              Service Manager → Python HTTP Server → yt-dlp Library
                    ↓
               DCA Audio Player → Voice Connection → Discord
```

## 🐳 Deployment

### Docker (Recommended)
```bash
# Using Docker Compose
cp .env.example .env     # Configure tokens
docker-compose up -d     # Start bot
docker-compose logs -f   # View logs

# Register commands (first time)
docker-compose exec pxnx-discord-bot ./pxnx-discord-bot --register-commands
```

### Manual Deployment
```bash
# Build binary
go build -o pxnx-discord-bot .

# Run with environment variables
export DISCORD_BOT_TOKEN=your_token
export OPENWEATHER_API_KEY=your_key
./pxnx-discord-bot
```

### Docker Features
- **Multi-architecture** (amd64/arm64)
- **Automatic builds** on GitHub Actions
- **Security scanning** with Trivy
- **Minimal Alpine runtime** (~25MB)
- **Complete dependencies** including Python/yt-dlp

## ⚠️ Current Status & Known Issues

### Working Components ✅
- All Discord commands functional
- Voice channel join/leave operations
- YouTube URL extraction and metadata
- Service management and health monitoring
- Docker deployment and CI/CD

### Active Issues 🔧
- **Music Playback**: EOF errors during streaming (infrastructure complete, investigating audio pipeline)
- **Root Cause**: YouTube stream URL expiration or DCA encoder compatibility
- **Investigation**: Focusing on stream reconnection and format selection

### Troubleshooting
```bash
# Check yt-dlp service health
curl http://localhost:8080/health

# View detailed logs
go run main.go --log-level debug

# Test components individually
go test ./music/player -v
go test ./services/ytdlp -v
```

## 🔧 Configuration

### Environment Variables
```env
# Required
DISCORD_BOT_TOKEN=your_discord_bot_token
OPENWEATHER_API_KEY=your_openweather_api_key

# Optional
LOG_LEVEL=info                    # debug, info, warn, error
YTDLP_SERVICE_PORT=8080          # yt-dlp service port
```

### Command Line Options
```bash
go run main.go --register-commands    # Register slash commands
go run main.go --log-level debug     # Enable debug logging
go run main.go --help               # Show all options
```

## 🤝 Contributing

1. **Follow TDD**: Write tests before implementation
2. **Use Go conventions**: Follow standard project layout
3. **Code quality**: Run `make dev-check` before commits
4. **Documentation**: Update README for significant changes

### Development Setup
```bash
# Install development tools
make install-tools

# Setup pre-commit hooks
make setup-hooks

# Verify setup
make dev-check
```

## 📄 License

This project is open source under the MIT License. See [LICENSE](LICENSE) for details.

---

**Note**: This bot is actively developed with a focus on code quality, testing, and maintainability. The music system infrastructure is complete with ongoing work to resolve streaming stability.