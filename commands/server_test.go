package commands

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pxnx-discord-bot/testutils"
)

func TestHandleServerCommand(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*testutils.MockSession)
		expectedError  bool
		expectedCalled bool
	}{
		{
			name: "successful server info response",
			setupMock: func(mock *testutils.MockSession) {
				mock.GuildReturn = testutils.CreateTestGuild("guild123", "Test Server", 250)
			},
			expectedError:  false,
			expectedCalled: true,
		},
		{
			name: "guild fetch error",
			setupMock: func(mock *testutils.MockSession) {
				mock.GuildError = assert.AnError
			},
			expectedError:  true,
			expectedCalled: false,
		},
		{
			name: "session respond error",
			setupMock: func(mock *testutils.MockSession) {
				mock.GuildReturn = testutils.CreateTestGuild("guild456", "Another Server", 500)
				mock.RespondError = assert.AnError
			},
			expectedError:  true,
			expectedCalled: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &testutils.MockSession{}
			tt.setupMock(mockSession)

			interaction := testutils.CreateTestInteraction("server", nil)
			interaction.GuildID = "guild123"

			err := HandleServerCommand(mockSession, interaction)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expectedCalled, mockSession.RespondCalled)

			if mockSession.RespondCalled && !tt.expectedError {
				require.NotNil(t, mockSession.RespondData)
				assert.NotEmpty(t, mockSession.RespondData.Embeds)
				embed := mockSession.RespondData.Embeds[0]
				assert.Contains(t, embed.Title, "Server Info")
				assert.NotEmpty(t, embed.Fields)
			}
		})
	}
}
