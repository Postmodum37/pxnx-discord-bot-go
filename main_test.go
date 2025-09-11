package main

import (
	"os"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestGetCommands(t *testing.T) {
	commands := getCommands()
	
	if len(commands) != 7 {
		t.Errorf("Expected 7 commands, got %d", len(commands))
	}
	
	// Test that all expected commands are present
	expectedCommands := map[string]string{
		"ping":     "Responds with Pong!",
		"peepee":   "PeePee Inspection Time!",
		"8ball":    "Ask the magic 8-ball a question",
		"coinflip": "Flip a coin and choose heads or tails",
		"server":   "Provides information about the server",
		"user":     "Replies with user info!",
		"weather":  "Get the weather forecast for a city",
	}
	
	foundCommands := make(map[string]bool)
	
	for _, cmd := range commands {
		if expectedDesc, exists := expectedCommands[cmd.Name]; exists {
			foundCommands[cmd.Name] = true
			if cmd.Description != expectedDesc {
				t.Errorf("Expected %s description '%s', got '%s'", cmd.Name, expectedDesc, cmd.Description)
			}
		} else {
			t.Errorf("Unexpected command found: %s", cmd.Name)
		}
	}
	
	// Check that all expected commands were found
	for cmdName := range expectedCommands {
		if !foundCommands[cmdName] {
			t.Errorf("Command '%s' not found", cmdName)
		}
	}
}

func TestGetRandomPhrase(t *testing.T) {
	username := "testuser"
	phrase := getRandomPhrase(username)
	
	if phrase == "" {
		t.Error("Expected non-empty phrase, got empty string")
	}
	
	// Check if the phrase starts with username and ends with "peepee!"
	if !strings.HasPrefix(phrase, username) {
		t.Errorf("Expected phrase to start with '%s', got '%s'", username, phrase)
	}
	
	if !strings.HasSuffix(phrase, "peepee!") {
		t.Errorf("Expected phrase to end with 'peepee!', got '%s'", phrase)
	}
	
	// Check if the middle part contains one of the definitions
	middlePart := strings.TrimPrefix(phrase, username+" ")
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
	
	// Check if description follows the expected format: "username definition peepee!"
	if !strings.HasPrefix(embed.Description, user.Username) {
		t.Errorf("Expected description to start with '%s', got '%s'", user.Username, embed.Description)
	}
	
	if !strings.HasSuffix(embed.Description, "peepee!") {
		t.Errorf("Expected description to end with 'peepee!', got '%s'", embed.Description)
	}
	
	// Check if the middle part contains one of the definitions
	middlePart := strings.TrimPrefix(embed.Description, user.Username+" ")
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

func TestPeepeeDefinitionsNotEmpty(t *testing.T) {
	if len(peepeeDefinitions) == 0 {
		t.Error("Expected peepeeDefinitions to contain definitions, got empty slice")
	}
	
	for i, definition := range peepeeDefinitions {
		if definition == "" {
			t.Errorf("Expected non-empty definition at index %d", i)
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

func TestGet8BallResponse(t *testing.T) {
	response := get8BallResponse()
	
	if response == "" {
		t.Error("Expected non-empty 8-ball response, got empty string")
	}
	
	// Check if response is one of the expected responses
	found := false
	for _, expectedResponse := range eightBallResponses {
		if response == expectedResponse {
			found = true
			break
		}
	}
	
	if !found {
		t.Errorf("Response '%s' not found in expected responses", response)
	}
}

func TestHandle8BallCommand(t *testing.T) {
	mockSession := &MockSession{}
	
	// Create interaction with question option
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{
						Name:  "question",
						Type:  discordgo.ApplicationCommandOptionString,
						Value: "Will this test pass?",
					},
				},
			},
		},
	}
	
	err := handle8BallCommand(mockSession, interaction)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if !mockSession.respondCalled {
		t.Error("Expected InteractionRespond to be called")
	}
}

func TestHandleCoinFlipCommand(t *testing.T) {
	mockSession := &MockSession{}
	
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{},
	}
	
	err := handleCoinFlipCommand(mockSession, interaction)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if !mockSession.respondCalled {
		t.Error("Expected InteractionRespond to be called")
	}
}

func TestHandleUserCommand(t *testing.T) {
	mockSession := &MockSession{}
	
	// Test with no target user (should use command invoker)
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Options: []*discordgo.ApplicationCommandInteractionDataOption{},
			},
		},
	}
	interaction.Member = &discordgo.Member{
		User: &discordgo.User{
			ID:       "123456789",
			Username: "testuser",
			Avatar:   "test_avatar",
		},
	}
	
	err := handleUserCommand(mockSession, interaction)
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if !mockSession.respondCalled {
		t.Error("Expected InteractionRespond to be called")
	}
}

func TestGetWeatherIcon(t *testing.T) {
	tests := []struct {
		condition string
		expected  string
	}{
		{"clear", "☀️"},
		{"Clear", "☀️"},
		{"cloudy", "☁️"},
		{"Clouds", "☁️"},
		{"rain", "🌧️"},
		{"Rain", "🌧️"},
		{"snow", "❄️"},
		{"Snow", "❄️"},
		{"thunder", "⛈️"},
		{"Thunderstorm", "⛈️"},
		{"mist", "🌫️"},
		{"fog", "🌫️"},
		{"unknown", "🌤️"},
		{"", "🌤️"},
	}
	
	for _, tt := range tests {
		t.Run(tt.condition, func(t *testing.T) {
			result := getWeatherIcon(tt.condition)
			if result != tt.expected {
				t.Errorf("getWeatherIcon(%q) = %q, want %q", tt.condition, result, tt.expected)
			}
		})
	}
}

func TestEightBallResponsesNotEmpty(t *testing.T) {
	if len(eightBallResponses) == 0 {
		t.Error("Expected eightBallResponses to contain responses, got empty slice")
	}
	
	for i, response := range eightBallResponses {
		if response == "" {
			t.Errorf("Expected non-empty response at index %d", i)
		}
	}
}

func TestCreateStringOption(t *testing.T) {
	option := createStringOption("test", "Test description", true)
	
	if option.Type != discordgo.ApplicationCommandOptionString {
		t.Errorf("Expected type String, got %v", option.Type)
	}
	if option.Name != "test" {
		t.Errorf("Expected name 'test', got %s", option.Name)
	}
	if option.Description != "Test description" {
		t.Errorf("Expected description 'Test description', got %s", option.Description)
	}
	if !option.Required {
		t.Error("Expected option to be required")
	}
}

func TestCreateUserOption(t *testing.T) {
	option := createUserOption("target", "Target user", false)
	
	if option.Type != discordgo.ApplicationCommandOptionUser {
		t.Errorf("Expected type User, got %v", option.Type)
	}
	if option.Name != "target" {
		t.Errorf("Expected name 'target', got %s", option.Name)
	}
	if option.Description != "Target user" {
		t.Errorf("Expected description 'Target user', got %s", option.Description)
	}
	if option.Required {
		t.Error("Expected option to not be required")
	}
}

func TestCreateErrorEmbed(t *testing.T) {
	embed := createErrorEmbed("Error Title", "Error Description", "Error message")
	
	if embed.Title != "Error Title" {
		t.Errorf("Expected title 'Error Title', got %s", embed.Title)
	}
	if embed.Description != "Error Description" {
		t.Errorf("Expected description 'Error Description', got %s", embed.Description)
	}
	if embed.Color != ColorRed {
		t.Errorf("Expected color ColorRed (%d), got %d", ColorRed, embed.Color)
	}
	if len(embed.Fields) != 1 {
		t.Errorf("Expected 1 field, got %d", len(embed.Fields))
	}
	if embed.Fields[0].Name != "Error" {
		t.Errorf("Expected field name 'Error', got %s", embed.Fields[0].Name)
	}
	if embed.Fields[0].Value != "Error message" {
		t.Errorf("Expected field value 'Error message', got %s", embed.Fields[0].Value)
	}
}

func TestColorConstants(t *testing.T) {
	// Test that color constants are defined correctly
	expectedColors := map[string]int{
		"ColorBlue":   0x3498db,
		"ColorPurple": 0x9b59b6,
		"ColorOrange": 0xf39c12,
		"ColorGreen":  0x2ecc71,
		"ColorRed":    0xe74c3c,
	}
	
	actualColors := map[string]int{
		"ColorBlue":   ColorBlue,
		"ColorPurple": ColorPurple,
		"ColorOrange": ColorOrange,
		"ColorGreen":  ColorGreen,
		"ColorRed":    ColorRed,
	}
	
	for name, expected := range expectedColors {
		if actual := actualColors[name]; actual != expected {
			t.Errorf("Expected %s to be 0x%x, got 0x%x", name, expected, actual)
		}
	}
}

// Test that covers the error handling in handleWeatherCommand
func TestHandleWeatherCommandError(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("OPENWEATHER_API_KEY")
	
	// Temporarily unset the API key to trigger error
	os.Unsetenv("OPENWEATHER_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENWEATHER_API_KEY", originalKey)
		}
	}()
	
	mockSession := &MockSession{}
	
	interaction := &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{
						Name:  "city",
						Type:  discordgo.ApplicationCommandOptionString,
						Value: "TestCity",
					},
				},
			},
		},
	}
	
	err := handleWeatherCommand(mockSession, interaction)
	
	if err != nil {
		t.Errorf("Expected no error (should handle gracefully), got %v", err)
	}
	
	if !mockSession.respondCalled {
		t.Error("Expected InteractionRespond to be called for error case")
	}
}

// Benchmark tests
func BenchmarkGetRandomPhrase(b *testing.B) {
	username := "benchuser"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getRandomPhrase(username)
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

func BenchmarkGet8BallResponse(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		get8BallResponse()
	}
}

func BenchmarkGetWeatherIcon(b *testing.B) {
	conditions := []string{"clear", "cloudy", "rain", "snow", "thunder"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		getWeatherIcon(conditions[i%len(conditions)])
	}
}