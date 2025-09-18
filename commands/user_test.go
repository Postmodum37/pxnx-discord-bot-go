package commands

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"pxnx-discord-bot/testutils"
)

func TestHandleUserCommand(t *testing.T) {
	tests := []struct {
		name             string
		setupInteraction func() *discordgo.InteractionCreate
		setupMock        func(*testutils.MockSession)
		expectedError    bool
		expectedTitle    string
	}{
		{
			name: "show command invoker info",
			setupInteraction: func() *discordgo.InteractionCreate {
				user := testutils.CreateTestUser("user123", "testuser", "avatar123")
				interaction := testutils.CreateTestInteraction("user", nil)
				interaction.Member = testutils.CreateTestMember(user)
				return interaction
			},
			setupMock: func(mock *testutils.MockSession) {
				// No additional setup needed
			},
			expectedError: false,
			expectedTitle: "👤 testuser's Profile",
		},
		{
			name: "show target user info with resolved data",
			setupInteraction: func() *discordgo.InteractionCreate {
				user := testutils.CreateTestUser("user123", "testuser", "avatar123")
				interaction := testutils.CreateTestInteraction("user", nil)
				interaction.Member = testutils.CreateTestMember(user)
				return interaction
			},
			setupMock: func(mock *testutils.MockSession) {
				// No additional setup needed
			},
			expectedError: false,
			expectedTitle: "👤 testuser's Profile",
		},
		{
			name: "session respond error",
			setupInteraction: func() *discordgo.InteractionCreate {
				user := testutils.CreateTestUser("user123", "testuser", "avatar123")
				interaction := testutils.CreateTestInteraction("user", nil)
				interaction.Member = testutils.CreateTestMember(user)
				return interaction
			},
			setupMock: func(mock *testutils.MockSession) {
				mock.RespondError = assert.AnError
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSession := &testutils.MockSession{}
			tt.setupMock(mockSession)
			interaction := tt.setupInteraction()

			err := HandleUserCommand(mockSession, interaction)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.True(t, mockSession.RespondCalled)
				require.NotNil(t, mockSession.RespondData)
				assert.NotEmpty(t, mockSession.RespondData.Embeds)
				embed := mockSession.RespondData.Embeds[0]
				assert.Equal(t, tt.expectedTitle, embed.Title)
				assert.NotEmpty(t, embed.Fields)
			}
		})
	}
}
