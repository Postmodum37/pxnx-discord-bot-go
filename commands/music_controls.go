package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// HandleStopCommand handles the /stop command using the simplified approach
func HandleStopCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	if SimplePlayer == nil {
		return respondWithInteraction(s, i, "Music system is not available")
	}

	player, connected := SimplePlayer.GetPlayer(i.GuildID)
	if !connected {
		return respondWithInteraction(s, i, "Not connected to a voice channel")
	}

	player.Stop()
	return respondWithInteraction(s, i, "⏹️ Stopped playback and cleared queue")
}

// HandleSkipCommand handles the /skip command using the simplified approach
func HandleSkipCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	if SimplePlayer == nil {
		return respondWithInteraction(s, i, "Music system is not available")
	}

	player, connected := SimplePlayer.GetPlayer(i.GuildID)
	if !connected {
		return respondWithInteraction(s, i, "Not connected to a voice channel")
	}

	if !player.IsPlaying() {
		return respondWithInteraction(s, i, "Nothing is currently playing")
	}

	player.Skip()

	queue := player.GetQueue()
	if len(queue) > 0 {
		return respondWithInteraction(s, i, fmt.Sprintf("⏭️ Skipped! Playing next track... (%d songs remaining)", len(queue)))
	} else {
		return respondWithInteraction(s, i, "⏭️ Skipped! No more songs in queue")
	}
}

// HandleQueueCommand handles the /queue command using the simplified approach
func HandleQueueCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	if SimplePlayer == nil {
		return respondWithInteraction(s, i, "Music system is not available")
	}

	player, connected := SimplePlayer.GetPlayer(i.GuildID)
	if !connected {
		return respondWithInteraction(s, i, "Not connected to a voice channel")
	}

	current := player.GetCurrent()
	queue := player.GetQueue()

	embed := &discordgo.MessageEmbed{
		Title: "🎵 Music Queue",
		Color: 0x3498db,
	}

	if current != nil {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "Now Playing",
			Value:  fmt.Sprintf("🎶 **%s**", current.Title),
			Inline: false,
		})
	}

	if len(queue) == 0 {
		embed.Description = "Queue is empty"
	} else {
		queueText := ""
		for i, track := range queue {
			if i >= 10 { // Limit to 10 tracks
				queueText += fmt.Sprintf("... and %d more tracks\n", len(queue)-10)
				break
			}
			queueText += fmt.Sprintf("%d. **%s**\n", i+1, track.Title)
		}
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:  fmt.Sprintf("Up Next (%d songs)", len(queue)),
			Value: queueText,
		})
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}