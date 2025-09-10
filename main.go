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

func interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	var err error
	switch i.ApplicationCommandData().Name {
	case "ping":
		err = handlePingCommand(s, i)
	case "peepee":
		err = handlePeepeeCommandWithReaction(s, i)
	}
	
	if err != nil {
		log.Printf("Error handling command '%s': %v", i.ApplicationCommandData().Name, err)
	}
}