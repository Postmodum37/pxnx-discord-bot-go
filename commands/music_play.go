package commands

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"pxnx-discord-bot/music/types"

	"github.com/bwmarrin/discordgo"
)

// HandlePlayCommand handles the /play slash command
func HandlePlayCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	// Immediately defer the response to avoid timeout
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return err
	}

	// Check if music manager is initialized
	if MusicManager == nil {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &[]string{"❌ Music system is not available."}[0],
		})
		return editErr
	}

	// Get the query/URL from command options
	var query string
	if len(i.ApplicationCommandData().Options) > 0 {
		query = i.ApplicationCommandData().Options[0].StringValue()
	}

	if query == "" {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &[]string{"❌ Please provide a song URL or search query."}[0],
		})
		return editErr
	}

	// Check if bot is connected to a voice channel
	if !MusicManager.IsConnected(i.GuildID) {
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &[]string{"❌ I need to be in a voice channel to play music. Use `/join` first."}[0],
		})
		return editErr
	}

	// Send searching status
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &[]string{"🔍 Searching for music..."}[0],
	})
	if err != nil {
		return err
	}

	// Create context with timeout ONLY for the search operation
	searchCtx, searchCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer searchCancel()

	// Try to get audio source from YouTube provider
	log.Printf("[MUSIC] Getting audio source for query: %s", query)
	audioSource, err := getAudioSourceFromQuery(searchCtx, query)
	if err != nil {
		log.Printf("[MUSIC] Failed to get audio source: %v", err)
		// Edit the message with error
		_, editErr := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Content: &[]string{fmt.Sprintf("❌ Failed to find audio: %v", err)}[0],
		})
		if editErr != nil {
			return editErr
		}
		return nil
	}
	log.Printf("[MUSIC] Successfully got audio source: %s", audioSource.Title)

	// Create a separate long-running context for playback (no timeout)
	playCtx := context.Background()
	log.Printf("[MUSIC] Starting playback with background context")

	// Try to play the audio
	err = MusicManager.Play(playCtx, i.GuildID, *audioSource)
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

// getAudioSourceFromQuery gets an AudioSource from a query using the first provider
func getAudioSourceFromQuery(ctx context.Context, query string) (*types.AudioSource, error) {
	// Get providers from the music manager
	if MusicManager == nil {
		return nil, fmt.Errorf("music manager is not initialized")
	}

	log.Printf("[MUSIC] Getting providers from music manager")
	providers := MusicManager.GetProviders()
	if len(providers) == 0 {
		return nil, fmt.Errorf("no audio providers available")
	}

	log.Printf("[MUSIC] Found %d providers", len(providers))

	// Try each provider until we find one that works
	var lastErr error
	for i, provider := range providers {
		log.Printf("[MUSIC] Trying provider %d: %s", i+1, provider.GetProviderName())

		// Try to get audio source
		log.Printf("[MUSIC] Searching for: %s", query)
		audioSource, err := provider.GetAudioSource(ctx, query)
		if err != nil {
			log.Printf("[MUSIC] Provider failed: %v", err)
			lastErr = err
			continue
		}

		if audioSource == nil {
			log.Printf("[MUSIC] Provider returned nil audio source")
			lastErr = fmt.Errorf("no audio found for query: %s", query)
			continue
		}

		log.Printf("[MUSIC] Successfully found audio: %s", audioSource.Title)
		// Set the requested by field (TODO: Get actual username from interaction)
		audioSource.RequestedBy = "User"
		return audioSource, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("no suitable provider found for query: %s", query)
}
