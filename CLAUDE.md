# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a simple Discord bot written in Go using the discordgo library. The bot is a single-file application (`main.go`) that implements slash command functionality with interactive features.

## Architecture

- **Single-file structure**: The entire bot logic is contained in `main.go`
- **Event-driven**: Uses Discord gateway events (ready, interactionCreate)
- **Dependencies**: Uses `github.com/bwmarrin/discordgo v0.29.0` for Discord API interaction
- **Intents**: Requires `GuildMessages` and `GuildEmojis` intents for message handling and emoji reactions

## Development Commands

### Setup
```bash
go mod tidy          # Download and organize dependencies
```

### Running the bot
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
- **`/weather` command**: Real weather data powered by OpenWeatherMap API
- Graceful shutdown on CTRL+C
- Automatic command registration on bot startup with global command reset (clears old/deleted commands)

## Code Structure

### Core Functions
- **`getCommands()`**: Returns list of application commands for registration
- **`registerCommands()`**: Handles command registration with Discord API, includes global command reset functionality
- **`getRandomPhrase(username)`**: Returns formatted phrase with username and random definition
- **`getUserAvatarURL()`**: Gets user avatar with fallback to default
- **`createPeepeeEmbed()`**: Creates embed response for peepee command
- **`getRandomEmoji()`**: Selects random emoji from server's emoji list with fallback

### Command Handlers
- **`handlePingCommand()`**: Handles `/ping` command (testable with interface)
- **`handlePeepeeCommand()`**: Handles `/peepee` command embed creation (testable)
- **`handlePeepeeCommandWithReaction()`**: Full `/peepee` command with emoji reactions
- **`handle8BallCommand()`**: Handles `/8ball` command with random responses
- **`handleCoinFlipCommand()`**: Handles `/coinflip` command
- **`handleServerCommand()`**: Handles `/server` command for guild information
- **`handleUserCommand()`**: Handles `/user` command for user profiles
- **`handleWeatherCommand()`**: Handles `/weather` command with OpenWeatherMap API
- **`getWeatherData()`**: Fetches real weather data from OpenWeatherMap
- **`getWeatherIcon()`**: Maps weather conditions to appropriate emojis
- **`interactionCreate()`**: Main Discord event handler routing commands

### Testing
- **`main_test.go`**: Comprehensive test suite with unit tests and benchmarks
- **`SessionInterface`**: Interface for mocking Discord session in tests
- **`MockSession`**: Test mock implementing SessionInterface
- Tests cover: command registration, username phrase formatting, embed creation, avatar handling, and command handlers
- **`peepeeDefinitions`**: 20 trendy size definitions for variety and humor