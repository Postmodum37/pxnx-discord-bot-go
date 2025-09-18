package commands

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"

	"pxnx-discord-bot/testutils"
)

func TestGet8BallResponse(t *testing.T) {
	// Test that function returns a valid response
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

func TestGet8BallResponseDistribution(t *testing.T) {
	// Test that multiple calls return different responses (not always the same)
	responses := make(map[string]int)
	iterations := 100

	for i := 0; i < iterations; i++ {
		response := get8BallResponse()
		responses[response]++
	}

	// Should have at least some variety (not just one response)
	if len(responses) == 1 {
		t.Error("Expected some variety in responses, got only one unique response")
	}

	// All responses should be from the expected list
	for response := range responses {
		found := false
		for _, expectedResponse := range eightBallResponses {
			if response == expectedResponse {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Unexpected response: %s", response)
		}
	}
}

func TestHandle8BallCommand(t *testing.T) {
	tests := []struct {
		name         string
		question     string
		sessionError error
		expectError  bool
	}{
		{
			name:         "successful command with question",
			question:     "Will this test pass?",
			sessionError: nil,
			expectError:  false,
		},
		{
			name:         "successful command with long question",
			question:     "This is a very long question that tests whether the 8-ball can handle lengthy inputs?",
			sessionError: nil,
			expectError:  false,
		},
		{
			name:         "session error",
			question:     "Any question?",
			sessionError: errors.New("respond error"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &testutils.MockSession{
				RespondError: tt.sessionError,
			}

			options := []*discordgo.ApplicationCommandInteractionDataOption{
				testutils.CreateStringOption("question", tt.question),
			}
			interaction := testutils.CreateTestInteraction("8ball", options)

			err := Handle8BallCommand(mockSession, interaction)

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

			// Verify embed structure
			if !tt.expectError && mockSession.RespondData != nil {
				if len(mockSession.RespondData.Embeds) != 1 {
					t.Errorf("Expected 1 embed, got %d", len(mockSession.RespondData.Embeds))
					return
				}

				embed := mockSession.RespondData.Embeds[0]
				if embed.Title != "🎱 Magic 8-Ball" {
					t.Errorf("Expected title '🎱 Magic 8-Ball', got '%s'", embed.Title)
				}

				if embed.Color != 0x9b59b6 {
					t.Errorf("Expected color 0x9b59b6, got 0x%x", embed.Color)
				}

				if len(embed.Fields) != 2 {
					t.Errorf("Expected 2 fields, got %d", len(embed.Fields))
					return
				}

				// Check question field
				if embed.Fields[0].Name != "Question" {
					t.Errorf("Expected first field name 'Question', got '%s'", embed.Fields[0].Name)
				}
				if embed.Fields[0].Value != tt.question {
					t.Errorf("Expected question '%s', got '%s'", tt.question, embed.Fields[0].Value)
				}

				// Check answer field
				if embed.Fields[1].Name != "Answer" {
					t.Errorf("Expected second field name 'Answer', got '%s'", embed.Fields[1].Name)
				}
				if embed.Fields[1].Value == "" {
					t.Error("Expected non-empty answer")
				}

				// Verify answer is from expected responses
				answer := embed.Fields[1].Value
				found := false
				for _, expectedResponse := range eightBallResponses {
					if answer == expectedResponse {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Answer '%s' not found in expected responses", answer)
				}
			}
		})
	}
}

func TestEightBallResponsesNotEmpty(t *testing.T) {
	if len(eightBallResponses) == 0 {
		t.Error("Expected eightBallResponses to contain responses, got empty slice")
	}

	expectedCount := 20 // Standard magic 8-ball has 20 responses
	if len(eightBallResponses) != expectedCount {
		t.Errorf("Expected %d responses (standard 8-ball), got %d", expectedCount, len(eightBallResponses))
	}

	for i, response := range eightBallResponses {
		if response == "" {
			t.Errorf("Expected non-empty response at index %d", i)
		}

		// Check response length is reasonable
		if len(response) > 30 {
			t.Errorf("Response at index %d is too long (%d chars): %s",
				i, len(response), response)
		}
	}

	// Check for duplicates
	seen := make(map[string]bool)
	for i, response := range eightBallResponses {
		if seen[response] {
			t.Errorf("Duplicate response at index %d: %s", i, response)
		}
		seen[response] = true
	}
}
