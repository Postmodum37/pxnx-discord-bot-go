package testutils

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// CreateTestUser creates a test user with default values
func CreateTestUser(id, username, avatar string) *discordgo.User {
	return &discordgo.User{
		ID:            id,
		Username:      username,
		Avatar:        avatar,
		Discriminator: "0001",
	}
}

// CreateTestMember creates a test member with default values
func CreateTestMember(user *discordgo.User) *discordgo.Member {
	return &discordgo.Member{
		User:     user,
		Nick:     "",
		JoinedAt: time.Now(),
	}
}

// CreateTestGuild creates a test guild with default values
func CreateTestGuild(id, name string, memberCount int) *discordgo.Guild {
	return &discordgo.Guild{
		ID:          id,
		Name:        name,
		Icon:        "test_icon_hash",
		OwnerID:     "owner_id_123",
		MemberCount: memberCount,
	}
}

// CreateTestInteraction creates a test interaction for commands
func CreateTestInteraction(commandName string, options []*discordgo.ApplicationCommandInteractionDataOption) *discordgo.InteractionCreate {
	// Build resolved data for user options
	resolved := &discordgo.ApplicationCommandInteractionDataResolved{
		Users: make(map[string]*discordgo.User),
	}
	
	for _, option := range options {
		if option.Type == discordgo.ApplicationCommandOptionUser {
			if userID, ok := option.Value.(string); ok {
				// Use the existing user that was passed to CreateUserOption
				// This gets set in CreateUserOptionWithUser below
				if option.Name == "target" {
					resolved.Users[userID] = CreateTestUser(userID, "targetuser", "target_avatar")
				} else {
					resolved.Users[userID] = CreateTestUser(userID, "resolved_user", "resolved_avatar")
				}
			}
		}
	}

	return &discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			ID:        "interaction_id_123",
			Type:      discordgo.InteractionApplicationCommand,
			GuildID:   "guild_id_123",
			ChannelID: "channel_id_123",
			Data: discordgo.ApplicationCommandInteractionData{
				ID:       "command_id_123",
				Name:     commandName,
				Options:  options,
				Resolved: resolved,
			},
		},
	}
}

// CreateStringOption creates a string command option for testing
func CreateStringOption(name, value string) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name:  name,
		Type:  discordgo.ApplicationCommandOptionString,
		Value: value,
	}
}

// CreateUserOption creates a user command option for testing
func CreateUserOption(name string, user *discordgo.User) *discordgo.ApplicationCommandInteractionDataOption {
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name:  name,
		Type:  discordgo.ApplicationCommandOptionUser,
		Value: user.ID,
	}
}

// CreateUserOptionWithResolved creates a user option and stores the user for resolved data
func CreateUserOptionWithResolved(name string, user *discordgo.User, resolved *discordgo.ApplicationCommandInteractionDataResolved) *discordgo.ApplicationCommandInteractionDataOption {
	if resolved.Users == nil {
		resolved.Users = make(map[string]*discordgo.User)
	}
	resolved.Users[user.ID] = user
	
	return &discordgo.ApplicationCommandInteractionDataOption{
		Name:  name,
		Type:  discordgo.ApplicationCommandOptionUser,
		Value: user.ID,
	}
}

// CreateTestEmojis creates test emojis for testing
func CreateTestEmojis() []*discordgo.Emoji {
	return []*discordgo.Emoji{
		{
			ID:   "emoji1",
			Name: "test_emoji_1",
		},
		{
			ID:   "emoji2",
			Name: "test_emoji_2",
		},
		{
			ID:   "emoji3",
			Name: "test_emoji_3",
		},
	}
}

// CreateTestMessage creates a test message for testing
func CreateTestMessage(id, content string) *discordgo.Message {
	return &discordgo.Message{
		ID:      id,
		Content: content,
		Author:  CreateTestUser("author_123", "test_author", "author_avatar"),
	}
}