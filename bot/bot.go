package bot

import (
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"

	"pxnx-discord-bot/commands"
	"pxnx-discord-bot/music/manager"
	"pxnx-discord-bot/music/providers"
)

// Bot represents the Discord bot instance
type Bot struct {
	Session *discordgo.Session
}

// New creates a new bot instance
func New(token string) (*Bot, error) {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		return nil, fmt.Errorf("error creating Discord session: %w", err)
	}

	return &Bot{Session: dg}, nil
}

// Setup configures the bot with handlers and intents
func (b *Bot) Setup() {
	b.Session.AddHandler(b.ready)
	b.Session.AddHandler(b.interactionCreate)
	b.Session.AddHandler(b.voiceStateUpdate)
	b.Session.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuildEmojis | discordgo.IntentsGuildVoiceStates

	// Initialize the music manager
	sessionWrapper := manager.NewSessionWrapper(b.Session)
	commands.MusicManager = manager.NewManager(sessionWrapper)

	// Register audio providers
	youtubeProvider := providers.NewYouTubeProvider()
	commands.MusicManager.RegisterProvider(youtubeProvider)
}

// Start opens the Discord connection
func (b *Bot) Start() error {
	return b.Session.Open()
}

// Stop closes the Discord connection
func (b *Bot) Stop() error {
	return b.Session.Close()
}

// ready handles the ready event
func (b *Bot) ready(s *discordgo.Session, event *discordgo.Ready) {
	fmt.Printf("Logged in as: %v#%v\n", s.State.User.Username, s.State.User.Discriminator)

	if shouldRegisterCommands {
		if err := RegisterCommands(s); err != nil {
			log.Printf("Error registering commands: %v", err)
			return
		}
		fmt.Println("Command registration complete. Bot is ready!")
	} else {
		fmt.Println("Bot is ready! (Use --register-commands flag to register slash commands)")
	}
}

// interactionCreate handles interaction events
func (b *Bot) interactionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Create session wrapper for all commands that use SessionInterface
	sessionWrapper := manager.NewSessionWrapper(s)

	var err error
	switch i.ApplicationCommandData().Name {
	case "ping":
		err = commands.HandlePingCommand(sessionWrapper, i)
	case "peepee":
		err = commands.HandlePeepeeCommandWithReaction(s, i)
	case "8ball":
		err = commands.Handle8BallCommand(sessionWrapper, i)
	case "coinflip":
		err = commands.HandleCoinFlipCommand(sessionWrapper, i)
	case "server":
		err = commands.HandleServerCommand(sessionWrapper, i)
	case "user":
		err = commands.HandleUserCommand(sessionWrapper, i)
	case "weather":
		err = commands.HandleWeatherCommand(sessionWrapper, i)
	case "roll":
		err = commands.HandleRollCommand(sessionWrapper, i)
	case "join":
		err = commands.HandleJoinCommand(sessionWrapper, i)
	case "leave":
		err = commands.HandleLeaveCommand(sessionWrapper, i)
	case "play":
		err = commands.HandlePlayCommand(sessionWrapper, i)
	}

	if err != nil {
		log.Printf("Error handling command '%s': %v", i.ApplicationCommandData().Name, err)
	}
}

// voiceStateUpdate handles voice state change events
func (b *Bot) voiceStateUpdate(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
	// Only process if we have a music manager
	if commands.MusicManager == nil {
		return
	}

	// Cast to the interface that has OnVoiceStateUpdate
	if manager, ok := commands.MusicManager.(*manager.Manager); ok {
		manager.OnVoiceStateUpdate(vsu.GuildID)
	}
}

// Global flag for command registration (will be set from main)
var shouldRegisterCommands bool

// SetShouldRegisterCommands sets the global flag for command registration
func SetShouldRegisterCommands(value bool) {
	shouldRegisterCommands = value
}
