package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// HandleUserCommand handles the user slash command
func HandleUserCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	var targetUser *discordgo.User

	// Check if a user was mentioned in options
	if len(i.ApplicationCommandData().Options) > 0 {
		targetUser = i.ApplicationCommandData().Options[0].UserValue(nil)
	} else {
		// Use the command invoker
		targetUser = i.Member.User
	}

	avatarURL := getUserAvatarURL(targetUser)
	userCreated, _ := discordgo.SnowflakeTimestamp(targetUser.ID)

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("👤 %s's Profile", targetUser.Username),
		Description: fmt.Sprintf("Here's some information about **%s**", targetUser.Mention()),
		Color:       0xe74c3c, // ColorRed
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: avatarURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🏷️ Username",
				Value:  targetUser.Username,
				Inline: true,
			},
			{
				Name:   "🆔 User ID",
				Value:  targetUser.ID,
				Inline: true,
			},
			{
				Name:   "🗓️ Account Created",
				Value:  fmt.Sprintf("<t:%d:F>", userCreated.Unix()),
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