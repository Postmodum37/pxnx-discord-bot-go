package commands

import (
	"context"
	"fmt"
	"time"

	"pxnx-discord-bot/music/types"

	"github.com/bwmarrin/discordgo"
)

// Global music manager instance - will be initialized by the bot
var MusicManager types.MusicManager

// HandleJoinCommand handles the /join slash command
func HandleJoinCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	// Check if music manager is initialized
	if MusicManager == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Music system is not available.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get the user's voice state from session state (more reliable than API call)
	var userVoiceChannelID string
	state := s.State()
	if state != nil {
		// Try to get voice state from session state first
		guild, err := state.Guild(i.GuildID)
		if err == nil {
			for _, vs := range guild.VoiceStates {
				if vs.UserID == i.Member.User.ID {
					userVoiceChannelID = vs.ChannelID
					break
				}
			}
		}
	}

	// If not found in state, try API call as fallback
	if userVoiceChannelID == "" {
		guild, err := s.Guild(i.GuildID)
		if err != nil {
			return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "❌ Failed to get guild information.",
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}

		// Find the user's voice channel
		for _, vs := range guild.VoiceStates {
			if vs.UserID == i.Member.User.ID {
				userVoiceChannelID = vs.ChannelID
				break
			}
		}
	}

	if userVoiceChannelID == "" {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ You need to be in a voice channel to use this command.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check if bot is already connected to the same channel
	if MusicManager.IsConnected(i.GuildID) {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "✅ I'm already connected to a voice channel.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get the channel name for immediate response
	channelName := userVoiceChannelID // fallback to ID

	// Method 1: Try to get the channel directly (most reliable)
	if channel, channelErr := s.Channel(userVoiceChannelID); channelErr == nil && channel != nil {
		channelName = channel.Name
	} else {
		// Method 2: Fallback to getting it from guild channels
		if guild, guildErr := s.Guild(i.GuildID); guildErr == nil && guild != nil {
			for _, guildChannel := range guild.Channels {
				if guildChannel != nil && guildChannel.ID == userVoiceChannelID {
					channelName = guildChannel.Name
					break
				}
			}
		}
	}

	// Send immediate response indicating we're joining
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("🔄 Joining **%s**...", channelName),
		},
	})
	if err != nil {
		return err
	}

	// Create context with shorter timeout for voice connection
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Join the voice channel
	err = MusicManager.JoinChannel(ctx, i.GuildID, userVoiceChannelID)
	if err != nil {
		// Edit the message with error
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &[]string{fmt.Sprintf("❌ Failed to join **%s**: %v", channelName, err)}[0],
		})
		if editErr != nil {
			return editErr
		}
		// Return nil since we've handled the error by informing the user
		return nil
	}

	// Edit the message with success
	_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &[]string{fmt.Sprintf("🎵 Successfully joined **%s**! Ready to play some music!", channelName)}[0],
	})
	return editErr
}

// HandleLeaveCommand handles the /leave slash command
func HandleLeaveCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	// Check if music manager is initialized
	if MusicManager == nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Music system is not available.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Check if bot is connected to a voice channel
	if !MusicManager.IsConnected(i.GuildID) {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ I'm not connected to any voice channel.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Leave the voice channel
	err := MusicManager.LeaveChannel(ctx, i.GuildID)
	if err != nil {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("❌ Failed to leave voice channel: %v", err),
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "👋 Left the voice channel. Thanks for listening!",
		},
	})
}
