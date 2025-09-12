package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Embed color constants
const (
	ColorBlue   = 0x3498db
	ColorPurple = 0x9b59b6
	ColorOrange = 0xf39c12
	ColorGreen  = 0x2ecc71
	ColorRed    = 0xe74c3c
)

// Initialize random seed once
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// Global flag for command registration
var shouldRegisterCommands bool

func main() {
	// Parse command line flags
	registerCommands := flag.Bool("register-commands", false, "Register bot commands with Discord (cleans up existing commands first)")
	flag.Parse()

	// Set global flag
	shouldRegisterCommands = *registerCommands

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

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
	defer func() {
		if err := dg.Close(); err != nil {
			log.Printf("Error closing Discord session: %v", err)
		}
	}()

	fmt.Println("Bot is running. Press CTRL+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	fmt.Println("Gracefully shutting down.")
}

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
			},
		},
	}
}

// registerCommands registers all bot commands with Discord (includes cleanup of existing commands)
func registerCommands(s *discordgo.Session) error {
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
	commands := getCommands()
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

func ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("Logged in as: %v#%v\n", s.State.User.Username, s.State.User.Discriminator)

	if shouldRegisterCommands {
		if err := registerCommands(s); err != nil {
			log.Printf("Error registering commands: %v", err)
			return
		}
		fmt.Println("Command registration complete. Bot is ready!")
	} else {
		fmt.Println("Bot is ready! (Use --register-commands flag to register slash commands)")
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

// WeatherData represents the response from OpenWeatherMap API
type WeatherData struct {
	Main struct {
		Temp      float64 `json:"temp"`
		FeelsLike float64 `json:"feels_like"`
		Humidity  int     `json:"humidity"`
	} `json:"main"`
	Weather []struct {
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
	} `json:"wind"`
	Name string `json:"name"`
	Sys  struct {
		Country string `json:"country"`
	} `json:"sys"`
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
	definition := peepeeDefinitions[rng.Intn(len(peepeeDefinitions))]
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

// createErrorEmbed creates a standardized error embed
func createErrorEmbed(title, description, errorMsg string) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       ColorRed,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:  "Error",
				Value: errorMsg,
			},
		},
	}
}

// createPeepeeEmbed creates an embed for the peepee command
func createPeepeeEmbed(user *discordgo.User) *discordgo.MessageEmbed {
	randomPhrase := getRandomPhrase(user.Username)
	avatarURL := getUserAvatarURL(user)

	return &discordgo.MessageEmbed{
		Title:       "PeePee Inspection Time",
		Description: randomPhrase,
		Color:       ColorBlue,
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
	randomEmoji := emojis[rng.Intn(len(emojis))]
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
				if err := s.MessageReactionAdd(i.ChannelID, msg.ID, emoji); err != nil {
					log.Printf("Error adding reaction: %v", err)
				}
			}
		}()
	}

	return nil
}

// get8BallResponse returns a random 8-ball response
func get8BallResponse() string {
	return eightBallResponses[rng.Intn(len(eightBallResponses))]
}

// handle8BallCommand handles the 8ball slash command
func handle8BallCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	question := options[0].StringValue()
	response := get8BallResponse()

	embed := &discordgo.MessageEmbed{
		Title: "🎱 Magic 8-Ball",
		Color: ColorPurple,
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
	result := "Heads"
	if rng.Intn(2) == 1 {
		result = "Tails"
	}

	embed := &discordgo.MessageEmbed{
		Title:       "🪙 Coin Flip",
		Description: fmt.Sprintf("The coin landed on **%s**!", result),
		Color:       ColorOrange,
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
		Color:       ColorGreen,
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
		Color:       ColorRed,
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

// getWeatherData fetches weather data from OpenWeatherMap API
func getWeatherData(city string) (*WeatherData, error) {
	apiKey := os.Getenv("OPENWEATHER_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENWEATHER_API_KEY environment variable is required")
	}

	// URL encode the city name to handle spaces and special characters
	encodedCity := url.QueryEscape(city)
	apiURL := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?q=%s&appid=%s&units=metric", encodedCity, apiKey)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch weather data: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("weather API returned status %d", resp.StatusCode)
	}

	var weatherData WeatherData
	if err := json.NewDecoder(resp.Body).Decode(&weatherData); err != nil {
		return nil, fmt.Errorf("failed to decode weather data: %w", err)
	}

	return &weatherData, nil
}

// getWeatherIcon returns an appropriate emoji for the weather condition
func getWeatherIcon(condition string) string {
	lowerCondition := strings.ToLower(condition)
	switch {
	case strings.Contains(lowerCondition, "clear"):
		return "☀️"
	case strings.Contains(lowerCondition, "cloud"):
		return "☁️"
	case strings.Contains(lowerCondition, "rain"):
		return "🌧️"
	case strings.Contains(lowerCondition, "snow"):
		return "❄️"
	case strings.Contains(lowerCondition, "thunder"):
		return "⛈️"
	case strings.Contains(lowerCondition, "mist") || strings.Contains(lowerCondition, "fog"):
		return "🌫️"
	default:
		return "🌤️"
	}
}

// handleWeatherCommand handles the weather slash command
func handleWeatherCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	options := i.ApplicationCommandData().Options
	city := options[0].StringValue()

	weatherData, err := getWeatherData(city)
	if err != nil {
		// Return error embed if API call fails
		errorEmbed := createErrorEmbed(
			"❌ Weather Error",
			fmt.Sprintf("Unable to fetch weather data for **%s**", city),
			"City not found or API error. Please check the city name and try again.",
		)
		errorEmbed.Footer = &discordgo.MessageEmbedFooter{
			Text: "Powered by OpenWeatherMap",
		}

		return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Embeds: []*discordgo.MessageEmbed{errorEmbed},
			},
		})
	}

	// Format temperature
	temp := fmt.Sprintf("%.1f°C", weatherData.Main.Temp)
	feelsLike := fmt.Sprintf("%.1f°C", weatherData.Main.FeelsLike)

	// Get weather condition and icon
	condition := "Unknown"
	description := "No description available"
	if len(weatherData.Weather) > 0 {
		condition = weatherData.Weather[0].Main
		titleCaser := cases.Title(language.English)
		description = titleCaser.String(weatherData.Weather[0].Description)
	}

	weatherIcon := getWeatherIcon(condition)
	location := weatherData.Name
	if weatherData.Sys.Country != "" {
		location = fmt.Sprintf("%s, %s", weatherData.Name, weatherData.Sys.Country)
	}

	embed := &discordgo.MessageEmbed{
		Title:       fmt.Sprintf("%s Weather in %s", weatherIcon, location),
		Description: description,
		Color:       ColorBlue,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "🌡️ Temperature",
				Value:  temp,
				Inline: true,
			},
			{
				Name:   "🤏 Feels Like",
				Value:  feelsLike,
				Inline: true,
			},
			{
				Name:   "💧 Humidity",
				Value:  fmt.Sprintf("%d%%", weatherData.Main.Humidity),
				Inline: true,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Powered by OpenWeatherMap",
		},
		Timestamp: time.Now().Format(time.RFC3339),
	}

	// Add wind information if available
	if weatherData.Wind.Speed > 0 {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   "💨 Wind Speed",
			Value:  fmt.Sprintf("%.1f m/s", weatherData.Wind.Speed),
			Inline: true,
		})
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
