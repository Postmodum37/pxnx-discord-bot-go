package bot

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestCreateStringOption(t *testing.T) {
	tests := []struct {
		name        string
		optionName  string
		description string
		required    bool
	}{
		{
			name:        "required string option",
			optionName:  "test",
			description: "Test description",
			required:    true,
		},
		{
			name:        "optional string option",
			optionName:  "optional",
			description: "Optional description",
			required:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := createStringOption(tt.optionName, tt.description, tt.required)

			if option.Type != discordgo.ApplicationCommandOptionString {
				t.Errorf("Expected type String, got %v", option.Type)
			}
			if option.Name != tt.optionName {
				t.Errorf("Expected name '%s', got '%s'", tt.optionName, option.Name)
			}
			if option.Description != tt.description {
				t.Errorf("Expected description '%s', got '%s'", tt.description, option.Description)
			}
			if option.Required != tt.required {
				t.Errorf("Expected required %v, got %v", tt.required, option.Required)
			}
		})
	}
}

func TestCreateUserOption(t *testing.T) {
	tests := []struct {
		name        string
		optionName  string
		description string
		required    bool
	}{
		{
			name:        "required user option",
			optionName:  "target",
			description: "Target user",
			required:    true,
		},
		{
			name:        "optional user option",
			optionName:  "user",
			description: "Optional user",
			required:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			option := createUserOption(tt.optionName, tt.description, tt.required)

			if option.Type != discordgo.ApplicationCommandOptionUser {
				t.Errorf("Expected type User, got %v", option.Type)
			}
			if option.Name != tt.optionName {
				t.Errorf("Expected name '%s', got '%s'", tt.optionName, option.Name)
			}
			if option.Description != tt.description {
				t.Errorf("Expected description '%s', got '%s'", tt.description, option.Description)
			}
			if option.Required != tt.required {
				t.Errorf("Expected required %v, got %v", tt.required, option.Required)
			}
		})
	}
}

func TestGetCommands(t *testing.T) {
	commands := GetCommands()

	expectedCount := 7
	if len(commands) != expectedCount {
		t.Errorf("Expected %d commands, got %d", expectedCount, len(commands))
	}

	// Test that all expected commands are present with correct structure
	expectedCommands := map[string]struct {
		description string
		hasOptions  bool
		optionCount int
	}{
		"ping":     {"Responds with Pong!", false, 0},
		"peepee":   {"PeePee Inspection Time!", false, 0},
		"8ball":    {"Ask the magic 8-ball a question", true, 1},
		"coinflip": {"Flip a coin and choose heads or tails", false, 0},
		"server":   {"Provides information about the server", false, 0},
		"user":     {"Replies with user info!", true, 1},
		"weather":  {"Get the weather forecast for a city", true, 1},
	}

	foundCommands := make(map[string]bool)

	for _, cmd := range commands {
		expected, exists := expectedCommands[cmd.Name]
		if !exists {
			t.Errorf("Unexpected command found: %s", cmd.Name)
			continue
		}

		foundCommands[cmd.Name] = true

		if cmd.Description != expected.description {
			t.Errorf("Command %s: expected description '%s', got '%s'", 
				cmd.Name, expected.description, cmd.Description)
		}

		if expected.hasOptions {
			if len(cmd.Options) != expected.optionCount {
				t.Errorf("Command %s: expected %d options, got %d", 
					cmd.Name, expected.optionCount, len(cmd.Options))
			}
		} else {
			if len(cmd.Options) != 0 {
				t.Errorf("Command %s: expected no options, got %d", 
					cmd.Name, len(cmd.Options))
			}
		}
	}

	// Check that all expected commands were found
	for cmdName := range expectedCommands {
		if !foundCommands[cmdName] {
			t.Errorf("Command '%s' not found", cmdName)
		}
	}
}

func TestGetCommandsOptionsValidation(t *testing.T) {
	commands := GetCommands()

	for _, cmd := range commands {
		switch cmd.Name {
		case "8ball":
			if len(cmd.Options) != 1 {
				t.Errorf("8ball command should have 1 option, got %d", len(cmd.Options))
			} else {
				option := cmd.Options[0]
				if option.Name != "question" {
					t.Errorf("8ball option should be named 'question', got '%s'", option.Name)
				}
				if option.Type != discordgo.ApplicationCommandOptionString {
					t.Errorf("8ball option should be string type, got %v", option.Type)
				}
				if !option.Required {
					t.Error("8ball question option should be required")
				}
			}

		case "user":
			if len(cmd.Options) != 1 {
				t.Errorf("user command should have 1 option, got %d", len(cmd.Options))
			} else {
				option := cmd.Options[0]
				if option.Name != "target" {
					t.Errorf("user option should be named 'target', got '%s'", option.Name)
				}
				if option.Type != discordgo.ApplicationCommandOptionUser {
					t.Errorf("user option should be user type, got %v", option.Type)
				}
				if option.Required {
					t.Error("user target option should not be required")
				}
			}

		case "weather":
			if len(cmd.Options) != 1 {
				t.Errorf("weather command should have 1 option, got %d", len(cmd.Options))
			} else {
				option := cmd.Options[0]
				if option.Name != "city" {
					t.Errorf("weather option should be named 'city', got '%s'", option.Name)
				}
				if option.Type != discordgo.ApplicationCommandOptionString {
					t.Errorf("weather option should be string type, got %v", option.Type)
				}
				if !option.Required {
					t.Error("weather city option should be required")
				}
			}
		}
	}
}