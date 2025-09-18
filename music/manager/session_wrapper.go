package manager

import (
	"github.com/bwmarrin/discordgo"
)

// SessionWrapper wraps a discordgo.Session to implement our SessionInterface
type SessionWrapper struct {
	session *discordgo.Session
}

// NewSessionWrapper creates a new session wrapper
func NewSessionWrapper(session *discordgo.Session) *SessionWrapper {
	return &SessionWrapper{
		session: session,
	}
}

// InteractionRespond wraps the session's InteractionRespond method
func (sw *SessionWrapper) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	return sw.session.InteractionRespond(interaction, resp, options...)
}

// InteractionResponseEdit wraps the session's InteractionResponseEdit method
func (sw *SessionWrapper) InteractionResponseEdit(interaction *discordgo.Interaction, newresp *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return sw.session.InteractionResponseEdit(interaction, newresp, options...)
}

// Guild wraps the session's Guild method
func (sw *SessionWrapper) Guild(guildID string, options ...discordgo.RequestOption) (*discordgo.Guild, error) {
	return sw.session.Guild(guildID, options...)
}

// Channel wraps the session's Channel method
func (sw *SessionWrapper) Channel(channelID string, options ...discordgo.RequestOption) (*discordgo.Channel, error) {
	return sw.session.Channel(channelID, options...)
}

// ChannelVoiceJoin wraps the session's ChannelVoiceJoin method
func (sw *SessionWrapper) ChannelVoiceJoin(guildID, channelID string, mute, deaf bool) (*discordgo.VoiceConnection, error) {
	return sw.session.ChannelVoiceJoin(guildID, channelID, mute, deaf)
}

// FollowupMessageCreate wraps the session's FollowupMessageCreate method
func (sw *SessionWrapper) FollowupMessageCreate(interaction *discordgo.Interaction, wait bool, data *discordgo.WebhookParams, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return sw.session.FollowupMessageCreate(interaction, wait, data, options...)
}

// GetVoiceConnection accesses the voice connection from the session's VoiceConnections map
func (sw *SessionWrapper) GetVoiceConnection(guildID string) *discordgo.VoiceConnection {
	return sw.session.VoiceConnections[guildID]
}

// State returns the session's state
func (sw *SessionWrapper) State() *discordgo.State {
	return sw.session.State
}
