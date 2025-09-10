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
# Set environment variable (copy from .env.example)
export DISCORD_BOT_TOKEN=your_bot_token_here
go run main.go       # Run the bot
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

The bot requires a `DISCORD_BOT_TOKEN` environment variable. Use `.env.example` as a template for creating your environment configuration.

## Bot Features

- **`/ping` command**: Simple ping-pong response
- **`/peepee` command**: Interactive inspection command with:
  - Random funny definitions in format: "{username} {definition} peepee!"
  - 20 trendy and humorous size definitions
  - Blue embed response with user's avatar as thumbnail
  - Random server emoji reaction
- Graceful shutdown on CTRL+C
- Automatic command registration on bot startup

## Code Structure

### Core Functions
- **`getCommands()`**: Returns list of application commands for registration
- **`registerCommands()`**: Handles command registration with Discord API
- **`getRandomPhrase(username)`**: Returns formatted phrase with username and random definition
- **`getUserAvatarURL()`**: Gets user avatar with fallback to default
- **`createPeepeeEmbed()`**: Creates embed response for peepee command
- **`getRandomEmoji()`**: Selects random emoji from server's emoji list with fallback

### Command Handlers
- **`handlePingCommand()`**: Handles `/ping` command (testable with interface)
- **`handlePeepeeCommand()`**: Handles `/peepee` command embed creation (testable)
- **`handlePeepeeCommandWithReaction()`**: Full `/peepee` command with emoji reactions
- **`interactionCreate()`**: Main Discord event handler routing commands

### Testing
- **`main_test.go`**: Comprehensive test suite with unit tests and benchmarks
- **`SessionInterface`**: Interface for mocking Discord session in tests
- **`MockSession`**: Test mock implementing SessionInterface
- Tests cover: command registration, username phrase formatting, embed creation, avatar handling, and command handlers
- **`peepeeDefinitions`**: 20 trendy size definitions for variety and humor