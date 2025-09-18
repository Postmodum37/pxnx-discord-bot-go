package commands

import (
	"errors"
	"testing"

	"pxnx-discord-bot/testutils"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestHandleJoinCommand(t *testing.T) {
	tests := []struct {
		name              string
		setupMocks        func(*testutils.MockSession, *testutils.MockMusicManager)
		interaction       *discordgo.InteractionCreate
		expectedResponse  string
		expectedEphemeral bool
		expectManagerCall bool
		expectFollowup    bool
	}{
		{
			name: "Music manager not initialized",
			setupMocks: func(session *testutils.MockSession, manager *testutils.MockMusicManager) {
				// Don't set up manager (will be nil)
			},
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID: "guild123",
					Member: &discordgo.Member{
						User: &discordgo.User{ID: "user123"},
					},
				},
			},
			expectedResponse:  "❌ Music system is not available.",
			expectedEphemeral: true,
			expectManagerCall: false,
			expectFollowup:    false,
		},
		{
			name: "User not in voice channel",
			setupMocks: func(session *testutils.MockSession, manager *testutils.MockMusicManager) {
				MusicManager = manager
				// Create guild without user in voice states
				session.GuildReturn = &discordgo.Guild{
					ID: "guild123",
					VoiceStates: []*discordgo.VoiceState{
						{UserID: "otheruser", ChannelID: "voice123"},
					},
				}
			},
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID: "guild123",
					Member: &discordgo.Member{
						User: &discordgo.User{ID: "user123"},
					},
				},
			},
			expectedResponse:  "❌ You need to be in a voice channel to use this command.",
			expectedEphemeral: true,
			expectManagerCall: false,
			expectFollowup:    false,
		},
		{
			name: "Bot already connected",
			setupMocks: func(session *testutils.MockSession, manager *testutils.MockMusicManager) {
				MusicManager = manager
				manager.IsConnectedReturn = true
				session.GuildReturn = &discordgo.Guild{
					ID: "guild123",
					VoiceStates: []*discordgo.VoiceState{
						{UserID: "user123", ChannelID: "voice123"},
					},
				}
			},
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID: "guild123",
					Member: &discordgo.Member{
						User: &discordgo.User{ID: "user123"},
					},
				},
			},
			expectedResponse:  "✅ I'm already connected to a voice channel.",
			expectedEphemeral: true,
			expectManagerCall: false,
			expectFollowup:    false,
		},
		{
			name: "Successful join",
			setupMocks: func(session *testutils.MockSession, manager *testutils.MockMusicManager) {
				MusicManager = manager
				manager.IsConnectedReturn = false
				guild := &discordgo.Guild{
					ID: "guild123",
					Channels: []*discordgo.Channel{
						{ID: "voice123", Name: "General", Type: discordgo.ChannelTypeGuildVoice},
					},
					VoiceStates: []*discordgo.VoiceState{
						{UserID: "user123", ChannelID: "voice123"},
					},
				}
				session.GuildReturn = guild
				// Also set up state for channel name lookup
				session.StateReturn = &discordgo.State{
					Ready: discordgo.Ready{
						Guilds: []*discordgo.Guild{guild},
					},
				}
				// Set up direct channel lookup (primary method)
				session.ChannelReturn = &discordgo.Channel{
					ID:   "voice123",
					Name: "General",
					Type: discordgo.ChannelTypeGuildVoice,
				}
			},
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID: "guild123",
					Member: &discordgo.Member{
						User: &discordgo.User{ID: "user123"},
					},
				},
			},
			expectedResponse:  "🔄 Joining **General**...",
			expectedEphemeral: false,
			expectManagerCall: true,
			expectFollowup:    false,
		},
		{
			name: "Join fails",
			setupMocks: func(session *testutils.MockSession, manager *testutils.MockMusicManager) {
				MusicManager = manager
				manager.IsConnectedReturn = false
				manager.JoinChannelError = errors.New("connection failed")
				session.GuildReturn = &discordgo.Guild{
					ID: "guild123",
					VoiceStates: []*discordgo.VoiceState{
						{UserID: "user123", ChannelID: "voice123"},
					},
				}
			},
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID: "guild123",
					Member: &discordgo.Member{
						User: &discordgo.User{ID: "user123"},
					},
				},
			},
			expectedResponse:  "🔄 Joining",
			expectedEphemeral: false,
			expectManagerCall: true,
			expectFollowup:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global music manager
			originalManager := MusicManager
			defer func() { MusicManager = originalManager }()

			mockSession := &testutils.MockSession{}
			mockManager := &testutils.MockMusicManager{}
			tt.setupMocks(mockSession, mockManager)

			err := HandleJoinCommand(mockSession, tt.interaction)

			assert.NoError(t, err)
			assert.True(t, mockSession.RespondCalled)

			if tt.expectFollowup {
				// Should have deferred response
				assert.Equal(t, discordgo.InteractionResponseDeferredChannelMessageWithSource,
					mockSession.RespondType)
				assert.True(t, mockSession.FollowupCalled)
			} else {
				// Should have immediate response
				assert.Contains(t, mockSession.RespondData.Content, tt.expectedResponse)
				if tt.expectedEphemeral {
					assert.Equal(t, discordgo.MessageFlagsEphemeral, mockSession.RespondData.Flags)
				}
			}

			if tt.expectManagerCall {
				assert.True(t, mockManager.JoinChannelCalled)
			}
		})
	}
}

func TestHandleLeaveCommand(t *testing.T) {
	tests := []struct {
		name              string
		setupMocks        func(*testutils.MockSession, *testutils.MockMusicManager)
		interaction       *discordgo.InteractionCreate
		expectedResponse  string
		expectedEphemeral bool
		expectManagerCall bool
	}{
		{
			name: "Music manager not initialized",
			setupMocks: func(session *testutils.MockSession, manager *testutils.MockMusicManager) {
				// Don't set up manager (will be nil)
			},
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID: "guild123",
				},
			},
			expectedResponse:  "❌ Music system is not available.",
			expectedEphemeral: true,
			expectManagerCall: false,
		},
		{
			name: "Bot not connected",
			setupMocks: func(session *testutils.MockSession, manager *testutils.MockMusicManager) {
				MusicManager = manager
				manager.IsConnectedReturn = false
			},
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID: "guild123",
				},
			},
			expectedResponse:  "❌ I'm not connected to any voice channel.",
			expectedEphemeral: true,
			expectManagerCall: false,
		},
		{
			name: "Successful leave",
			setupMocks: func(session *testutils.MockSession, manager *testutils.MockMusicManager) {
				MusicManager = manager
				manager.IsConnectedReturn = true
			},
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID: "guild123",
				},
			},
			expectedResponse:  "👋 Left the voice channel. Thanks for listening!",
			expectedEphemeral: false,
			expectManagerCall: true,
		},
		{
			name: "Leave fails",
			setupMocks: func(session *testutils.MockSession, manager *testutils.MockMusicManager) {
				MusicManager = manager
				manager.IsConnectedReturn = true
				manager.LeaveChannelError = errors.New("disconnect failed")
			},
			interaction: &discordgo.InteractionCreate{
				Interaction: &discordgo.Interaction{
					GuildID: "guild123",
				},
			},
			expectedResponse:  "❌ Failed to leave voice channel: disconnect failed",
			expectedEphemeral: true,
			expectManagerCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset global music manager
			originalManager := MusicManager
			defer func() { MusicManager = originalManager }()

			mockSession := &testutils.MockSession{}
			mockManager := &testutils.MockMusicManager{}
			tt.setupMocks(mockSession, mockManager)

			err := HandleLeaveCommand(mockSession, tt.interaction)

			assert.NoError(t, err)
			assert.True(t, mockSession.RespondCalled)
			assert.Contains(t, mockSession.RespondData.Content, tt.expectedResponse)

			if tt.expectedEphemeral {
				assert.Equal(t, discordgo.MessageFlagsEphemeral, mockSession.RespondData.Flags)
			}

			if tt.expectManagerCall {
				assert.True(t, mockManager.LeaveChannelCalled)
			}
		})
	}
}

func TestHandleJoinCommand_VoiceChannelDetection(t *testing.T) {
	// Reset global music manager
	originalManager := MusicManager
	defer func() { MusicManager = originalManager }()

	t.Run("User in voice channel via session state", func(t *testing.T) {
		mockSession := &testutils.MockSession{}
		mockManager := &testutils.MockMusicManager{
			IsConnectedReturn: false,
		}
		MusicManager = mockManager

		// Create mock state with user in voice channel
		mockGuild := &discordgo.Guild{
			ID: "guild123",
			VoiceStates: []*discordgo.VoiceState{
				{
					UserID:    "user456",
					ChannelID: "voice789",
					GuildID:   "guild123",
				},
			},
			Channels: []*discordgo.Channel{
				{
					ID:   "voice789",
					Name: "General",
					Type: discordgo.ChannelTypeGuildVoice,
				},
			},
		}

		mockState := &discordgo.State{
			Ready: discordgo.Ready{
				Guilds: []*discordgo.Guild{mockGuild},
			},
		}
		mockSession.StateReturn = mockState
		// Also set the guild return for fallback scenarios
		mockSession.GuildReturn = mockGuild
		// Set up direct channel lookup (primary method)
		mockSession.ChannelReturn = &discordgo.Channel{
			ID:   "voice789",
			Name: "General",
			Type: discordgo.ChannelTypeGuildVoice,
		}

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				GuildID: "guild123",
				Member: &discordgo.Member{
					User: &discordgo.User{
						ID: "user456",
					},
				},
			},
		}

		err := HandleJoinCommand(mockSession, interaction)

		assert.NoError(t, err)
		assert.True(t, mockSession.RespondCalled)
		assert.Equal(t, discordgo.InteractionResponseChannelMessageWithSource, mockSession.RespondType)
		assert.True(t, mockManager.JoinChannelCalled)
		assert.Equal(t, "guild123", mockManager.JoinChannelGuildID)
		assert.Equal(t, "voice789", mockManager.JoinChannelChannelID)
		assert.True(t, mockSession.InteractionResponseEditCalled)
	})

	t.Run("User in voice channel via API fallback when state is nil", func(t *testing.T) {
		mockSession := &testutils.MockSession{}
		mockManager := &testutils.MockMusicManager{
			IsConnectedReturn: false,
		}
		MusicManager = mockManager

		// Set state to nil to trigger API fallback
		mockSession.StateReturn = nil

		// Set guild return for API fallback
		mockSession.GuildReturn = &discordgo.Guild{
			ID: "guild123",
			VoiceStates: []*discordgo.VoiceState{
				{
					UserID:    "user456",
					ChannelID: "voice789",
					GuildID:   "guild123",
				},
			},
			Channels: []*discordgo.Channel{
				{
					ID:   "voice789",
					Name: "Music Room",
					Type: discordgo.ChannelTypeGuildVoice,
				},
			},
		}
		// Set up direct channel lookup (primary method)
		mockSession.ChannelReturn = &discordgo.Channel{
			ID:   "voice789",
			Name: "Music Room",
			Type: discordgo.ChannelTypeGuildVoice,
		}

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				GuildID: "guild123",
				Member: &discordgo.Member{
					User: &discordgo.User{
						ID: "user456",
					},
				},
			},
		}

		err := HandleJoinCommand(mockSession, interaction)

		assert.NoError(t, err)
		assert.True(t, mockSession.RespondCalled)
		assert.Equal(t, discordgo.InteractionResponseChannelMessageWithSource, mockSession.RespondType)
		assert.True(t, mockManager.JoinChannelCalled)
		assert.Equal(t, "guild123", mockManager.JoinChannelGuildID)
		assert.Equal(t, "voice789", mockManager.JoinChannelChannelID)
		assert.True(t, mockSession.InteractionResponseEditCalled)
		assert.True(t, mockSession.GuildCalled) // Should have called Guild as fallback
	})

	t.Run("User not in voice channel with empty state", func(t *testing.T) {
		mockSession := &testutils.MockSession{}
		mockManager := &testutils.MockMusicManager{
			IsConnectedReturn: false,
		}
		MusicManager = mockManager

		// Create mock state with NO voice states for our user
		mockGuild := &discordgo.Guild{
			ID:          "guild123",
			VoiceStates: []*discordgo.VoiceState{}, // Empty - user not in voice
		}

		mockState := &discordgo.State{
			Ready: discordgo.Ready{
				Guilds: []*discordgo.Guild{mockGuild},
			},
		}
		mockSession.StateReturn = mockState

		// Also set Guild fallback to empty
		mockSession.GuildReturn = mockGuild

		interaction := &discordgo.InteractionCreate{
			Interaction: &discordgo.Interaction{
				GuildID: "guild123",
				Member: &discordgo.Member{
					User: &discordgo.User{
						ID: "user456",
					},
				},
			},
		}

		err := HandleJoinCommand(mockSession, interaction)

		assert.NoError(t, err)
		assert.True(t, mockSession.RespondCalled)
		assert.Equal(t, discordgo.InteractionResponseChannelMessageWithSource, mockSession.RespondType)
		assert.Contains(t, mockSession.RespondData.Content, "❌ You need to be in a voice channel")
		assert.False(t, mockManager.JoinChannelCalled)
	})
}
