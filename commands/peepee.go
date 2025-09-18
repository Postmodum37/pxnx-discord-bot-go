package commands

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Initialize random seed once for peepee command
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// sessionWrapper is a minimal wrapper to adapt *discordgo.Session to SessionInterface
type sessionWrapper struct {
	session *discordgo.Session
}

func (sw *sessionWrapper) InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error {
	return sw.session.InteractionRespond(interaction, resp, options...)
}

func (sw *sessionWrapper) InteractionResponseEdit(interaction *discordgo.Interaction, newresp *discordgo.WebhookEdit, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return sw.session.InteractionResponseEdit(interaction, newresp, options...)
}

func (sw *sessionWrapper) FollowupMessageCreate(interaction *discordgo.Interaction, wait bool, data *discordgo.WebhookParams, options ...discordgo.RequestOption) (*discordgo.Message, error) {
	return sw.session.FollowupMessageCreate(interaction, wait, data, options...)
}

func (sw *sessionWrapper) Guild(guildID string, options ...discordgo.RequestOption) (*discordgo.Guild, error) {
	return sw.session.Guild(guildID, options...)
}

func (sw *sessionWrapper) Channel(channelID string, options ...discordgo.RequestOption) (*discordgo.Channel, error) {
	return sw.session.Channel(channelID, options...)
}

func (sw *sessionWrapper) State() *discordgo.State {
	return sw.session.State
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

// getRandomPhrase returns a random phrase with display name from the peepee definitions
func getRandomPhrase(displayName string) string {
	definition := peepeeDefinitions[rng.Intn(len(peepeeDefinitions))]
	return fmt.Sprintf("%s %s peepee!", displayName, definition)
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
	// Use GlobalName if available, otherwise fallback to Username
	displayName := user.GlobalName
	if displayName == "" {
		displayName = user.Username
	}

	randomPhrase := getRandomPhrase(displayName)
	avatarURL := getUserAvatarURL(user)

	return &discordgo.MessageEmbed{
		Title:       "PeePee Inspection Time",
		Description: randomPhrase,
		Color:       0x3498db, // ColorBlue
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

// HandlePeepeeCommand handles the peepee slash command (without emoji reaction)
func HandlePeepeeCommand(s SessionInterface, i *discordgo.InteractionCreate) error {
	embed := createPeepeeEmbed(i.Member.User)

	return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// HandlePeepeeCommandWithReaction handles the peepee command with emoji reaction
func HandlePeepeeCommandWithReaction(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Create a simple wrapper for the session since we need SessionInterface
	sessionWrapper := &sessionWrapper{session: s}
	err := HandlePeepeeCommand(sessionWrapper, i)
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
