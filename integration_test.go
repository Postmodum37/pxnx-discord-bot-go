// integration_test.go
// These tests verify that all packages work together correctly

package main

import (
	"os"
	"testing"

	"github.com/bwmarrin/discordgo"

	"pxnx-discord-bot/bot"
	"pxnx-discord-bot/commands"
	"pxnx-discord-bot/services"
	"pxnx-discord-bot/testutils"
)

func TestBotCommandIntegration(t *testing.T) {
	// Test that bot package properly integrates with commands
	commandList := bot.GetCommands()

	if len(commandList) == 0 {
		t.Fatal("Expected commands to be available")
	}

	// Test each command can be created successfully
	mockSession := &testutils.MockSession{}

	tests := []struct {
		name    string
		command string
		setup   func() *discordgo.InteractionCreate
	}{
		{
			name:    "ping command integration",
			command: "ping",
			setup: func() *discordgo.InteractionCreate {
				return testutils.CreateTestInteraction("ping", nil)
			},
		},
		{
			name:    "peepee command integration",
			command: "peepee",
			setup: func() *discordgo.InteractionCreate {
				interaction := testutils.CreateTestInteraction("peepee", nil)
				user := testutils.CreateTestUser("123", "testuser", "avatar")
				interaction.Member = testutils.CreateTestMember(user)
				return interaction
			},
		},
		{
			name:    "8ball command integration",
			command: "8ball",
			setup: func() *discordgo.InteractionCreate {
				options := []*discordgo.ApplicationCommandInteractionDataOption{
					testutils.CreateStringOption("question", "Test question?"),
				}
				return testutils.CreateTestInteraction("8ball", options)
			},
		},
		{
			name:    "coinflip command integration",
			command: "coinflip",
			setup: func() *discordgo.InteractionCreate {
				return testutils.CreateTestInteraction("coinflip", nil)
			},
		},
		{
			name:    "user command integration",
			command: "user",
			setup: func() *discordgo.InteractionCreate {
				interaction := testutils.CreateTestInteraction("user", nil)
				user := testutils.CreateTestUser("456", "testuser2", "avatar2")
				interaction.Member = testutils.CreateTestMember(user)
				return interaction
			},
		},
		{
			name:    "weather command integration",
			command: "weather",
			setup: func() *discordgo.InteractionCreate {
				options := []*discordgo.ApplicationCommandInteractionDataOption{
					testutils.CreateStringOption("city", "London"),
				}
				return testutils.CreateTestInteraction("weather", options)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession.Reset()
			interaction := tt.setup()

			var err error

			// This mimics the bot's interaction routing
			switch tt.command {
			case "ping":
				err = commands.HandlePingCommand(mockSession, interaction)
			case "peepee":
				err = commands.HandlePeepeeCommand(mockSession, interaction)
			case "8ball":
				err = commands.Handle8BallCommand(mockSession, interaction)
			case "coinflip":
				err = commands.HandleCoinFlipCommand(mockSession, interaction)
			case "user":
				err = commands.HandleUserCommand(mockSession, interaction)
			case "weather":
				err = commands.HandleWeatherCommand(mockSession, interaction)
			default:
				t.Errorf("Unknown command: %s", tt.command)
				return
			}

			if err != nil && tt.command != "weather" {
				t.Errorf("Command %s failed: %v", tt.command, err)
			}

			// Weather command might fail due to missing API key, which is expected
			if tt.command == "weather" && err != nil {
				t.Logf("Weather command failed as expected (no API key): %v", err)
			}

			if !mockSession.RespondCalled {
				t.Errorf("Command %s did not call InteractionRespond", tt.command)
			}
		})
	}
}

func TestWeatherServiceIntegration(t *testing.T) {
	// Test weather service integration with commands

	// Save original API key
	originalKey := os.Getenv("OPENWEATHER_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENWEATHER_API_KEY", originalKey)
		} else {
			os.Unsetenv("OPENWEATHER_API_KEY")
		}
	}()

	t.Run("no API key scenario", func(t *testing.T) {
		os.Unsetenv("OPENWEATHER_API_KEY")

		_, err := services.GetWeatherData("London")
		if err == nil {
			t.Error("Expected error when API key is missing")
		}

		// Test that command handles this gracefully
		mockSession := &testutils.MockSession{}
		options := []*discordgo.ApplicationCommandInteractionDataOption{
			testutils.CreateStringOption("city", "London"),
		}
		interaction := testutils.CreateTestInteraction("weather", options)

		err = commands.HandleWeatherCommand(mockSession, interaction)
		if err != nil {
			t.Errorf("Weather command should handle missing API key gracefully, got error: %v", err)
		}

		if !mockSession.RespondCalled {
			t.Error("Weather command should still respond when API key is missing")
		}
	})

	t.Run("with API key scenario", func(t *testing.T) {
		os.Setenv("OPENWEATHER_API_KEY", "test_key")

		// This will likely fail due to network/invalid key, but should not panic
		_, err := services.GetWeatherData("London")
		if err != nil {
			t.Logf("Expected network error with test API key: %v", err)
		}
	})
}

func TestPackageInterfaces(t *testing.T) {
	// Test that all packages implement expected interfaces correctly

	t.Run("command handlers implement SessionInterface", func(t *testing.T) {
		mockSession := &testutils.MockSession{}

		// Test that our mock satisfies the SessionInterface
		var _ commands.SessionInterface = mockSession

		// Test that command handlers accept the interface
		interaction := testutils.CreateTestInteraction("ping", nil)
		err := commands.HandlePingCommand(mockSession, interaction)

		if err != nil {
			t.Errorf("Command handler failed with SessionInterface: %v", err)
		}
	})
}

func TestBotLifecycle(t *testing.T) {
	// Test bot creation and setup lifecycle

	t.Run("bot creation and setup", func(t *testing.T) {
		// Test bot creation
		botInstance, err := bot.New("test.token.here")
		if err != nil {
			t.Fatalf("Failed to create bot: %v", err)
		}

		if botInstance == nil {
			t.Fatal("Expected bot instance but got nil")
		}

		if botInstance.Session == nil {
			t.Fatal("Expected session to be initialized")
		}

		// Test bot setup - handlers are unexported, so we can't test them directly
		botInstance.Setup()

		// Test command registration flag
		bot.SetShouldRegisterCommands(true)
		bot.SetShouldRegisterCommands(false) // Just test it doesn't panic
	})
}

func TestEndToEndCommandFlow(t *testing.T) {
	// Test complete command flow from interaction to response

	tests := []struct {
		name        string
		command     string
		expectText  bool
		expectEmbed bool
	}{
		{
			name:        "ping command flow",
			command:     "ping",
			expectText:  true,
			expectEmbed: false,
		},
		{
			name:        "peepee command flow",
			command:     "peepee",
			expectText:  false,
			expectEmbed: true,
		},
		{
			name:        "8ball command flow",
			command:     "8ball",
			expectText:  false,
			expectEmbed: true,
		},
		{
			name:        "coinflip command flow",
			command:     "coinflip",
			expectText:  false,
			expectEmbed: true,
		},
		{
			name:        "user command flow",
			command:     "user",
			expectText:  false,
			expectEmbed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &testutils.MockSession{}

			// Create appropriate interaction
			var interaction *discordgo.InteractionCreate
			switch tt.command {
			case "ping":
				interaction = testutils.CreateTestInteraction("ping", nil)
			case "peepee", "user":
				interaction = testutils.CreateTestInteraction(tt.command, nil)
				user := testutils.CreateTestUser("123", "testuser", "avatar")
				interaction.Member = testutils.CreateTestMember(user)
			case "8ball":
				options := []*discordgo.ApplicationCommandInteractionDataOption{
					testutils.CreateStringOption("question", "Test?"),
				}
				interaction = testutils.CreateTestInteraction("8ball", options)
			case "coinflip":
				interaction = testutils.CreateTestInteraction("coinflip", nil)
			}

			// Execute command
			var err error
			switch tt.command {
			case "ping":
				err = commands.HandlePingCommand(mockSession, interaction)
			case "peepee":
				err = commands.HandlePeepeeCommand(mockSession, interaction)
			case "8ball":
				err = commands.Handle8BallCommand(mockSession, interaction)
			case "coinflip":
				err = commands.HandleCoinFlipCommand(mockSession, interaction)
			case "user":
				err = commands.HandleUserCommand(mockSession, interaction)
			}

			if err != nil {
				t.Errorf("Command %s failed: %v", tt.command, err)
			}

			if !mockSession.RespondCalled {
				t.Errorf("Command %s did not respond", tt.command)
			}

			// Verify response type
			if mockSession.RespondData != nil {
				hasText := mockSession.RespondData.Content != ""
				hasEmbed := len(mockSession.RespondData.Embeds) > 0

				if tt.expectText && !hasText {
					t.Errorf("Command %s expected text response but got none", tt.command)
				}
				if tt.expectEmbed && !hasEmbed {
					t.Errorf("Command %s expected embed response but got none", tt.command)
				}
				if !tt.expectText && hasText {
					t.Errorf("Command %s did not expect text response but got: %s",
						tt.command, mockSession.RespondData.Content)
				}
				if !tt.expectEmbed && hasEmbed {
					t.Errorf("Command %s did not expect embed response but got %d embeds",
						tt.command, len(mockSession.RespondData.Embeds))
				}
			}
		})
	}
}
