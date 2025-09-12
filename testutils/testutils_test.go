package testutils

import (
	"errors"
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestMockSession(t *testing.T) {
	mock := &MockSession{}

	// Test InteractionRespond
	interaction := &discordgo.Interaction{}
	response := &discordgo.InteractionResponse{}
	
	err := mock.InteractionRespond(interaction, response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !mock.RespondCalled {
		t.Error("Expected RespondCalled to be true")
	}

	// Test Reset
	mock.Reset()
	if mock.RespondCalled {
		t.Error("Expected RespondCalled to be false after reset")
	}
}

func TestCreateTestUser(t *testing.T) {
	user := CreateTestUser("123", "testuser", "avatar_hash")
	
	if user.ID != "123" {
		t.Errorf("Expected ID '123', got '%s'", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", user.Username)
	}
	if user.Avatar != "avatar_hash" {
		t.Errorf("Expected avatar 'avatar_hash', got '%s'", user.Avatar)
	}
	if user.Discriminator != "0001" {
		t.Errorf("Expected discriminator '0001', got '%s'", user.Discriminator)
	}
}

func TestCreateTestMember(t *testing.T) {
	user := CreateTestUser("456", "memberuser", "member_avatar")
	member := CreateTestMember(user)
	
	if member.User != user {
		t.Error("Expected member to have the provided user")
	}
	if member.Nick != "" {
		t.Errorf("Expected empty nick, got '%s'", member.Nick)
	}
	if member.JoinedAt.IsZero() {
		t.Error("Expected JoinedAt to be set")
	}
}

func TestCreateTestGuild(t *testing.T) {
	guild := CreateTestGuild("guild123", "Test Guild", 100)
	
	if guild.ID != "guild123" {
		t.Errorf("Expected ID 'guild123', got '%s'", guild.ID)
	}
	if guild.Name != "Test Guild" {
		t.Errorf("Expected name 'Test Guild', got '%s'", guild.Name)
	}
	if guild.MemberCount != 100 {
		t.Errorf("Expected member count 100, got %d", guild.MemberCount)
	}
	if guild.Icon != "test_icon_hash" {
		t.Errorf("Expected icon 'test_icon_hash', got '%s'", guild.Icon)
	}
	if guild.OwnerID != "owner_id_123" {
		t.Errorf("Expected owner ID 'owner_id_123', got '%s'", guild.OwnerID)
	}
}

func TestCreateTestInteraction(t *testing.T) {
	options := []*discordgo.ApplicationCommandInteractionDataOption{
		CreateStringOption("param1", "value1"),
	}
	
	interaction := CreateTestInteraction("testcommand", options)
	
	if interaction.ID != "interaction_id_123" {
		t.Errorf("Expected ID 'interaction_id_123', got '%s'", interaction.ID)
	}
	if interaction.Type != discordgo.InteractionApplicationCommand {
		t.Errorf("Expected type ApplicationCommand, got %v", interaction.Type)
	}
	if interaction.GuildID != "guild_id_123" {
		t.Errorf("Expected guild ID 'guild_id_123', got '%s'", interaction.GuildID)
	}
	if interaction.ChannelID != "channel_id_123" {
		t.Errorf("Expected channel ID 'channel_id_123', got '%s'", interaction.ChannelID)
	}
	
	data := interaction.Data.(discordgo.ApplicationCommandInteractionData)
	if data.Name != "testcommand" {
		t.Errorf("Expected command name 'testcommand', got '%s'", data.Name)
	}
	if len(data.Options) != 1 {
		t.Errorf("Expected 1 option, got %d", len(data.Options))
	}
}

func TestCreateStringOption(t *testing.T) {
	option := CreateStringOption("testparam", "testvalue")
	
	if option.Name != "testparam" {
		t.Errorf("Expected name 'testparam', got '%s'", option.Name)
	}
	if option.Type != discordgo.ApplicationCommandOptionString {
		t.Errorf("Expected string type, got %v", option.Type)
	}
	if option.Value != "testvalue" {
		t.Errorf("Expected value 'testvalue', got %v", option.Value)
	}
}

func TestCreateUserOption(t *testing.T) {
	user := CreateTestUser("789", "optionuser", "option_avatar")
	option := CreateUserOption("targetuser", user)
	
	if option.Name != "targetuser" {
		t.Errorf("Expected name 'targetuser', got '%s'", option.Name)
	}
	if option.Type != discordgo.ApplicationCommandOptionUser {
		t.Errorf("Expected user type, got %v", option.Type)
	}
	if option.Value != user.ID {
		t.Errorf("Expected value to be the user ID '%s', got %v", user.ID, option.Value)
	}
}

func TestCreateTestEmojis(t *testing.T) {
	emojis := CreateTestEmojis()
	
	if len(emojis) != 3 {
		t.Errorf("Expected 3 emojis, got %d", len(emojis))
	}
	
	for i, emoji := range emojis {
		if emoji.ID == "" {
			t.Errorf("Expected non-empty ID for emoji %d", i)
		}
		if emoji.Name == "" {
			t.Errorf("Expected non-empty name for emoji %d", i)
		}
	}
	
	// Check that emojis are unique
	ids := make(map[string]bool)
	names := make(map[string]bool)
	for i, emoji := range emojis {
		if ids[emoji.ID] {
			t.Errorf("Duplicate emoji ID at index %d: %s", i, emoji.ID)
		}
		if names[emoji.Name] {
			t.Errorf("Duplicate emoji name at index %d: %s", i, emoji.Name)
		}
		ids[emoji.ID] = true
		names[emoji.Name] = true
	}
}

func TestCreateTestMessage(t *testing.T) {
	message := CreateTestMessage("msg123", "Hello world")
	
	if message.ID != "msg123" {
		t.Errorf("Expected ID 'msg123', got '%s'", message.ID)
	}
	if message.Content != "Hello world" {
		t.Errorf("Expected content 'Hello world', got '%s'", message.Content)
	}
	if message.Author == nil {
		t.Error("Expected author to be set")
	} else {
		if message.Author.ID != "author_123" {
			t.Errorf("Expected author ID 'author_123', got '%s'", message.Author.ID)
		}
		if message.Author.Username != "test_author" {
			t.Errorf("Expected author username 'test_author', got '%s'", message.Author.Username)
		}
	}
}

func TestMockSessionAdvanced(t *testing.T) {
	mock := &MockSession{}
	
	// Test Guild method
	testGuild := CreateTestGuild("test_guild", "Test Guild", 50)
	mock.GuildReturn = testGuild
	
	guild, err := mock.Guild("test_guild")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if guild != testGuild {
		t.Error("Expected returned guild to match set guild")
	}
	if !mock.GuildCalled {
		t.Error("Expected GuildCalled to be true")
	}
	
	// Test Guild method with error
	mock.Reset()
	mock.GuildError = errors.New("guild error")
	
	guild, err = mock.Guild("test_guild")
	if err == nil {
		t.Error("Expected error but got none")
	}
	if guild != nil {
		t.Error("Expected nil guild when error occurs")
	}
}

func TestMockSessionEmojis(t *testing.T) {
	mock := &MockSession{}
	
	// Test GuildEmojis method
	testEmojis := CreateTestEmojis()
	mock.GuildEmojisReturn = testEmojis
	
	emojis, err := mock.GuildEmojis("guild123")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if len(emojis) != len(testEmojis) {
		t.Errorf("Expected %d emojis, got %d", len(testEmojis), len(emojis))
	}
	if !mock.GuildEmojisCalled {
		t.Error("Expected GuildEmojisCalled to be true")
	}
}