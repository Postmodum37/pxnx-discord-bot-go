# CLAUDE.md - AI Development Guide

This file provides comprehensive guidance for AI assistants (Claude Code) when working with this Discord bot project. The project follows **Go best practices**, **Test-Driven Development (TDD)**, and **clean architecture principles**.

## Project Overview

**PXNX Discord Bot** is a production-ready Discord bot written in Go, featuring:
- **Modern Go architecture** following standard project layout
- **Test-Driven Development** with comprehensive test coverage
- **Service-oriented design** with external service integrations
- **Production deployment** via Docker with CI/CD pipeline
- **Music system** with YouTube integration via yt-dlp service

## Go Best Practices & Project Structure

### Current Project Structure

This project uses a pragmatic Go layout suitable for its size and complexity:

```
pxnx-discord-bot-go/
├── main.go               # Application entrypoint
├── bot/                  # Core bot logic and session management
├── commands/             # Discord command handlers
├── music/                # Music system (manager, player, queue, providers)
│   ├── manager/         # Voice connection management
│   ├── player/          # DCA audio player
│   ├── queue/           # Thread-safe queue
│   ├── providers/       # Audio providers (YouTube)
│   └── types/           # Interfaces and types
├── services/             # External service integrations
│   ├── ytdlp/           # yt-dlp service integration
│   └── weather.go       # OpenWeatherMap API
├── testutils/            # Test utilities and mocks
├── utils/                # Shared utility functions
├── scripts/              # Build and deployment scripts
├── go.mod               # Go module definition
└── go.sum               # Go module checksums
```

**Structure Notes**:
- **Pragmatic over dogmatic**: While the [Go Standard Project Layout](https://github.com/golang-standards/project-layout) with `cmd/` and `internal/` is ideal for larger projects, this structure avoids unnecessary complexity
- **Clear organization**: Each package has a focused responsibility
- **Easy navigation**: Flat structure makes imports and navigation straightforward
- **Test co-location**: Tests live alongside the code they test

### Key Architecture Principles

#### 1. **Dependency Injection & Interfaces**
```go
// Good: Define interfaces in consuming packages
type AudioPlayer interface {
    Play(ctx context.Context, source AudioSource) error
    Stop() error
    IsPlaying() bool
}

// Implementation in separate package
func NewDCAAudioPlayer(guildID string, conn *discordgo.VoiceConnection) AudioPlayer {
    return &DCAAudioPlayer{...}
}
```

#### 2. **Context Usage**
```go
// Always pass context as first parameter
func (m *Manager) Play(ctx context.Context, guildID string, source AudioSource) error {
    // Use context for cancellation and timeouts
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        // Continue with operation
    }
}
```

#### 3. **Error Handling**
```go
// Wrap errors with context
func (p *Provider) GetAudioSource(ctx context.Context, query string) (*AudioSource, error) {
    data, err := p.fetchData(ctx, query)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch audio source for %q: %w", query, err)
    }
    return p.parseData(data)
}
```

#### 4. **Package Organization**
- **`internal/`**: Private application code, cannot be imported by external packages
- **`pkg/`**: Public library code that can be reused
- **Single responsibility**: Each package has a clear, focused purpose
- **Minimal interfaces**: Keep interfaces small and focused

## Test-Driven Development (TDD)

### TDD Workflow

1. **Write Test First** (Red)
2. **Implement Minimum Code** (Green)
3. **Refactor** (Refactor)

### Test Structure

#### Unit Tests
```go
// commands/ping_test.go
func TestHandlePingCommand(t *testing.T) {
    // Arrange
    mockSession := &testutils.MockSession{}
    interaction := testutils.CreateTestInteraction("ping", nil)

    // Act
    err := HandlePingCommand(mockSession, interaction)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Pong!", mockSession.LastResponse.Content)
}
```

#### Integration Tests
```go
// test/integration/music_test.go
func TestMusicSystemIntegration(t *testing.T) {
    // Test complete workflow: join → play → leave
    manager := setupTestMusicManager(t)
    defer manager.Cleanup(context.Background())

    // Test actual voice connection and playback
    err := manager.JoinChannel(ctx, testGuildID, testChannelID)
    require.NoError(t, err)

    source := testutils.CreateTestAudioSource()
    err = manager.Play(ctx, testGuildID, source)
    assert.NoError(t, err)
}
```

#### Test Organization
- **File naming**: `*_test.go` in same package as code under test
- **Test data**: Use `testdata/` directories for fixtures
- **Mocks**: Centralized in `pkg/testutils/` for reuse
- **Table tests**: Use for multiple test cases

```go
func TestUserCommand(t *testing.T) {
    tests := []struct {
        name           string
        targetUserID   string
        expectedResult string
        expectError    bool
    }{
        {"valid user", "123456789", "User: TestUser#1234", false},
        {"invalid user", "invalid", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Testing Guidelines

#### What to Test
- **All public functions**: Especially exported functions and methods
- **Error conditions**: Test both happy path and error scenarios
- **Business logic**: Core functionality and edge cases
- **Integration points**: External service interactions

#### What NOT to Test
- **Third-party libraries**: Don't test discordgo, yt-dlp, etc.
- **Simple getters/setters**: Unless they contain logic
- **Private implementation details**: Test behavior, not implementation

#### Mock Strategy
```go
// pkg/testutils/mocks.go
type MockSession struct {
    LastResponse *discordgo.InteractionResponse
    ShouldError  bool
}

func (m *MockSession) InteractionRespond(
    interaction *discordgo.Interaction,
    resp *discordgo.InteractionResponse,
) error {
    if m.ShouldError {
        return errors.New("mock error")
    }
    m.LastResponse = resp
    return nil
}
```

## Development Commands & Workflow

### Essential Commands
```bash
# Development workflow
make format         # Format code with goimports
make lint          # Run golangci-lint with comprehensive checks
make test          # Run all tests with coverage
make dev-check     # Run format + lint + test (pre-commit)

# Testing
go test ./...                    # Run all tests
go test -v ./internal/commands   # Verbose output for specific package
go test -race ./...              # Test for race conditions
go test -cover ./...             # Coverage report
go test -bench=. ./...           # Run benchmarks

# Build and run
go build .                       # Build binary
go run . --help                  # Run with options
air                              # Hot reload during development
```

### Code Quality Tools

#### Required Tools
- **goimports**: Code formatting and import organization
- **golangci-lint**: Comprehensive linting with multiple checkers
- **air**: Hot reload for development

#### Linting Configuration
The project uses `.golangci.yml` with strict settings:
- **govet**: Suspicious constructs
- **errcheck**: Unchecked errors
- **staticcheck**: Advanced static analysis
- **gocritic**: Code review automation
- **gosec**: Security issues
- **misspell**: Spelling errors

## Service Architecture Patterns

### Music System Example

The music system demonstrates proper Go architecture patterns:

#### 1. **Interface Segregation**
```go
// Focused interfaces for specific responsibilities
type AudioPlayer interface {
    Play(ctx context.Context, source AudioSource) error
    Stop() error
    Pause() error
    Resume() error
}

type Queue interface {
    Add(source AudioSource)
    Next() (*AudioSource, bool)
    Clear()
}
```

#### 2. **Dependency Injection**
```go
// Manager depends on interfaces, not concrete types
type Manager struct {
    session   SessionInterface      // Testable via interface
    players   map[string]AudioPlayer
    providers map[string]AudioProvider
}

func NewManager(session SessionInterface) *Manager {
    return &Manager{
        session:   session,
        players:   make(map[string]AudioPlayer),
        providers: make(map[string]AudioProvider),
    }
}
```

#### 3. **Service Integration**
```go
// External service with health monitoring
type YouTubeProvider struct {
    client         *ytdlp.Client
    serviceManager *ytdlp.ServiceManager
}

func (p *YouTubeProvider) GetAudioSource(ctx context.Context, query string) (*AudioSource, error) {
    // Health check before operation
    if !p.serviceManager.IsRunning() {
        if err := p.serviceManager.Start(ctx); err != nil {
            return nil, fmt.Errorf("service unavailable: %w", err)
        }
    }

    return p.client.Search(ctx, query)
}
```

## Current System Status

### Working Components ✅
- **Core Discord Integration**: Bot connection, command handling, event processing
- **Command System**: All slash commands functional with proper error handling
- **Music Infrastructure**: Voice connections, yt-dlp service, DCA player setup
- **Service Management**: Health monitoring, auto-restart, error recovery
- **Testing Framework**: Comprehensive test coverage with mocks and integration tests
- **CI/CD Pipeline**: Docker builds, security scanning, automated deployment

### Active Development Areas 🔧
- **Audio Streaming**: Investigating EOF errors in DCA encoder pipeline
- **Stream Reliability**: YouTube URL expiration and reconnection logic
- **Performance Optimization**: Memory usage and connection pooling

### Investigation Priorities
1. **Stream URL Lifetime**: Test YouTube URL validity duration
2. **DCA Compatibility**: Alternative audio formats and encoder settings
3. **Error Recovery**: Enhanced reconnection and retry mechanisms

## AI Assistant Guidelines

### When Working on This Project

#### 1. **Always Follow TDD**
- Write tests before implementation
- Use red-green-refactor cycle
- Maintain high test coverage

#### 2. **Respect Go Conventions**
- Use standard project layout
- Follow naming conventions (PascalCase for exported, camelCase for private)
- Implement proper error handling with wrapped errors
- Use context.Context for cancellation and timeouts

#### 3. **Code Quality Checks**
- Run `make dev-check` before suggesting code
- Ensure linting passes
- Verify test coverage doesn't decrease

#### 4. **Architecture Decisions**
- Prefer interfaces over concrete types for dependencies
- Keep packages focused and cohesive
- Use dependency injection for testability
- Implement proper error handling and logging

#### 5. **Documentation**
- Update README.md for user-facing changes
- Update this CLAUDE.md for development workflow changes
- Include code comments for complex business logic
- Document public APIs with godoc-style comments

### Common Tasks & Patterns

#### Adding New Commands
```go
// 1. Define test first
func TestNewCommand(t *testing.T) {
    mockSession := &testutils.MockSession{}
    interaction := testutils.CreateTestInteraction("newcommand", nil)

    err := HandleNewCommand(mockSession, interaction)

    assert.NoError(t, err)
    // Assert expected behavior
}

// 2. Implement handler
func HandleNewCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
    // Implementation
}

// 3. Register in bot/commands.go
// 4. Add to interaction handler in bot/handlers.go
```

#### Working with Music System
```go
// Always check service health first
if !provider.IsServiceRunning() {
    if err := provider.Start(ctx); err != nil {
        return fmt.Errorf("service startup failed: %w", err)
    }
}

// Use proper context handling
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Handle errors gracefully with user-friendly messages
if err != nil {
    return &types.MusicError{
        Type:    "service_error",
        Message: "Unable to process request. Please try again.",
        Err:     err,
    }
}
```

### Testing Strategy

#### For New Features
1. **Unit tests** for core logic
2. **Integration tests** for service interactions
3. **Mock tests** for external dependencies
4. **Error scenario tests** for robust error handling

#### For Bug Fixes
1. **Reproduce with test** that fails
2. **Fix implementation** to make test pass
3. **Add regression tests** to prevent future issues

## Deployment & Production

### Docker Best Practices
- **Multi-stage builds** for minimal image size
- **Non-root user** for security
- **Health checks** for service monitoring
- **Multi-architecture** support (amd64/arm64)

### Environment Configuration
- **Secure secrets** handling via environment variables
- **Configuration validation** at startup
- **Graceful shutdown** with proper cleanup
- **Structured logging** for production monitoring

---

**Remember**: This project emphasizes **quality over speed**. Always prioritize proper testing, clean architecture, and maintainable code over quick implementations.

- Always double-check package dependencies if they are legit, supported and maintained. Don't over use it, but use it where it seems necessary.