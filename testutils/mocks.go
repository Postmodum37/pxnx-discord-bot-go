package testutils

import (
	"github.com/bwmarrin/discordgo"
)

// MockSession implements a mock Discord session for testing
type MockSession struct {
	RespondCalled   bool
	RespondError    error
	RespondData     *discordgo.InteractionResponseData
	GuildCalled     bool
	GuildError      error
	GuildReturn     *discordgo.Guild
	GuildEmojisCalled bool
	GuildEmojisError  error
	GuildEmojisReturn []*discordgo.Emoji
	InteractionResponseCalled bool
	InteractionResponseError  error
	InteractionResponseReturn *discordgo.Message
	MessageReactionAddCalled bool
	MessageReactionAddError  error
}

// InteractionRespond mocks the Discord session InteractionRespond method
func (m *MockSession) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	m.RespondCalled = true
	if resp.Data != nil {
		m.RespondData = resp.Data
	}
	return m.RespondError
}

// Guild mocks the Discord session Guild method
func (m *MockSession) Guild(guildID string) (*discordgo.Guild, error) {
	m.GuildCalled = true
	if m.GuildError != nil {
		return nil, m.GuildError
	}
	return m.GuildReturn, nil
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

// Reset resets all mock state
func (m *MockSession) Reset() {
	m.RespondCalled = false
	m.RespondError = nil
	m.RespondData = nil
	m.GuildCalled = false
	m.GuildError = nil
	m.GuildReturn = nil
	m.GuildEmojisCalled = false
	m.GuildEmojisError = nil
	m.GuildEmojisReturn = nil
	m.InteractionResponseCalled = false
	m.InteractionResponseError = nil
	m.InteractionResponseReturn = nil
	m.MessageReactionAddCalled = false
	m.MessageReactionAddError = nil
}