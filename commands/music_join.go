package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"pxnx-discord-bot/utils"
)

// HandleJoinCommand handles the /join command using the simplified approach
func HandleJoinCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	// Check if simple player is initialized
	if SimplePlayer == nil {
		return respondWithInteraction(s, i, "Music system is not available")
	}

	// Get the user's voice state from session state (more reliable than API call)
	var userChannelID string
	state := s.State()
	if state != nil {
		// Try to get voice state from session state first
		guild, err := state.Guild(i.GuildID)
		if err == nil {
			utils.LogDebug("Found %d voice states in session state", len(guild.VoiceStates))
			for _, vs := range guild.VoiceStates {
				utils.LogDebug("Voice state: UserID=%s, ChannelID=%s", vs.UserID, vs.ChannelID)
				if vs.UserID == i.Member.User.ID {
					userChannelID = vs.ChannelID
					utils.LogDebug("Found user %s in voice channel %s via session state", vs.UserID, vs.ChannelID)
					break
				}
			}
		} else {
			utils.LogDebug("Failed to get guild from session state: %v", err)
		}
	} else {
		utils.LogDebug("Session state is nil")
	}

	// If not found in state, try API call as fallback
	if userChannelID == "" {
		utils.LogDebug("User not found in session state, trying API call fallback")
		guild, err := s.Guild(i.GuildID)
		if err != nil {
			utils.LogDebug("Failed to get guild via API: %v", err)
			return respondWithInteraction(s, i, "Failed to get server information")
		}

		utils.LogDebug("Found %d voice states in API response", len(guild.VoiceStates))
		// Find the user's voice channel via API
		for _, vs := range guild.VoiceStates {
			utils.LogDebug("API Voice state: UserID=%s, ChannelID=%s", vs.UserID, vs.ChannelID)
			if vs.UserID == i.Member.User.ID {
				userChannelID = vs.ChannelID
				utils.LogDebug("Found user %s in voice channel %s via API", vs.UserID, vs.ChannelID)
				break
			}
		}
	}

	if userChannelID == "" {
		return respondWithInteraction(s, i, "You need to be in a voice channel first!")
	}

	// Join the voice channel
	err := SimplePlayer.JoinChannel(i.GuildID, userChannelID)
	if err != nil {
		return respondWithInteraction(s, i, fmt.Sprintf("Failed to join voice channel: %v", err))
	}

	// Get channel name for response
	channel, err := s.Channel(userChannelID)
	channelName := "voice channel"
	if err == nil {
		channelName = channel.Name
	}

	return respondWithInteraction(s, i, fmt.Sprintf("✅ Joined **%s**", channelName))
}

// HandleLeaveCommand handles the /leave command using the simplified approach
func HandleLeaveCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	// Check if simple player is initialized
	if SimplePlayer == nil {
		return respondWithInteraction(s, i, "Music system is not available")
	}

	// Leave the voice channel
	err := SimplePlayer.LeaveChannel(i.GuildID)
	if err != nil {
		return respondWithInteraction(s, i, fmt.Sprintf("Failed to leave voice channel: %v", err))
	}

	return respondWithInteraction(s, i, "👋 Left voice channel and cleared queue")
}