package commands

import (
	"github.com/bwmarrin/discordgo"
)

var eightBallResponses = []string{
	"It is certain",
	"It is decidedly so",
	"Without a doubt",
	"Yes definitely",
	"You may rely on it",
	"As I see it, yes",
	"Most likely",
	"Outlook good",
	"Yes",
	"Signs point to yes",
	"Reply hazy, try again",
	"Ask again later",
	"Better not tell you now",
	"Cannot predict now",
	"Concentrate and ask again",
	"Don't count on it",
	"My reply is no",
	"My sources say no",
	"Outlook not so good",
	"Very doubtful",
}

// get8BallResponse returns a random 8-ball response
func get8BallResponse() string {
	return eightBallResponses[rng.Intn(len(eightBallResponses))]
}

// Handle8BallCommand handles the 8ball slash command
func Handle8BallCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	question := options[0].StringValue()
	response := get8BallResponse()

	embed := &discordgo.MessageEmbed{
		Title: "🎱 Magic 8-Ball",
		Color: 0x9b59b6, // ColorPurple
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Question",
				Value: question,
			},
			{
				Name:  "Answer",
				Value: response,
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
