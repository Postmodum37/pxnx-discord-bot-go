# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a simple Discord bot written in Go using the discordgo library. The bot has been refactored from a single-file application into a modular structure with organized packages for better maintainability and testing.

## Architecture

- **Modular structure**: Organized into packages for better maintainability:
  - `bot/` - Core bot initialization, Discord session management, and command routing
  - `commands/` - Individual command handlers (ping, user, weather, peepee, 8ball, etc.)
  - `services/` - External service integrations (OpenWeatherMap API)
  - `testutils/` - Test utilities, mocks, and fixtures for comprehensive testing
  - `utils/` - Shared utility functions (avatar handling, random selections)
- **Event-driven**: Uses Discord gateway events (ready, interactionCreate)
- **Dependencies**: Uses `github.com/bwmarrin/discordgo v0.29.0` for Discord API interaction
- **Intents**: Requires `GuildMessages` and `GuildEmojis` intents for message handling and emoji reactions

## Development Commands

### Setup
```bash
go mod tidy          # Download and organize dependencies
```

### Running the bot

#### Local Development
```bash
# Copy .env.example to .env and add your tokens
cp .env.example .env
# Edit .env with your actual tokens
# Then run the bot (it will automatically load .env)
go run main.go
```

#### Hot Reload Development
For development with automatic restart on file changes:
```bash
# Install air (one-time setup)
go install github.com/air-verse/air@latest

# Run with hot reload (automatically restarts on .go file changes)
air

# The configuration is in .air.toml and excludes test files and tmp directory
```

Alternatively, you can set environment variables manually:
```bash
export DISCORD_BOT_TOKEN=your_bot_token_here
export OPENWEATHER_API_KEY=your_openweathermap_api_key_here
go run main.go
```

#### Docker Deployment (Production)
The application is containerized and available via GitHub Container Registry:

```bash
# 1. Create environment file
cp .env.example .env
# Edit .env with your actual tokens

# 2. Pull and run with Docker Compose
docker-compose pull
docker-compose up -d

# 3. View logs
docker-compose logs -f

# 4. Stop the bot
docker-compose down
```

**Docker Image**: `ghcr.io/postmodum37/pxnx-discord-bot-go:latest`
- Multi-architecture support (amd64/arm64)
- Automatic builds on every commit
- Security scanned with Trivy
- Minimal Alpine-based runtime (~15MB)

#### Command Registration
The bot doesn't register slash commands automatically. You need to register them once:
```bash
# Register slash commands with Discord (only needed once or when commands change)
go run main.go --register-commands

# Normal bot startup (fast, no command registration)
go run main.go
```

#### Command line options
```bash
go run main.go --register-commands    # Register bot commands with Discord (cleans up existing commands first)
go run main.go --help               # Show all available command line options
```

### Building
```bash
go build            # Build executable
./pxnx-discord-bot   # Run the built executable
```

### Testing
```bash
go test             # Run all tests
go test -v          # Run tests with verbose output
go test -bench=.    # Run tests with benchmarks
go test -cover      # Run tests with coverage report
```

### Code Quality (Linting & Formatting)
The project uses industry-standard Go tools for code quality:

```bash
# Using Make commands (recommended)
make help           # Show all available commands
make format         # Format code with goimports
make lint           # Run golangci-lint
make check          # Format + lint
make dev-check      # Format + lint + test (use before committing)

# Direct tool usage
goimports -w *.go                    # Format code and organize imports
golangci-lint run --timeout=2m      # Run comprehensive linter
```

## Environment Configuration

The bot requires the following environment variables:
- `DISCORD_BOT_TOKEN`: Your Discord bot token
- `OPENWEATHER_API_KEY`: Your OpenWeatherMap API key (get one free at https://openweathermap.org/api)

The bot automatically loads environment variables from a `.env` file if present. Copy `.env.example` to `.env` and add your actual tokens, or set the environment variables manually.

## Bot Features

### Core Commands
- **`/ping` command**: Simple ping-pong response
- **`/peepee` command**: Interactive inspection command with random funny definitions and emoji reactions
- **`/8ball` command**: Magic 8-ball with 20 classic responses
- **`/coinflip` command**: Random heads/tails coin flip
- **`/roll` command**: Roll a dice with customizable maximum value (default: 1-100, supports 1-1000000)
- **`/server` command**: Display server information (member count, creation date, etc.)
- **`/user` command**: Show user profile information with optional target parameter
- **`/weather` command**: Real weather data powered by OpenWeatherMap API with support for current weather, 1-day, and 5-day forecasts

### Music System (Infrastructure Complete - Audio Playback Issues)
⚠️ **Current Status**: All infrastructure components working, but actual audio playback has intermittent failures with EOF errors

- **`/join` command**: ✅ Connect bot to your voice channel with validation and error handling
- **`/leave` command**: ✅ Disconnect from voice channel and clean up all resources
- **`/play` command**: ⚠️ **YouTube integration with known issues**
  - ✅ **Real search functionality**: No API key required, uses yt-dlp's search capabilities
  - ✅ **Direct URL support**: Complete URL parsing (youtube.com, youtu.be, mobile, shorts)
  - ✅ **Rich embeds**: Metadata with thumbnails, duration, uploader, and view counts
  - ⚠️ **Audio playback**: Stream URL extraction works but playback fails with EOF errors

#### **🏗️ yt-dlp Service Architecture**
- **Python HTTP Service**: Separate async service wrapping yt-dlp functionality
- **Go HTTP Client**: Full-featured client with health checking and error recovery
- **Service Manager**: Process lifecycle management with auto-start/stop/restart
- **Circuit Breaker**: Resilience patterns with automatic retry logic and failure recovery
- **Caching System**: TTL-based caching for improved performance and reduced load

#### **🎵 Audio System**
- **DCA Audio Player**: Production-ready audio player with volume control and playback management
- **Queue System**: Thread-safe FIFO operations with add, remove, shuffle, and position-based management
- **Voice Management**: Thread-safe connection handling with auto-disconnect timer
- **Format Intelligence**: Automatic best audio format selection (opus/webm preferred)

#### **🔧 Production Features**
- **Automatic Service Management**: yt-dlp service starts/stops with bot automatically
- **Health Monitoring**: Continuous service health checking and status reporting
- **Comprehensive Error Handling**: User-friendly error messages and proper resource cleanup
- **Thread-Safe Operations**: All music operations are concurrent-safe
- **Resource Management**: Proper cleanup and memory management

### System Features
- Graceful shutdown on CTRL+C
- Manual command registration with `--register-commands` flag (includes cleanup of old commands)
- Fast startup without automatic command registration

## Code Structure

### Package Organization

#### `bot/` Package
- **`bot.go`**: Core bot structure, Discord session management, and initialization
- **`commands.go`**: Command definitions and registration logic
- **`handlers.go`**: Main interaction routing and command dispatch

#### `commands/` Package
- **`ping.go`**: Simple ping/pong response handler
- **`peepee.go`**: Interactive inspection command with random phrases and emoji reactions
- **`eightball.go`**: Magic 8-ball with predefined responses
- **`coinflip.go`**: Coin flip randomization
- **`roll.go`**: Dice rolling with customizable maximum values
- **`user.go`**: User profile display with avatar and account information
- **`server.go`**: Server/guild information display
- **`weather.go`**: Weather command integrating with OpenWeatherMap API
- **`music_join.go`**: Voice channel join/leave command handlers
- **`music_play.go`**: Music playback command handler with full YouTube integration
- **`interfaces.go`**: SessionInterface definition for testability

#### `services/` Package
- **`weather.go`**: OpenWeatherMap API integration and data structures
- **`ytdlp/`**: Complete yt-dlp service integration:
  - **`server.py`**: Python HTTP service wrapping yt-dlp functionality
  - **`client.go`**: Go HTTP client with health checking and error recovery
  - **`manager.go`**: Service process lifecycle management
  - **`types.go`**: Type definitions and data structures
  - **`resilience.go`**: Circuit breaker and retry patterns
  - **`requirements.txt`**: Python dependencies

#### `music/` Package
- **`types/interfaces.go`**: Core music system interfaces and type definitions (AudioPlayer, Queue, AudioProvider, MusicManager)
- **`manager/manager.go`**: Music manager implementation with voice connection handling and provider management
- **`manager/session_wrapper.go`**: Discord session wrapper for voice functionality
- **`player/player.go`**: DCA audio player implementation with volume control and playback management
- **`queue/queue.go`**: Thread-safe FIFO queue implementation with shuffle and position-based operations
- **`providers/youtube.go`**: Legacy YouTube provider (basic functionality)
- **`providers/youtube_ytdlp.go`**: **Production YouTube provider** with full yt-dlp integration
- **`manager/ytdlp_integration.go`**: Integration helpers for yt-dlp service management

#### `utils/` Package
- **`avatar.go`**: User avatar URL generation with fallbacks
- **`random.go`**: Random phrase generation and emoji selection utilities

#### `testutils/` Package
- **`mocks.go`**: MockSession implementing SessionInterface for testing
- **`music_mocks.go`**: Music system mocks for testing voice and audio functionality
- **`fixtures.go`**: Test data factories (users, guilds, interactions, emojis)

### Testing Structure
- **Package-specific tests**: Each package has comprehensive test coverage
- **Mock interfaces**: SessionInterface allows for isolated unit testing
- **Music system testing**: Comprehensive mocks for voice connections, audio players, and queue management
- **Test utilities**: Centralized mock creation and fixture generation
- **Coverage**: Enhanced test coverage with music system additions
- **Test organization**: Clear separation between unit tests, integration tests, and benchmarks

## Future Feature Roadmap

### 🎵 Music System (Infrastructure Complete - Debugging Phase)
**✅ Completed Infrastructure:**
- Voice channel join/leave with validation and error handling
- **Full YouTube integration** with URL parsing, metadata extraction, and search functionality
- **Complete `/play` command** with rich embeds and queue management
- Thread-safe music manager with comprehensive testing and provider management
- **DCA-ready audio player** with volume control, pause/resume, and track management
- **Queue system** with FIFO operations, shuffle, and position-based management
- Session wrapper for voice functionality with fixed auto-disconnect timer
- **yt-dlp service integration** with HTTP client and service management

**🚧 Current Priority - Audio Streaming Fixes:**
- **EOF Error Resolution**: Investigate YouTube stream URL lifetime and connection issues
- **Stream Reconnection**: Implement robust reconnection for interrupted YouTube streams
- **Alternative Audio Sources**: Test with non-YouTube sources to isolate issues
- **Error Pattern Analysis**: Analyze timing and frequency of streaming failures

**🎯 Future Enhancements:**
- **Enhanced Commands**: `/queue`, `/skip`, `/pause`, `/resume`, `/stop`, `/volume`, `/now-playing`
- **Advanced features**: Search result selection, playlist support, multiple provider support
- **Quality Options**: Bitrate selection, audio format preferences

### 🎮 RPG Game System
- **Character creation** with classes and stats
- **Combat system** with monsters, experience, and loot
- **Inventory management** with gear and upgrades

### 📈 Financial Data
- **Stock market tracking** with price alerts and trends
- **Cryptocurrency portfolio** management and price monitoring

### 🎯 Gaming Integrations
- **World of Warcraft** API for character/server data
- **Archon.gg tier lists** for competitive game rankings

### 🤖 AI Integration
- **AI chatbot** for conversational interactions and natural language responses

### 🛠️ Utility Commands
- **Reminder system** for scheduled notifications
- **Poll creation** with voting and results
- **URL shortener** for Discord links
- **Meme generator** with templates and custom text
- **Trivia game** with categories and scoring

These features represent potential expansions to the bot's functionality. The music and RPG systems would be the most complex implementations, requiring additional dependencies and persistent data storage.