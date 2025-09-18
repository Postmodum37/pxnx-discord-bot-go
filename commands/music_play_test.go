package commands

import (
	"testing"

	"pxnx-discord-bot/testutils"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestHandlePlayCommand_Basic(t *testing.T) {
	t.Run("Music manager not initialized", func(t *testing.T) {
		// Reset global music manager
		originalManager := MusicManager
		defer func() { MusicManager = originalManager }()
		MusicManager = nil

		mockSession := &testutils.MockSession{}
		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				GuildID: "guild123",
			},
		}

		err := HandlePlayCommand(mockSession, interaction)

		assert.NoError(t, err)
		assert.True(t, mockSession.RespondCalled)
		assert.Equal(t, discordgo.MessageFlagsEphemeral, mockSession.RespondData.Flags)
		assert.Contains(t, mockSession.RespondData.Content, "❌ Music system is not available.")
	})

	t.Run("Bot not connected to voice", func(t *testing.T) {
		// Reset global music manager
		originalManager := MusicManager
		defer func() { MusicManager = originalManager }()

		mockSession := &testutils.MockSession{}
		mockManager := &testutils.MockMusicManager{}
		mockManager.IsConnectedReturn = false
		MusicManager = mockManager

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				GuildID: "guild123",
			},
		}

		err := HandlePlayCommand(mockSession, interaction)

		assert.NoError(t, err)
		assert.True(t, mockSession.RespondCalled)
		assert.Equal(t, discordgo.MessageFlagsEphemeral, mockSession.RespondData.Flags)
		assert.Contains(t, mockSession.RespondData.Content, "❌ I need to be in a voice channel")
		assert.True(t, mockManager.IsConnectedCalled)
	})
}
