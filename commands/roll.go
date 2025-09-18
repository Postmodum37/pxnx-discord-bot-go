package commands

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

var rollRng = rand.New(rand.NewSource(time.Now().UnixNano()))

func rollDice(max int) int {
	if max < 1 {
		max = 1
	}
	return rollRng.Intn(max) + 1
}

func HandleRollCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	max := 100

	if len(i.ApplicationCommandData().Options) > 0 {
		if i.ApplicationCommandData().Options[0].Type == discordgo.ApplicationCommandOptionInteger {
			userMax := int(i.ApplicationCommandData().Options[0].IntValue())
			if userMax > 0 {
				max = userMax
			}
		}
	}

	result := rollDice(max)

	embed := &discordgo.MessageEmbed{
		Title:       "🎲 Dice Roll",
		Description: fmt.Sprintf("You rolled **%d** (1-%d)", result, max),
		Color:       0x00ff00,
	}

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
