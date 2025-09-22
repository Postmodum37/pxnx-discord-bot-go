package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// HandleJoinCommand handles the /join command using the simplified approach
func HandleJoinCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	// Check if simple player is initialized
	if SimplePlayer == nil {
		return respondWithInteraction(s, i, "Music system is not available")
	}

	// Find user's voice channel
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		return respondWithInteraction(s, i, "Failed to get server information")
	}

	var userChannelID string
	for _, vs := range guild.VoiceStates {
		if vs.UserID == i.Member.User.ID {
			userChannelID = vs.ChannelID
			break
		}
	}

	if userChannelID == "" {
		return respondWithInteraction(s, i, "You need to be in a voice channel first!")
	}

	// Join the voice channel
	err = SimplePlayer.JoinChannel(i.GuildID, userChannelID)
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