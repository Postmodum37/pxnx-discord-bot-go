package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/joho/godotenv"

	"pxnx-discord-bot/bot"
	"pxnx-discord-bot/utils"
)

func main() {
	// Parse command line flags
	registerCommands := flag.Bool("register-commands", false, "Register bot commands with Discord (cleans up existing commands first)")
	logLevel := flag.String("log-level", "info", "Set log level (error, warn, info, debug)")
	flag.Parse()

	// Initialize logger
	if err := utils.InitLogger("logs", utils.GetLogLevelFromString(*logLevel)); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	defer utils.CloseLogger()

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		utils.LogInfo("No .env file found, using system environment variables")
	}

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		utils.LogError("DISCORD_BOT_TOKEN environment variable is required")
		os.Exit(1)
	}

	// Set global flag for command registration
	bot.SetShouldRegisterCommands(*registerCommands)

	// Create new bot instance
	botInstance, err := bot.New(token)
	if err != nil {
		utils.LogError("Error creating bot: %v", err)
		os.Exit(1)
	}

	// Setup bot handlers and intents
	botInstance.Setup()

	// Start the bot
	err = botInstance.Start()
	if err != nil {
		utils.LogError("Error opening connection: %v", err)
		os.Exit(1)
	}
	defer func() {
		if err := botInstance.Stop(); err != nil {
			utils.LogError("Error closing Discord session: %v", err)
		}
	}()

	fmt.Println("Bot is running. Press CTRL+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	utils.LogInfo("Gracefully shutting down")
	fmt.Println("Gracefully shutting down.")
}
