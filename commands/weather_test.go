package commands

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"

	"pxnx-discord-bot/testutils"
)

func TestGetWeatherIcon(t *testing.T) {
	tests := []struct {
		condition string
		expected  string
	}{
		{"clear", "☀️"},
		{"Clear", "☀️"},
		{"CLEAR SKY", "☀️"},
		{"cloudy", "☁️"},
		{"Clouds", "☁️"},
		{"PARTLY CLOUDY", "☁️"},
		{"rain", "🌧️"},
		{"Rain", "🌧️"},
		{"LIGHT RAIN", "🌧️"},
		{"snow", "❄️"},
		{"Snow", "❄️"},
		{"HEAVY SNOW", "❄️"},
		{"thunder", "⛈️"},
		{"Thunderstorm", "⛈️"},
		{"THUNDER AND LIGHTNING", "⛈️"},
		{"mist", "🌫️"},
		{"fog", "🌫️"},
		{"Fog", "🌫️"},
		{"DENSE FOG", "🌫️"},
		{"unknown condition", "🌤️"},
		{"", "🌤️"},
		{"random weather", "🌤️"},
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

func TestCreateErrorEmbed(t *testing.T) {
	tests := []struct {
		name        string
		title       string
		description string
		errorMsg    string
	}{
		{
			name:        "standard error",
			title:       "Error Title",
			description: "Error Description",
			errorMsg:    "Something went wrong",
		},
		{
			name:        "empty strings",
			title:       "",
			description: "",
			errorMsg:    "",
		},
		{
			name:        "long messages",
			title:       "Very Long Error Title That Exceeds Normal Length",
			description: "Very long error description with lots of details about what went wrong",
			errorMsg:    "Very detailed error message with technical information",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			embed := createErrorEmbed(tt.title, tt.description, tt.errorMsg)

			if embed.Title != tt.title {
				t.Errorf("Expected title '%s', got '%s'", tt.title, embed.Title)
			}
			if embed.Description != tt.description {
				t.Errorf("Expected description '%s', got '%s'", tt.description, embed.Description)
			}
			if embed.Color != 0xe74c3c {
				t.Errorf("Expected color 0xe74c3c, got 0x%x", embed.Color)
			}
			if len(embed.Fields) != 1 {
				t.Errorf("Expected 1 field, got %d", len(embed.Fields))
			}
			if len(embed.Fields) > 0 {
				if embed.Fields[0].Name != "Error" {
					t.Errorf("Expected field name 'Error', got '%s'", embed.Fields[0].Name)
				}
				if embed.Fields[0].Value != tt.errorMsg {
					t.Errorf("Expected field value '%s', got '%s'", tt.errorMsg, embed.Fields[0].Value)
				}
			}
		})
	}
}

func TestHandleWeatherCommand(t *testing.T) {
	// Save original env var
	originalKey := os.Getenv("OPENWEATHER_API_KEY")
	defer func() {
		if originalKey != "" {
			os.Setenv("OPENWEATHER_API_KEY", originalKey)
		}
	}()

	tests := []struct {
		name         string
		city         string
		apiKey       string
		sessionError error
		expectError  bool
		expectEmbed  bool
	}{
		{
			name:         "no API key - should return error embed",
			city:         "London",
			apiKey:       "",
			sessionError: nil,
			expectError:  false, // Function handles gracefully
			expectEmbed:  true,
		},
		{
			name:         "valid city with API key",
			city:         "TestCity",
			apiKey:       "test_api_key",
			sessionError: nil,
			expectError:  false,
			expectEmbed:  true,
		},
		{
			name:         "session error",
			city:         "London",
			apiKey:       "test_key",
			sessionError: errors.New("respond error"),
			expectError:  true,
			expectEmbed:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.apiKey == "" {
				os.Unsetenv("OPENWEATHER_API_KEY")
			} else {
				os.Setenv("OPENWEATHER_API_KEY", tt.apiKey)
			}

			mockSession := &testutils.MockSession{
				RespondError: tt.sessionError,
			}

			options := []*discordgo.ApplicationCommandInteractionDataOption{
				testutils.CreateStringOption("city", tt.city),
			}
			interaction := testutils.CreateTestInteraction("weather", options)

			err := HandleWeatherCommand(mockSession, interaction)

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

			if tt.expectEmbed && mockSession.RespondData != nil {
				if len(mockSession.RespondData.Embeds) != 1 {
					t.Errorf("Expected 1 embed, got %d", len(mockSession.RespondData.Embeds))
					return
				}

				embed := mockSession.RespondData.Embeds[0]

				// For error cases, check error embed
				if tt.apiKey == "" {
					if !strings.Contains(embed.Title, "Weather Error") {
						t.Errorf("Expected error title, got '%s'", embed.Title)
					}
					if embed.Color != 0xe74c3c {
						t.Errorf("Expected error color 0xe74c3c, got 0x%x", embed.Color)
					}
				}

				// Check that footer is present for weather embeds
				if embed.Footer != nil {
					if embed.Footer.Text != "Powered by OpenWeatherMap" {
						t.Errorf("Expected footer 'Powered by OpenWeatherMap', got '%s'",
							embed.Footer.Text)
					}
				}
			}
		})
	}
}
