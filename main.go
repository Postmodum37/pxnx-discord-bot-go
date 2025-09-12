package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/joho/godotenv"

	"pxnx-discord-bot/bot"
)

func main() {
	// Parse command line flags
	registerCommands := flag.Bool("register-commands", false, "Register bot commands with Discord (cleans up existing commands first)")
	flag.Parse()

	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	token := os.Getenv("DISCORD_BOT_TOKEN")
	if token == "" {
		log.Fatal("DISCORD_BOT_TOKEN environment variable is required")
	}

	// Set global flag for command registration
	bot.SetShouldRegisterCommands(*registerCommands)

	// Create new bot instance
	botInstance, err := bot.New(token)
	if err != nil {
		log.Fatal("Error creating bot:", err)
	}

	// Setup bot handlers and intents
	botInstance.Setup()

	// Start the bot
	err = botInstance.Start()
	if err != nil {
		log.Fatal("Error opening connection:", err)
	}
	defer func() {
		if err := botInstance.Stop(); err != nil {
			log.Printf("Error closing Discord session: %v", err)
		}
	}()

	fmt.Println("Bot is running. Press CTRL+C to exit.")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop

	fmt.Println("Gracefully shutting down.")
}
