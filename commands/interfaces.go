package commands

import (
	"github.com/bwmarrin/discordgo"
)

// SessionInterface defines the methods we need from a Discord session for testing
// This interface covers basic Discord functionality used by most commands
type SessionInterface interface {
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
	InteractionResponseEdit(interaction *discordgo.Interaction, newresp *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error)
	FollowupMessageCreate(interaction *discordgo.Interaction, wait bool, data *discordgo.WebhookParams, options ...discordgo.RequestOption) (*discordgo.Message, error)
	Guild(guildID string, options ...discordgo.RequestOption) (*discordgo.Guild, error)
	Channel(channelID string, options ...discordgo.RequestOption) (*discordgo.Channel, error)
	// Access to session state for voice channel detection
	State() *discordgo.State
}
