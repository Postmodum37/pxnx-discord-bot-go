# PXNX Discord Bot

A simple Discord bot written in Go using the discordgo library. This bot provides various slash commands with interactive features and real-time data integration.

## Features

- **`/ping`** - Simple ping-pong response
- **`/peepee`** - Interactive inspection command with random funny definitions and emoji reactions
- **`/8ball`** - Magic 8-ball with 20 classic responses
- **`/coinflip`** - Random heads/tails coin flip
- **`/server`** - Display server information (member count, creation date, etc.)
- **`/user`** - Show user profile information with optional target parameter
- **`/weather`** - Real weather data powered by OpenWeatherMap API
- Graceful shutdown on CTRL+C
- Manual command registration with `--register-commands` flag (includes cleanup of old commands)
- Fast startup without automatic command registration

## Quick Setup

1. **Clone and install dependencies**:
   ```bash
   git clone <repository-url>
   cd pxnx-discord-bot-go
   go mod tidy
   ```

2. **Set up environment variables**:
   ```bash
   cp .env.example .env
   # Edit .env with your actual tokens
   ```

3. **Required environment variables**:
   - `DISCORD_BOT_TOKEN`: Your Discord bot token
   - `OPENWEATHER_API_KEY`: Your OpenWeatherMap API key ([get one free](https://openweathermap.org/api))

4. **Register slash commands** (only needed once):
   ```bash
   go run main.go --register-commands
   ```

5. **Run the bot**:
   ```bash
   go run main.go
   ```

## Usage

### Command Registration
The bot doesn't register slash commands automatically for faster startup times. You need to register them once:

```bash
# Register slash commands with Discord (only needed once or when commands change)
go run main.go --register-commands

# Normal bot startup (fast, no command registration)
go run main.go
```

### Command Line Options
```bash
go run main.go --register-commands    # Register bot commands with Discord (cleans up existing commands first)
go run main.go --help               # Show all available command line options
```

### Building
```bash
go build            # Build executable
./pxnx-discord-bot   # Run the built executable
```

## Development

### Code Quality Tools
This project uses industry-standard Go linting and formatting tools:

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

**Required tools** (installed automatically via go install):
- `goimports` - Code formatting and import organization
- `golangci-lint` - Comprehensive Go linter with multiple checks

### Testing
```bash
go test             # Run all tests
go test -v          # Run tests with verbose output
go test -bench=.    # Run tests with benchmarks
go test -cover      # Run tests with coverage report
```

### Architecture
- **Modular structure**: Organized into packages for better maintainability
  - `bot/` - Core bot initialization and Discord session management
  - `commands/` - Individual command handlers (ping, user, weather, etc.)
  - `services/` - External service integrations (OpenWeatherMap API)
  - `testutils/` - Test utilities, mocks, and fixtures
- **Event-driven**: Uses Discord gateway events (ready, interactionCreate)
- **Dependencies**: Uses `github.com/bwmarrin/discordgo v0.29.0` for Discord API interaction
- **Intents**: Requires `GuildMessages` and `GuildEmojis` intents for message handling and emoji reactions
- **Test coverage**: Comprehensive test suite with unit tests and benchmarks (33.1% coverage)

## Environment Setup

The bot automatically loads environment variables from a `.env` file if present. Alternatively, you can set environment variables manually:

```bash
export DISCORD_BOT_TOKEN=your_bot_token_here
export OPENWEATHER_API_KEY=your_openweathermap_api_key_here
go run main.go
```

## License

This project is open source. Feel free to contribute or modify as needed.