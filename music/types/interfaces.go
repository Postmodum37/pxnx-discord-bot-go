package types

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

// SessionInterface defines the methods needed for music functionality specifically
// This interface includes both basic Discord functionality and voice capabilities
type SessionInterface interface {
	// Basic Discord functionality
	InteractionRespond(interaction *discordgo.Interaction, resp *discordgo.InteractionResponse, options ...discordgo.RequestOption) error
	FollowupMessageCreate(interaction *discordgo.Interaction, wait bool, data *discordgo.WebhookParams, options ...discordgo.RequestOption) (*discordgo.Message, error)
	Guild(guildID string, options ...discordgo.RequestOption) (*discordgo.Guild, error)

	// Voice functionality for music commands
	ChannelVoiceJoin(guildID, channelID string, mute, deaf bool) (*discordgo.VoiceConnection, error)

	// Access to voice connections - we'll need to access the field through a wrapper
	GetVoiceConnection(guildID string) *discordgo.VoiceConnection
}

// MusicManager defines the main interface for music management
type MusicManager interface {
	// Provider management
	RegisterProvider(provider AudioProvider)
	GetProviders() []AudioProvider

	// Voice channel management
	JoinChannel(ctx context.Context, guildID, channelID string) error
	LeaveChannel(ctx context.Context, guildID string) error
	IsConnected(guildID string) bool

	// Playback control
	Play(ctx context.Context, guildID string, audioSource AudioSource) error
	Pause(ctx context.Context, guildID string) error
	Resume(ctx context.Context, guildID string) error
	Stop(ctx context.Context, guildID string) error
	Skip(ctx context.Context, guildID string) error

	// Queue management
	AddToQueue(ctx context.Context, guildID string, audioSource AudioSource) error
	GetQueue(ctx context.Context, guildID string) ([]AudioSource, error)
	RemoveFromQueue(ctx context.Context, guildID string, position int) error
	ClearQueue(ctx context.Context, guildID string) error
	ShuffleQueue(ctx context.Context, guildID string) error

	// Volume and status
	SetVolume(ctx context.Context, guildID string, volume int) error
	GetVolume(ctx context.Context, guildID string) (int, error)
	GetNowPlaying(ctx context.Context, guildID string) (*AudioSource, error)
	GetPlayerStatus(ctx context.Context, guildID string) (PlayerStatus, error)

	// Cleanup
	Cleanup(ctx context.Context) error
}

// AudioPlayer defines the interface for audio playback
type AudioPlayer interface {
	Play(ctx context.Context, source AudioSource) error
	Pause() error
	Resume() error
	Stop() error
	SetVolume(volume int) error
	GetVolume() int
	IsPlaying() bool
	IsPaused() bool
	GetCurrentSource() *AudioSource
}

// Queue defines the interface for queue management
type Queue interface {
	Add(source AudioSource)
	Remove(position int) error
	Get(position int) (*AudioSource, error)
	GetAll() []AudioSource
	Clear()
	Shuffle()
	Next() (*AudioSource, bool)
	Size() int
	IsEmpty() bool
}

// AudioProvider defines the interface for audio source providers (YouTube, etc.)
type AudioProvider interface {
	GetAudioSource(ctx context.Context, query string) (*AudioSource, error)
	Search(ctx context.Context, query string, maxResults int) ([]AudioSource, error)
	SupportsURL(url string) bool
	GetProviderName() string
}

// PlayerStatus represents the current state of a music player
type PlayerStatus int

const (
	StatusIdle PlayerStatus = iota
	StatusPlaying
	StatusPaused
	StatusStopped
	StatusBuffering
	StatusError
)

func (s PlayerStatus) String() string {
	switch s {
	case StatusIdle:
		return "idle"
	case StatusPlaying:
		return "playing"
	case StatusPaused:
		return "paused"
	case StatusStopped:
		return "stopped"
	case StatusBuffering:
		return "buffering"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}

// AudioSource represents a playable audio source
type AudioSource struct {
	Title       string
	URL         string
	Duration    string
	Thumbnail   string
	Provider    string
	RequestedBy string
	StreamURL   string                 // The actual streaming URL for playback
	Metadata    map[string]interface{} // Additional metadata for provider-specific data
}

// VoiceChannelError represents voice channel specific errors
type VoiceChannelError struct {
	Type    string
	Message string
	GuildID string
}

func (e *VoiceChannelError) Error() string {
	return e.Message
}

// MusicError represents music system specific errors
type MusicError struct {
	Type    string
	Message string
	Err     error
}

func (e *MusicError) Error() string {
	if e.Err != nil {
		return e.Message + ": " + e.Err.Error()
	}
	return e.Message
}

func (e *MusicError) Unwrap() error {
	return e.Err
}
