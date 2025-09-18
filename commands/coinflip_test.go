package commands

import (
	"errors"
	"strings"
	"testing"

	"pxnx-discord-bot/testutils"
)

func TestHandleCoinFlipCommand(t *testing.T) {
	tests := []struct {
		name         string
		sessionError error
		expectError  bool
	}{
		{
			name:         "successful coin flip",
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

			interaction := testutils.CreateTestInteraction("coinflip", nil)

			err := HandleCoinFlipCommand(mockSession, interaction)

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
				if embed.Title != "🪙 Coin Flip" {
					t.Errorf("Expected title '🪙 Coin Flip', got '%s'", embed.Title)
				}

				if embed.Color != 0xf39c12 {
					t.Errorf("Expected color 0xf39c12, got 0x%x", embed.Color)
				}

				// Check that result is either Heads or Tails
				description := embed.Description
				if !strings.Contains(description, "Heads") && !strings.Contains(description, "Tails") {
					t.Errorf("Expected description to contain 'Heads' or 'Tails', got '%s'", description)
				}

				// Check format
				if !strings.Contains(description, "The coin landed on") {
					t.Errorf("Expected description to contain 'The coin landed on', got '%s'", description)
				}
			}
		})
	}
}

func TestCoinFlipDistribution(t *testing.T) {
	// Test that coin flip results have some randomness
	headsCount := 0
	tailsCount := 0
	iterations := 100

	for i := 0; i < iterations; i++ {
		mockSession := &testutils.MockSession{}
		interaction := testutils.CreateTestInteraction("coinflip", nil)

		err := HandleCoinFlipCommand(mockSession, interaction)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		if mockSession.RespondData != nil && len(mockSession.RespondData.Embeds) > 0 {
			description := mockSession.RespondData.Embeds[0].Description
			if strings.Contains(description, "Heads") {
				headsCount++
			} else if strings.Contains(description, "Tails") {
				tailsCount++
			}
		}
	}

	// Both heads and tails should occur at least once in 100 iterations
	if headsCount == 0 {
		t.Error("Expected at least one Heads result in 100 iterations")
	}
	if tailsCount == 0 {
		t.Error("Expected at least one Tails result in 100 iterations")
	}

	totalResults := headsCount + tailsCount
	if totalResults != iterations {
		t.Errorf("Expected %d total results, got %d", iterations, totalResults)
	}

	// Results should be somewhat balanced (allow for randomness)
	if headsCount < 20 || tailsCount < 20 {
		t.Errorf("Distribution seems skewed: Heads=%d, Tails=%d", headsCount, tailsCount)
	}
}
