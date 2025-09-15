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

- **`/ping` command**: Simple ping-pong response
- **`/peepee` command**: Interactive inspection command with random funny definitions and emoji reactions
- **`/8ball` command**: Magic 8-ball with 20 classic responses
- **`/coinflip` command**: Random heads/tails coin flip
- **`/server` command**: Display server information (member count, creation date, etc.)
- **`/user` command**: Show user profile information with optional target parameter
- **`/weather` command**: Real weather data powered by OpenWeatherMap API with support for current weather, 1-day, and 5-day forecasts
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
- **`user.go`**: User profile display with avatar and account information
- **`server.go`**: Server/guild information display
- **`weather.go`**: Weather command integrating with OpenWeatherMap API
- **`interfaces.go`**: SessionInterface definition for testability

#### `services/` Package
- **`weather.go`**: OpenWeatherMap API integration and data structures

#### `utils/` Package
- **`avatar.go`**: User avatar URL generation with fallbacks
- **`random.go`**: Random phrase generation and emoji selection utilities

#### `testutils/` Package
- **`mocks.go`**: MockSession implementing SessionInterface for testing
- **`fixtures.go`**: Test data factories (users, guilds, interactions, emojis)

### Testing Structure
- **Package-specific tests**: Each package has comprehensive test coverage
- **Mock interfaces**: SessionInterface allows for isolated unit testing
- **Test utilities**: Centralized mock creation and fixture generation
- **Coverage**: 33.1% overall coverage with room for improvement in complex Discord interactions
- **Test organization**: Clear separation between unit tests, integration tests, and benchmarks