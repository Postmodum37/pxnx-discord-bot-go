package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// HandleServerCommand handles the server slash command
func HandleServerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		return err
	}

	memberCount := guild.MemberCount
	createdAt, _ := discordgo.SnowflakeTimestamp(guild.ID)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📊 %s Server Info", guild.Name),
		Description: fmt.Sprintf("Here's some information about **%s**", guild.Name),
		Color:       0x2ecc71, // ColorGreen
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: guild.IconURL("256"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "👥 Members",
				Value:  fmt.Sprintf("%d", memberCount),
				Inline: true,
			},
			{
				Name:   "🆔 Server ID",
				Value:  guild.ID,
				Inline: true,
			},
			{
				Name:   "👑 Owner",
				Value:  fmt.Sprintf("<@%s>", guild.OwnerID),
				Inline: true,
			},
			{
				Name:   "🗓️ Created",
				Value:  fmt.Sprintf("<t:%d:F>", createdAt.Unix()),
				Inline: false,
			},
		},
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}