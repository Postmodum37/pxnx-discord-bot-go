package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/bwmarrin/discordgo"
)

func main() {
	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable is required")
	}

	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session:", err)
	}

	dg.AddHandler(ready)
	dg.AddHandler(interactionCreate)

	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildEmojis

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer dg.Close()

	fmt.Println("Bot is running. Press CTRL+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	fmt.Println("Gracefully shutting down.")
}

// getCommands returns the list of application commands for the bot
func getCommands() []*discordgo.ApplicationCommand {
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
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "question",
					Description: "Your question for the magic 8-ball",
					Required:    true,
				},
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
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "target",
					Description: "The user to get info about (optional)",
					Required:    false,
				},
			},
		},
		{
			Name:        "weather",
			Description: "Get the weather forecast for a city",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "city",
					Description: "City name to get weather for",
					Required:    true,
				},
			},
		},
	}
}

// registerCommands registers all bot commands with Discord
func registerCommands(s *discordgo.Session) error {
	commands := getCommands()
	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			return fmt.Errorf("cannot create '%v' command: %w", cmd.Name, err)
		}
	}
	return nil
}

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("Logged in as: %v#%v\n", s.State.User.Username, s.State.User.Discriminator)

	if err := registerCommands(s); err != nil {
		log.Printf("Error registering commands: %v", err)
	}
}

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

var peepeeDefinitions = []string{
	"has a certified micro",
	"is packing a pocket-sized",
	"owns a travel-friendly",
	"rocks a fun-sized",
	"sports a bite-sized",
	"carries a portable",
	"features a mini",
	"displays a compact",
	"showcases a teeny-weeny",
	"presents a smol bean",
	"boasts a keychain-sized",
	"flaunts a coin-sized",
	"exhibits a thimble-sized",
	"demonstrates a whisper-quiet",
	"manifests an itty-bitty",
	"reveals a microscopic",
	"unveils a nano-scale",
	"shows off a barely-there",
	"owns a limited edition tiny",
	"has equipped a stealth mode",
}

// getRandomPhrase returns a random phrase with username from the peepee definitions
func getRandomPhrase(username string) string {
	rand.Seed(time.Now().UnixNano())
	definition := peepeeDefinitions[rand.Intn(len(peepeeDefinitions))]
	return fmt.Sprintf("%s %s peepee!", username, definition)
}

// getUserAvatarURL gets the user's avatar URL with fallback to default
func getUserAvatarURL(user *discordgo.User) string {
	avatarURL := user.AvatarURL("256")
	if avatarURL == "" {
		// Use Discord's default avatar URL format
		avatarURL = "https://cdn.discordapp.com/embed/avatars/" + user.Discriminator[len(user.Discriminator)-1:] + ".png"
	}
	return avatarURL
}

// createPeepeeEmbed creates an embed for the peepee command
func createPeepeeEmbed(user *discordgo.User) *discordgo.MessageEmbed {
	randomPhrase := getRandomPhrase(user.Username)
	avatarURL := getUserAvatarURL(user)
	
	return &discordgo.MessageEmbed{
		Title:       "PeePee Inspection Time",
		Description: randomPhrase,
		Color:       0x3498db, // Blue color
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: avatarURL,
		},
	}
}

func getRandomEmoji(s *discordgo.Session, guildID string) string {
	if s == nil || guildID == "" {
		return "🔍" // fallback emoji
	}
	
	emojis, err := s.GuildEmojis(guildID)
	if err != nil || len(emojis) == 0 {
		return "🔍" // fallback emoji
	}
	randomEmoji := emojis[rand.Intn(len(emojis))]
	return randomEmoji.APIName()
}

// SessionInterface defines the methods we need from a Discord session for testing
type SessionInterface interface {
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
}

// handlePingCommand handles the ping slash command
func handlePingCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "Pong!",
		},
	})
}

// handlePeepeeCommand handles the peepee slash command (without emoji reaction)
func handlePeepeeCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	embed := createPeepeeEmbed(i.Member.User)
	
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// handlePeepeeCommandWithReaction handles the peepee command with emoji reaction
func handlePeepeeCommandWithReaction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := handlePeepeeCommand(s, i)
	if err != nil {
		return err
	}
	
	// Add random emoji reaction
	if i.GuildID != "" {
		emoji := getRandomEmoji(s, i.GuildID)
		go func() {
			// Small delay to ensure message is sent
			time.Sleep(100 * time.Millisecond)
			// Get the interaction response message
			msg, err := s.InteractionResponse(i.Interaction)
			if err == nil {
				s.MessageReactionAdd(i.ChannelID, msg.ID, emoji)
			}
		}()
	}
	
	return nil
}

// get8BallResponse returns a random 8-ball response
func get8BallResponse() string {
	rand.Seed(time.Now().UnixNano())
	return eightBallResponses[rand.Intn(len(eightBallResponses))]
}

// handle8BallCommand handles the 8ball slash command
func handle8BallCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	question := options[0].StringValue()
	response := get8BallResponse()
	
	embed := &discordgo.MessageEmbed{
		Title: "🎱 Magic 8-Ball",
		Color: 0x9b59b6,
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

// handleCoinFlipCommand handles the coinflip slash command
func handleCoinFlipCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	rand.Seed(time.Now().UnixNano())
	result := "Heads"
	if rand.Intn(2) == 1 {
		result = "Tails"
	}
	
	embed := &discordgo.MessageEmbed{
		Title:       "🪙 Coin Flip",
		Description: fmt.Sprintf("The coin landed on **%s**!", result),
		Color:       0xf39c12,
	}
	
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// handleServerCommand handles the server slash command
func handleServerCommand(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild, err := s.Guild(i.GuildID)
	if err != nil {
		return err
	}
	
	memberCount := guild.MemberCount
	createdAt, _ := discordgo.SnowflakeTimestamp(guild.ID)
	
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("📊 %s Server Info", guild.Name),
		Description: fmt.Sprintf("Here's some information about **%s**", guild.Name),
		Color:       0x2ecc71,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: guild.IconURL("256"),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "👥 Members",
				Value:  fmt.Sprintf("%d", memberCount),
				Inline: true,
			},
			{
				Name:   "🆔 Server ID",
				Value:  guild.ID,
				Inline: true,
			},
			{
				Name:   "👑 Owner",
				Value:  fmt.Sprintf("<@%s>", guild.OwnerID),
				Inline: true,
			},
			{
				Name:   "🗓️ Created",
				Value:  fmt.Sprintf("<t:%d:F>", createdAt.Unix()),
				Inline: false,
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

// handleUserCommand handles the user slash command
func handleUserCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	var targetUser *discordgo.User
	
	// Check if a user was mentioned in options
	if len(i.ApplicationCommandData().Options) > 0 {
		targetUser = i.ApplicationCommandData().Options[0].UserValue(nil)
	} else {
		// Use the command invoker
		targetUser = i.Member.User
	}
	
	avatarURL := getUserAvatarURL(targetUser)
	userCreated, _ := discordgo.SnowflakeTimestamp(targetUser.ID)
	
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("👤 %s's Profile", targetUser.Username),
		Description: fmt.Sprintf("Here's some information about **%s**", targetUser.Mention()),
		Color:       0xe74c3c,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: avatarURL,
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🏷️ Username",
				Value:  targetUser.Username,
				Inline: true,
			},
			{
				Name:   "🆔 User ID",
				Value:  targetUser.ID,
				Inline: true,
			},
			{
				Name:   "🗓️ Account Created",
				Value:  fmt.Sprintf("<t:%d:F>", userCreated.Unix()),
				Inline: false,
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

// handleWeatherCommand handles the weather slash command
func handleWeatherCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	city := options[0].StringValue()
	
	// This is a mock weather response since we don't have a real API
	weatherConditions := []string{"Sunny", "Cloudy", "Rainy", "Snowy", "Partly Cloudy", "Stormy"}
	rand.Seed(time.Now().UnixNano())
	condition := weatherConditions[rand.Intn(len(weatherConditions))]
	temp := rand.Intn(35) + 5 // Random temp between 5-40°C
	
	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("🌤️ Weather in %s", city),
		Description: fmt.Sprintf("Here's the current weather forecast for **%s**", city),
		Color:       0x3498db,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🌡️ Temperature",
				Value:  fmt.Sprintf("%d°C", temp),
				Inline: true,
			},
			{
				Name:   "☁️ Condition",
				Value:  condition,
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "⚠️ This is a mock weather service for demonstration purposes",
		},
	}
	
	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error
	switch i.ApplicationCommandData().Name {
	case "ping":
		err = handlePingCommand(s, i)
	case "peepee":
		err = handlePeepeeCommandWithReaction(s, i)
	case "8ball":
		err = handle8BallCommand(s, i)
	case "coinflip":
		err = handleCoinFlipCommand(s, i)
	case "server":
		err = handleServerCommand(s, i)
	case "user":
		err = handleUserCommand(s, i)
	case "weather":
		err = handleWeatherCommand(s, i)
	}
	
	if err != nil {
		log.Printf("Error handling command '%s': %v", i.ApplicationCommandData().Name, err)
	}
}