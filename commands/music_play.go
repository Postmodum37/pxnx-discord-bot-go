package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	"pxnx-discord-bot/music/providers"
	"pxnx-discord-bot/music/types"

	"github.com/bwmarrin/discordgo"
)

// HandlePlayCommand handles the /play slash command
func HandlePlayCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
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
				Content: "❌ I need to be in a voice channel to play music. Use `/join` first.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Get the query/URL from command options
	var query string
	if len(i.ApplicationCommandData().Options) > 0 {
		query = i.ApplicationCommandData().Options[0].StringValue()
	}

	if query == "" {
		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "❌ Please provide a song URL or search query.",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}

	// Send initial response indicating we're searching
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "🔍 Searching for music...",
		},
	})
	if err != nil {
		return err
	}

	// Create context with timeout for the search and play operation
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Try to get audio source from YouTube provider
	audioSource, err := getAudioSourceFromQuery(ctx, query)
	if err != nil {
		// Edit the message with error
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &[]string{fmt.Sprintf("❌ Failed to find audio: %v", err)}[0],
		})
		if editErr != nil {
			return editErr
		}
		return nil
	}

	// Try to play the audio
	err = MusicManager.Play(ctx, i.GuildID, *audioSource)
	if err != nil {
		// Edit the message with error
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &[]string{fmt.Sprintf("❌ Failed to play audio: %v", err)}[0],
		})
		if editErr != nil {
			return editErr
		}
		return nil
	}

	// Create success embed
	embed := &discordgo.MessageEmbed{
		Title:       "🎵 Now Playing",
		Description: fmt.Sprintf("**[%s](%s)**", audioSource.Title, audioSource.URL),
		Color:       0x1DB954, // Spotify green
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Duration",
				Value:  audioSource.Duration,
				Inline: true,
			},
			{
				Name:   "Provider",
				Value:  strings.ToUpper(string(audioSource.Provider[0])) + audioSource.Provider[1:],
				Inline: true,
			},
			{
				Name:   "Requested by",
				Value:  audioSource.RequestedBy,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /pause, /skip, or /stop to control playback",
		},
	}

	if audioSource.Thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: audioSource.Thumbnail,
		}
	}

	// Edit the message with success
	_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &[]string{""}[0], // Clear the "searching" message
		Embeds:  &[]*discordgo.MessageEmbed{embed},
	})
	return editErr
}

// getAudioSourceFromQuery gets an AudioSource from a query using YouTube provider
func getAudioSourceFromQuery(ctx context.Context, query string) (*types.AudioSource, error) {
	// For now, we'll use the YouTube provider directly since it's the only one we have
	// TODO: In the future, make this more dynamic to support multiple providers

	// Import the provider here to avoid circular dependencies
	youtubeProvider := getYouTubeProvider()
	if youtubeProvider == nil {
		return nil, fmt.Errorf("YouTube provider not available")
	}

	// Try to get audio source
	audioSource, err := youtubeProvider.GetAudioSource(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get audio source: %w", err)
	}

	if audioSource == nil {
		return nil, fmt.Errorf("no audio found for query: %s", query)
	}

	// Set the requested by field (would need to get from interaction context)
	audioSource.RequestedBy = "User" // TODO: Get actual username from interaction
	return audioSource, nil
}

// getYouTubeProvider creates a YouTube provider instance
// This is a temporary solution until we have better provider management
func getYouTubeProvider() types.AudioProvider {
	// We'll import the providers package and create a new instance
	// This avoids the need to access internal manager state
	return providers.NewYouTubeProvider()
}
