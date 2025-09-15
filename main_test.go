package main

import (
	"testing"

	"pxnx-discord-bot/bot"
)

func TestMainPackageIntegration(t *testing.T) {
	// Test that main package can access bot functionality

	t.Run("can access bot commands", func(t *testing.T) {
		commands := bot.GetCommands()
		if len(commands) == 0 {
			t.Error("Expected to be able to access bot commands from main package")
		}
	})

	t.Run("can create bot instance", func(t *testing.T) {
		botInstance, err := bot.New("test.token")
		if err != nil {
			t.Errorf("Expected to create bot from main package, got error: %v", err)
		}
		if botInstance == nil {
			t.Error("Expected non-nil bot instance")
		}
	})

	t.Run("can set command registration flag", func(t *testing.T) {
		// This should not panic
		bot.SetShouldRegisterCommands(true)
		bot.SetShouldRegisterCommands(false)
	})
}
