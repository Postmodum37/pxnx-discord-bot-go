package commands

import (
	"fmt"
	"pxnx-discord-bot/music"

	"github.com/bwmarrin/discordgo"
)

// SimplePlayer is a global instance of the simplified music player
var SimplePlayer *music.SimplePlayer

// InitializeSimplePlayer initializes the global simple player
func InitializeSimplePlayer(session *discordgo.Session) {
	SimplePlayer = music.NewSimplePlayer(session)
}

// HandlePlayCommand handles the /play slash command using the simplified approach
func HandlePlayCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	// Defer response to avoid timeout
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
	if err != nil {
		return fmt.Errorf("failed to defer response: %w", err)
	}

	// Check if simple player is initialized
	if SimplePlayer == nil {
		return respondWithError(s, i, "Music system is not available")
	}

	// Get the query from command options
	var query string
	if len(i.ApplicationCommandData().Options) > 0 {
		query = i.ApplicationCommandData().Options[0].StringValue()
	}

	if query == "" {
		return respondWithError(s, i, "Please provide a song name or YouTube URL")
	}

	// Check if bot is connected to a voice channel
	player, connected := SimplePlayer.GetPlayer(i.GuildID)
	if !connected {
		return respondWithError(s, i, "I need to be in a voice channel first. Use `/join` command")
	}

	// Send searching status
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &[]string{"🔍 Searching for music..."}[0],
	})
	if err != nil {
		return fmt.Errorf("failed to update response: %w", err)
	}

	// Try to play the track
	track, err := SimplePlayer.Play(i.GuildID, query)
	if err != nil {
		return respondWithError(s, i, fmt.Sprintf("Failed to play music: %v", err))
	}

	// Create success response
	var content string
	var embed *discordgo.MessageEmbed

	if player.IsPlaying() {
		// Currently playing - added to queue
		queuePosition := len(player.GetQueue())
		content = fmt.Sprintf("🎵 Added to queue (position %d)", queuePosition)
		embed = createTrackEmbed(track, "Added to Queue", 0x3498db, i.Member.User) // Blue
	} else {
		// Started playing immediately
		content = "🎵 Now playing"
		embed = createTrackEmbed(track, "Now Playing", 0x1db954, i.Member.User) // Spotify green
	}

	// Edit the response with success
	_, err = s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &content,
		Embeds:  &[]*discordgo.MessageEmbed{embed},
	})

	return err
}

// Helper functions

func createTrackEmbed(track *music.AudioTrack, title string, color int, requestedBy *discordgo.User) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       title,
		Description: fmt.Sprintf("**[%s](%s)**", track.Title, track.URL),
		Color:       color,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Duration",
				Value:  track.Duration,
				Inline: true,
			},
			{
				Name:   "Provider",
				Value:  "Youtube",
				Inline: true,
			},
			{
				Name:   "Requested by",
				Value:  requestedBy.Username,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Use /pause, /skip, or /stop to control playback",
		},
	}

	// Add thumbnail if available
	if track.Thumbnail != "" {
		embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
			URL: track.Thumbnail,
		}
	}

	return embed
}

func respondWithError(s SessionInterface, i *discordgo.InteractionCreate, message string) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &[]string{fmt.Sprintf("❌ %s", message)}[0],
	})
	return err
}

func respondWithInteraction(s SessionInterface, i *discordgo.InteractionCreate, message string) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: message,
		},
	})
}