package commands

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// HandleCoinFlipCommand handles the coinflip slash command
func HandleCoinFlipCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	result := "Heads"
	if rng.Intn(2) == 1 {
		result = "Tails"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🪙 Coin Flip",
		Description: fmt.Sprintf("The coin landed on **%s**!", result),
		Color:       0xf39c12, // ColorOrange
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
