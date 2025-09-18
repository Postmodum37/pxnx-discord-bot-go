package testutils

import (
	"github.com/bwmarrin/discordgo"
)

// MockSession implements a mock Discord session for testing
type MockSession struct {
	RespondCalled                 bool
	RespondError                  error
	RespondData                   *discordgo.InteractionResponseData
	RespondType                   discordgo.InteractionResponseType
	InteractionResponseEditCalled bool
	InteractionResponseEditError  error
	InteractionResponseEditReturn *discordgo.Message
	FollowupCalled                bool
	FollowupError                 error
	FollowupReturn                *discordgo.Message
	GuildCalled                   bool
	GuildError                    error
	GuildReturn                   *discordgo.Guild
	ChannelCalled                 bool
	ChannelError                  error
	ChannelReturn                 *discordgo.Channel
	StateCalled                   bool
	StateReturn                   *discordgo.State
	GuildEmojisCalled             bool
	GuildEmojisError              error
	GuildEmojisReturn             []*discordgo.Emoji
	InteractionResponseCalled     bool
	InteractionResponseError      error
	InteractionResponseReturn     *discordgo.Message
	MessageReactionAddCalled      bool
	MessageReactionAddError       error
}

// InteractionRespond mocks the Discord session InteractionRespond method
func (m *MockSession) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	m.RespondCalled = true
	m.RespondType = resp.Type
	if resp.Data != nil {
		m.RespondData = resp.Data
	}
	return m.RespondError
}

// InteractionResponseEdit mocks the Discord session InteractionResponseEdit method
func (m *MockSession) InteractionResponseEdit(interaction *discordgo.Interaction, newresp *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.InteractionResponseEditCalled = true
	if m.InteractionResponseEditError != nil {
		return nil, m.InteractionResponseEditError
	}
	return m.InteractionResponseEditReturn, nil
}

// FollowupMessageCreate mocks the Discord session FollowupMessageCreate method
func (m *MockSession) FollowupMessageCreate(interaction *discordgo.Interaction, wait bool, data *discordgo.WebhookParams, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	m.FollowupCalled = true
	if m.FollowupError != nil {
		return nil, m.FollowupError
	}
	return m.FollowupReturn, nil
}

// Guild mocks the Discord session Guild method
func (m *MockSession) Guild(guildID string, options ...discordgo.RequestOption) (*discordgo.Guild, error) {
	m.GuildCalled = true
	if m.GuildError != nil {
		return nil, m.GuildError
	}
	return m.GuildReturn, nil
}

// Channel mocks the Discord session Channel method
func (m *MockSession) Channel(channelID string, options ...discordgo.RequestOption) (*discordgo.Channel, error) {
	m.ChannelCalled = true
	if m.ChannelError != nil {
		return nil, m.ChannelError
	}
	return m.ChannelReturn, nil
}

// GuildEmojis mocks the Discord session GuildEmojis method
func (m *MockSession) GuildEmojis(guildID string) ([]*discordgo.Emoji, error) {
	m.GuildEmojisCalled = true
	if m.GuildEmojisError != nil {
		return nil, m.GuildEmojisError
	}
	return m.GuildEmojisReturn, nil
}

// InteractionResponse mocks the Discord session InteractionResponse method
func (m *MockSession) InteractionResponse(interaction *discordgo.Interaction) (*discordgo.Message, error) {
	m.InteractionResponseCalled = true
	if m.InteractionResponseError != nil {
		return nil, m.InteractionResponseError
	}
	return m.InteractionResponseReturn, nil
}

// MessageReactionAdd mocks the Discord session MessageReactionAdd method
func (m *MockSession) MessageReactionAdd(channelID, messageID, emojiID string) error {
	m.MessageReactionAddCalled = true
	return m.MessageReactionAddError
}

// State mocks the Discord session State method
func (m *MockSession) State() *discordgo.State {
	m.StateCalled = true
	return m.StateReturn
}

// Reset resets all mock state
func (m *MockSession) Reset() {
	m.RespondCalled = false
	m.RespondError = nil
	m.RespondData = nil
	m.RespondType = 0
	m.InteractionResponseEditCalled = false
	m.InteractionResponseEditError = nil
	m.InteractionResponseEditReturn = nil
	m.FollowupCalled = false
	m.FollowupError = nil
	m.FollowupReturn = nil
	m.GuildCalled = false
	m.GuildError = nil
	m.GuildReturn = nil
	m.ChannelCalled = false
	m.ChannelError = nil
	m.ChannelReturn = nil
	m.StateCalled = false
	m.StateReturn = nil
	m.GuildEmojisCalled = false
	m.GuildEmojisError = nil
	m.GuildEmojisReturn = nil
	m.InteractionResponseCalled = false
	m.InteractionResponseError = nil
	m.InteractionResponseReturn = nil
	m.MessageReactionAddCalled = false
	m.MessageReactionAddError = nil
}
