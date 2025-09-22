package manager

import (
	"pxnx-discord-bot/music/providers"
)

// SetupYTDLPProvider creates and registers a yt-dlp provider with the manager
func (m *Manager) SetupYTDLPProvider() error {
	ytdlpProvider := providers.NewYouTubeYTDLPProvider()
	m.RegisterProvider(ytdlpProvider)
	return nil
}