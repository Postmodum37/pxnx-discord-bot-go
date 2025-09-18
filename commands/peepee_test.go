package commands

import (
	"errors"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"

	"pxnx-discord-bot/testutils"
)

func TestGetRandomPhrase(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
	}{
		{name: "regular display name", displayName: "testuser"},
		{name: "empty display name", displayName: ""},
		{name: "special characters", displayName: "user@123"},
		{name: "long display name", displayName: "verylongdisplaynamethatexceedsnormallimits"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phrase := getRandomPhrase(tt.displayName)

			if phrase == "" {
				t.Error("Expected non-empty phrase, got empty string")
			}

			if !strings.HasPrefix(phrase, tt.displayName) {
				t.Errorf("Expected phrase to start with '%s', got '%s'", tt.displayName, phrase)
			}

			if !strings.HasSuffix(phrase, "peepee!") {
				t.Errorf("Expected phrase to end with 'peepee!', got '%s'", phrase)
			}

			// Check if the middle part contains one of the definitions
			middlePart := strings.TrimPrefix(phrase, tt.displayName+" ")
			middlePart = strings.TrimSuffix(middlePart, " peepee!")

			found := false
			for _, definition := range peepeeDefinitions {
				if middlePart == definition {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("Definition '%s' not found in expected definitions", middlePart)
			}
		})
	}
}

func TestGetUserAvatarURL(t *testing.T) {
	tests := []struct {
		name           string
		user           *discordgo.User
		expectContains string
	}{
		{
			name: "user with custom avatar",
			user: &discordgo.User{
				ID:     "123456789",
				Avatar: "custom_avatar_hash",
			},
			expectContains: "custom_avatar_hash",
		},
		{
			name: "user without custom avatar",
			user: &discordgo.User{
				ID:            "123456789",
				Avatar:        "",
				Discriminator: "0001",
			},
			expectContains: "discordapp.com",
		},
		{
			name: "user with nil avatar",
			user: &discordgo.User{
				ID:            "987654321",
				Discriminator: "0005",
			},
			expectContains: "discordapp.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			avatarURL := getUserAvatarURL(tt.user)

			if avatarURL == "" {
				t.Error("Expected non-empty avatar URL")
			}

			if !strings.Contains(avatarURL, tt.expectContains) {
				t.Errorf("Expected avatar URL to contain '%s', got '%s'",
					tt.expectContains, avatarURL)
			}
		})
	}
}

func TestCreatePeepeeEmbed(t *testing.T) {
	t.Run("user with GlobalName", func(t *testing.T) {
		user := testutils.CreateTestUser("123456789", "testuser", "test_avatar_hash")
		user.GlobalName = "Test Display Name"

		embed := createPeepeeEmbed(user)

		if embed.Title != "PeePee Inspection Time" {
			t.Errorf("Expected title 'PeePee Inspection Time', got '%s'", embed.Title)
		}

		if embed.Color != 0x3498db {
			t.Errorf("Expected color 0x3498db, got 0x%x", embed.Color)
		}

		if embed.Description == "" {
			t.Error("Expected non-empty description")
		}

		if embed.Thumbnail == nil {
			t.Error("Expected thumbnail to be set")
		} else {
			if embed.Thumbnail.URL == "" {
				t.Error("Expected thumbnail URL to be set")
			}
		}

		// Check if description uses GlobalName
		if !strings.HasPrefix(embed.Description, user.GlobalName) {
			t.Errorf("Expected description to start with GlobalName '%s', got '%s'",
				user.GlobalName, embed.Description)
		}

		if !strings.HasSuffix(embed.Description, "peepee!") {
			t.Errorf("Expected description to end with 'peepee!', got '%s'",
				embed.Description)
		}
	})

	t.Run("user without GlobalName fallback to Username", func(t *testing.T) {
		user := testutils.CreateTestUser("123456789", "testuser", "test_avatar_hash")
		user.GlobalName = "" // Explicitly empty

		embed := createPeepeeEmbed(user)

		// Check if description falls back to Username
		if !strings.HasPrefix(embed.Description, user.Username) {
			t.Errorf("Expected description to start with Username '%s' when GlobalName is empty, got '%s'",
				user.Username, embed.Description)
		}

		if !strings.HasSuffix(embed.Description, "peepee!") {
			t.Errorf("Expected description to end with 'peepee!', got '%s'",
				embed.Description)
		}
	})
}

func TestGetRandomEmoji(t *testing.T) {
	// Test fallback cases only since getRandomEmoji expects *discordgo.Session

	t.Run("nil session returns fallback", func(t *testing.T) {
		result := getRandomEmoji(nil, "guild123")
		if result != "🔍" {
			t.Errorf("Expected fallback emoji '🔍', got '%s'", result)
		}
	})

	t.Run("empty guild ID returns fallback", func(t *testing.T) {
		// Even with a nil session, empty guild ID should return fallback
		result := getRandomEmoji(nil, "")
		if result != "🔍" {
			t.Errorf("Expected fallback emoji '🔍', got '%s'", result)
		}
	})

	// Note: Testing with real Discord session would require more complex mocking
	// as getRandomEmoji expects *discordgo.Session specifically, not our interface
}

func TestHandlePeepeeCommand(t *testing.T) {
	tests := []struct {
		name         string
		sessionError error
		expectError  bool
	}{
		{
			name:         "successful command",
			sessionError: nil,
			expectError:  false,
		},
		{
			name:         "session error",
			sessionError: errors.New("respond error"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &testutils.MockSession{
				RespondError: tt.sessionError,
			}

			user := testutils.CreateTestUser("123", "testuser", "avatar")
			interaction := testutils.CreateTestInteraction("peepee", nil)
			interaction.Member = testutils.CreateTestMember(user)

			err := HandlePeepeeCommand(mockSession, interaction)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			if !mockSession.RespondCalled {
				t.Error("Expected InteractionRespond to be called")
			}

			// Verify embed was created correctly
			if !tt.expectError && mockSession.RespondData != nil {
				if len(mockSession.RespondData.Embeds) != 1 {
					t.Errorf("Expected 1 embed, got %d", len(mockSession.RespondData.Embeds))
				}
			}
		})
	}
}

func TestPeepeeDefinitionsNotEmpty(t *testing.T) {
	if len(peepeeDefinitions) == 0 {
		t.Error("Expected peepeeDefinitions to contain definitions, got empty slice")
	}

	for i, definition := range peepeeDefinitions {
		if definition == "" {
			t.Errorf("Expected non-empty definition at index %d", i)
		}

		// Check that definitions are reasonable in length
		if len(definition) > 50 {
			t.Errorf("Definition at index %d is too long (%d chars): %s",
				i, len(definition), definition)
		}
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for i, definition := range peepeeDefinitions {
		if seen[definition] {
			t.Errorf("Duplicate definition at index %d: %s", i, definition)
		}
		seen[definition] = true
	}
}
