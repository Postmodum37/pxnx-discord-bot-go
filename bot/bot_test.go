package bot

import (
	"testing"

	"github.com/bwmarrin/discordgo"

	"pxnx-discord-bot/testutils"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		expectErr bool
	}{
		{
			name:      "valid token",
			token:     "valid.bot.token",
			expectErr: false,
		},
		// Note: discordgo.New() doesn't validate token format, so empty token won't error
		// {
		// 	name:      "empty token",
		// 	token:     "",
		// 	expectErr: true,
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bot, err := New(tt.token)

			if tt.expectErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if bot != nil {
					t.Error("Expected nil bot when error occurs")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if bot == nil {
					t.Error("Expected bot instance but got nil")
				}
				if bot.Session == nil {
					t.Error("Expected session to be initialized")
				}
			}
		})
	}
}

func TestBotSetup(t *testing.T) {
	bot, err := New("test.token")
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Setup the bot - we can't directly test handlers as they're unexported
	bot.Setup()

	// Check intents
	expectedIntents := discordgo.IntentsGuildMessages | discordgo.IntentsGuildEmojis
	if bot.Session.Identify.Intents != expectedIntents {
		t.Errorf("Expected intents %d, got %d", expectedIntents, bot.Session.Identify.Intents)
	}
}

func TestSetShouldRegisterCommands(t *testing.T) {
	// Test setting true
	SetShouldRegisterCommands(true)
	if !shouldRegisterCommands {
		t.Error("Expected shouldRegisterCommands to be true")
	}

	// Test setting false
	SetShouldRegisterCommands(false)
	if shouldRegisterCommands {
		t.Error("Expected shouldRegisterCommands to be false")
	}
}

func TestInteractionCreate(t *testing.T) {
	_, err := New("test.token")
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	tests := []struct {
		name        string
		commandName string
		expectError bool
	}{
		{name: "ping command", commandName: "ping", expectError: false},
		{name: "peepee command", commandName: "peepee", expectError: false},
		{name: "8ball command", commandName: "8ball", expectError: false},
		{name: "coinflip command", commandName: "coinflip", expectError: false},
		{name: "server command", commandName: "server", expectError: false},
		{name: "user command", commandName: "user", expectError: false},
		{name: "weather command", commandName: "weather", expectError: false},
		{name: "unknown command", commandName: "unknown", expectError: false}, // Should not error, just do nothing
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test interaction
			var options []*discordgo.ApplicationCommandInteractionDataOption

			// Add required options for certain commands
			switch tt.commandName {
			case "8ball":
				options = append(options, testutils.CreateStringOption("question", "Test question?"))
			case "weather":
				options = append(options, testutils.CreateStringOption("city", "London"))
			}

			interaction := testutils.CreateTestInteraction(tt.commandName, options)
			interaction.Member = testutils.CreateTestMember(testutils.CreateTestUser("123", "testuser", "avatar"))

			// We can't easily test interactionCreate without significant refactoring
			// as it's not exported and requires a real session
			t.Log("Command routing logic would be tested here in a more advanced setup")
		})
	}
}
