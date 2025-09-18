package commands

import (
	"errors"
	"testing"

	"pxnx-discord-bot/testutils"
)

func TestHandlePingCommand(t *testing.T) {
	tests := []struct {
		name          string
		sessionError  error
		expectError   bool
		expectContent string
	}{
		{
			name:          "successful ping",
			sessionError:  nil,
			expectError:   false,
			expectContent: "Pong!",
		},
		{
			name:         "session error",
			sessionError: errors.New("session error"),
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &testutils.MockSession{
				RespondError: tt.sessionError,
			}

			interaction := testutils.CreateTestInteraction("ping", nil)

			err := HandlePingCommand(mockSession, interaction)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				if err != tt.sessionError {
					t.Errorf("Expected error %v, got %v", tt.sessionError, err)
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
				if mockSession.RespondData.Content != tt.expectContent {
					t.Errorf("Expected content '%s', got '%s'",
						tt.expectContent, mockSession.RespondData.Content)
				}
			}
		})
	}
}
