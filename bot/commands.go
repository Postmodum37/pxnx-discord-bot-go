package bot

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
)

// createStringOption creates a string application command option
func createStringOption(name, description string, required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// createUserOption creates a user application command option
func createUserOption(name, description string, required bool) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionUser,
		Name:        name,
		Description: description,
		Required:    required,
	}
}

// createIntegerOption creates an integer application command option
func createIntegerOption(name, description string, required bool, minValue, maxValue *float64) *discordgo.ApplicationCommandOption {
	option := &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionInteger,
		Name:        name,
		Description: description,
		Required:    required,
	}

	if minValue != nil {
		option.MinValue = minValue
	}
	if maxValue != nil {
		option.MaxValue = *maxValue
	}

	return option
}

// createStringChoiceOption creates a string application command option with choices
func createStringChoiceOption(name, description string, required bool, choices []*discordgo.ApplicationCommandOptionChoice) *discordgo.ApplicationCommandOption {
	return &discordgo.ApplicationCommandOption{
		Type:        discordgo.ApplicationCommandOptionString,
		Name:        name,
		Description: description,
		Required:    required,
		Choices:     choices,
	}
}

// GetCommands returns the list of application commands for the bot
func GetCommands() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "ping",
			Description: "Responds with Pong!",
		},
		{
			Name:        "peepee",
			Description: "PeePee Inspection Time!",
		},
		{
			Name:        "8ball",
			Description: "Ask the magic 8-ball a question",
			Options: []*discordgo.ApplicationCommandOption{
				createStringOption("question", "Your question for the magic 8-ball", true),
			},
		},
		{
			Name:        "coinflip",
			Description: "Flip a coin and choose heads or tails",
		},
		{
			Name:        "server",
			Description: "Provides information about the server",
		},
		{
			Name:        "user",
			Description: "Replies with user info!",
			Options: []*discordgo.ApplicationCommandOption{
				createUserOption("target", "The user to get info about (optional)", false),
			},
		},
		{
			Name:        "weather",
			Description: "Get the weather forecast for a city",
			Options: []*discordgo.ApplicationCommandOption{
				createStringOption("city", "City name to get weather for", true),
				createStringChoiceOption("duration", "Weather forecast duration", false, []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "Current Weather",
						Value: "current",
					},
					{
						Name:  "1-Day Forecast",
						Value: "1-day",
					},
					{
						Name:  "5-Day Forecast",
						Value: "5-day",
					},
				}),
			},
		},
		{
			Name:        "roll",
			Description: "Roll a dice with specified maximum value (default: 100)",
			Options: []*discordgo.ApplicationCommandOption{
				createIntegerOption("max", "Maximum value for the dice roll (1-1000000)", false, func() *float64 { v := float64(1); return &v }(), func() *float64 { v := float64(1000000); return &v }()),
			},
		},
	}
}

// RegisterCommands registers all bot commands with Discord (includes cleanup of existing commands)
func RegisterCommands(s *discordgo.Session) error {
	fmt.Println("Starting command registration process...")

	// Always clean up existing commands first to ensure clean state
	fmt.Println("Cleaning up existing commands...")

	// Clear all existing global commands
	fmt.Println("Retrieving existing global commands...")
	existingCommands, err := s.ApplicationCommands(s.State.User.ID, "")
	if err != nil {
		return fmt.Errorf("cannot retrieve existing commands: %w", err)
	}

	// Delete all existing global commands
	if len(existingCommands) > 0 {
		fmt.Printf("Deleting %d existing global commands...\n", len(existingCommands))
		for _, cmd := range existingCommands {
			fmt.Printf("Deleting global command: %s\n", cmd.Name)
			err := s.ApplicationCommandDelete(s.State.User.ID, "", cmd.ID)
			if err != nil {
				return fmt.Errorf("cannot delete existing command '%v': %w", cmd.Name, err)
			}
		}
	} else {
		fmt.Println("No existing global commands found.")
	}

	// Also clear guild-specific commands for all guilds the bot is in
	fmt.Println("Clearing guild-specific commands...")
	for _, guild := range s.State.Guilds {
		guildCommands, err := s.ApplicationCommands(s.State.User.ID, guild.ID)
		if err != nil {
			fmt.Printf("Warning: Could not retrieve commands for guild %s: %v\n", guild.ID, err)
			continue
		}

		if len(guildCommands) > 0 {
			fmt.Printf("Deleting %d commands from guild %s...\n", len(guildCommands), guild.ID)
			for _, cmd := range guildCommands {
				fmt.Printf("Deleting guild command: %s from guild %s\n", cmd.Name, guild.ID)
				err := s.ApplicationCommandDelete(s.State.User.ID, guild.ID, cmd.ID)
				if err != nil {
					fmt.Printf("Warning: Could not delete command '%s' from guild %s: %v\n", cmd.Name, guild.ID, err)
				}
			}
		}
	}

	// Register the current commands as global commands
	fmt.Println("Registering new global commands...")
	commands := GetCommands()
	for _, cmd := range commands {
		fmt.Printf("Creating global command: %s\n", cmd.Name)
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("cannot create '%v' command: %w", cmd.Name, err)
		}
	}

	fmt.Printf("Successfully registered %d commands!\n", len(commands))
	return nil
}