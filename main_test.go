package main

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestGetCommands(t *testing.T) {
	commands := getCommands()
	
	if len(commands) != 2 {
		t.Errorf("Expected 2 commands, got %d", len(commands))
	}
	
	// Test ping command
	pingFound := false
	peepeeFound := false
	
	for _, cmd := range commands {
		switch cmd.Name {
		case "ping":
			pingFound = true
			if cmd.Description != "Responds with Pong!" {
				t.Errorf("Expected ping description 'Responds with Pong!', got '%s'", cmd.Description)
			}
		case "peepee":
			peepeeFound = true
			if cmd.Description != "PeePee Inspection Time!" {
				t.Errorf("Expected peepee description 'PeePee Inspection Time!', got '%s'", cmd.Description)
			}
		}
	}
	
	if !pingFound {
		t.Error("Ping command not found")
	}
	if !peepeeFound {
		t.Error("Peepee command not found")
	}
}

func TestGetRandomPhrase(t *testing.T) {
	phrase := getRandomPhrase()
	
	if phrase == "" {
		t.Error("Expected non-empty phrase, got empty string")
	}
	
	// Check if the phrase is one of the expected phrases
	found := false
	for _, expected := range peepeePhrasces {
		if phrase == expected {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Phrase '%s' not found in expected phrases", phrase)
	}
}

func TestGetUserAvatarURL(t *testing.T) {
	// Test with user that has custom avatar
	userWithAvatar := &discordgo.User{
		ID:     "123456789",
		Avatar: "custom_avatar_hash",
	}
	
	avatarURL := getUserAvatarURL(userWithAvatar)
	if !strings.Contains(avatarURL, "custom_avatar_hash") {
		t.Errorf("Expected avatar URL to contain custom_avatar_hash, got %s", avatarURL)
	}
	
	// Test with user that has no custom avatar
	userWithoutAvatar := &discordgo.User{
		ID:            "123456789",
		Avatar:        "",
		Discriminator: "0001",
	}
	
	defaultURL := getUserAvatarURL(userWithoutAvatar)
	if !strings.Contains(defaultURL, "discordapp.com") {
		t.Errorf("Expected default avatar URL to contain discordapp.com, got %s", defaultURL)
	}
}

func TestCreatePeepeeEmbed(t *testing.T) {
	user := &discordgo.User{
		ID:       "123456789",
		Username: "testuser",
		Avatar:   "test_avatar_hash",
	}
	
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
	
	// Check if description is one of the expected phrases
	found := false
	for _, expected := range peepeePhrasces {
		if embed.Description == expected {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Description '%s' not found in expected phrases", embed.Description)
	}
}

func TestGetRandomEmoji(t *testing.T) {
	// This test checks the fallback behavior when session is nil or guild has no emojis
	// We can't easily test the success case without a real Discord session
	// So we'll test that the function doesn't panic and returns the fallback
	
	// Test with empty guild ID - should return fallback
	fallbackEmoji := getRandomEmoji(nil, "")
	
	if fallbackEmoji != "🔍" {
		t.Errorf("Expected fallback emoji '🔍', got '%s'", fallbackEmoji)
	}
}

func TestPeepeePhrasesNotEmpty(t *testing.T) {
	if len(peepeePhrasces) == 0 {
		t.Error("Expected peepeePhrasces to contain phrases, got empty slice")
	}
	
	for i, phrase := range peepeePhrasces {
		if phrase == "" {
			t.Errorf("Expected non-empty phrase at index %d", i)
		}
	}
}

// MockSession implements a minimal mock for testing command handlers
type MockSession struct {
	respondCalled bool
	respondError  error
}

func (m *MockSession) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	m.respondCalled = true
	return m.respondError
}

func TestHandlePingCommand(t *testing.T) {
	mockSession := &MockSession{}
	
	// Create a minimal interaction for testing
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{},
	}
	
	// Test successful ping command
	err := handlePingCommand(mockSession, interaction)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if !mockSession.respondCalled {
		t.Error("Expected InteractionRespond to be called")
	}
}

func TestHandlePeepeeCommand(t *testing.T) {
	mockSession := &MockSession{}
	
	// Create a minimal interaction with user data for testing
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{},
	}
	interaction.Member = &discordgo.Member{
		User: &discordgo.User{
			ID:       "123456789",
			Username: "testuser",
			Avatar:   "test_avatar",
		},
	}
	
	// Test successful peepee command
	err := handlePeepeeCommand(mockSession, interaction)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if !mockSession.respondCalled {
		t.Error("Expected InteractionRespond to be called")
	}
}

// Benchmark tests
func BenchmarkGetRandomPhrase(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getRandomPhrase()
	}
}

func BenchmarkCreatePeepeeEmbed(b *testing.B) {
	user := &discordgo.User{
		ID:       "123456789",
		Username: "testuser",
		Avatar:   "test_avatar_hash",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createPeepeeEmbed(user)
	}
}