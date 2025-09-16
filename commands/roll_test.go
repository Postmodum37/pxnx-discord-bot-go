package commands

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"

	"pxnx-discord-bot/testutils"
)

func TestRollDice(t *testing.T) {
	tests := []struct {
		name string
		max  int
	}{
		{name: "standard roll 1-100", max: 100},
		{name: "small roll 1-6", max: 6},
		{name: "large roll 1-1000", max: 1000},
		{name: "minimum roll 1-1", max: 1},
		{name: "invalid zero max", max: 0},
		{name: "negative max", max: -5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rollDice(tt.max)

			expectedMax := tt.max
			if expectedMax < 1 {
				expectedMax = 1
			}

			if result < 1 {
				t.Errorf("Expected result >= 1, got %d", result)
			}

			if result > expectedMax {
				t.Errorf("Expected result <= %d, got %d", expectedMax, result)
			}
		})
	}
}

func TestRollDiceDistribution(t *testing.T) {
	max := 6
	iterations := 10000
	counts := make(map[int]int)

	for i := 0; i < iterations; i++ {
		result := rollDice(max)
		counts[result]++
	}

	for i := 1; i <= max; i++ {
		if counts[i] == 0 {
			t.Errorf("Expected at least one occurrence of %d in %d rolls", i, iterations)
		}

		expectedFreq := float64(iterations) / float64(max)
		actualFreq := float64(counts[i])
		deviation := (actualFreq - expectedFreq) / expectedFreq

		if deviation < -0.2 || deviation > 0.2 {
			t.Errorf("Roll %d frequency deviation too high: %.2f%% (got %d, expected ~%.0f)",
				i, deviation*100, counts[i], expectedFreq)
		}
	}
}

func TestHandleRollCommand(t *testing.T) {
	tests := []struct {
		name         string
		options      []*discordgo.ApplicationCommandInteractionDataOption
		sessionError error
		expectError  bool
	}{
		{
			name:         "default roll (no options)",
			options:      nil,
			sessionError: nil,
			expectError:  false,
		},
		{
			name: "custom max value",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Name:  "max",
					Value: int64(20),
				},
			},
			sessionError: nil,
			expectError:  false,
		},
		{
			name: "zero max value",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Name:  "max",
					Value: int64(0),
				},
			},
			sessionError: nil,
			expectError:  false,
		},
		{
			name: "negative max value",
			options: []*discordgo.ApplicationCommandInteractionDataOption{
				{
					Type:  discordgo.ApplicationCommandOptionInteger,
					Name:  "max",
					Value: int64(-10),
				},
			},
			sessionError: nil,
			expectError:  false,
		},
		{
			name:         "session error",
			options:      nil,
			sessionError: errors.New("respond error"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &testutils.MockSession{
				RespondError: tt.sessionError,
			}

			interaction := testutils.CreateTestInteraction("roll", tt.options)

			err := HandleRollCommand(mockSession, interaction)

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

			if !tt.expectError && mockSession.RespondData != nil {
				if len(mockSession.RespondData.Embeds) != 1 {
					t.Errorf("Expected 1 embed, got %d", len(mockSession.RespondData.Embeds))
				}

				embed := mockSession.RespondData.Embeds[0]

				if embed.Title != "🎲 Dice Roll" {
					t.Errorf("Expected title '🎲 Dice Roll', got '%s'", embed.Title)
				}

				if embed.Color != 0x00ff00 {
					t.Errorf("Expected color 0x00ff00, got 0x%x", embed.Color)
				}

				if embed.Description == "" {
					t.Error("Expected non-empty description")
				}

				if !strings.Contains(embed.Description, "You rolled") {
					t.Errorf("Expected description to contain 'You rolled', got '%s'", embed.Description)
				}

				expectedMax := 100
				if len(tt.options) > 0 && tt.options[0].Type == discordgo.ApplicationCommandOptionInteger {
					userMax := int(tt.options[0].IntValue())
					if userMax > 0 {
						expectedMax = userMax
					}
				}

				expectedSuffix := "(1-" + strconv.Itoa(expectedMax) + ")"
				if !strings.Contains(embed.Description, expectedSuffix) {
					t.Errorf("Expected description to contain '%s', got '%s'", expectedSuffix, embed.Description)
				}
			}
		})
	}
}