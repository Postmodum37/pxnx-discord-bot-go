package testutils

import (
	"context"

	"pxnx-discord-bot/music/types"

	"github.com/bwmarrin/discordgo"
)

// MockMusicSession implements the music SessionInterface for testing
type MockMusicSession struct {
	*MockSession // Embed basic mock session

	// Voice connection mocking
	VoiceJoinCalled    bool
	VoiceJoinError     error
	VoiceJoinReturn    *discordgo.VoiceConnection
	GetVoiceConnCalled bool
	GetVoiceConnReturn *discordgo.VoiceConnection
}

// NewMockMusicSession creates a new mock music session
func NewMockMusicSession() *MockMusicSession {
	return &MockMusicSession{
		MockSession: &MockSession{},
	}
}

// ChannelVoiceJoin mocks the voice join functionality
func (m *MockMusicSession) ChannelVoiceJoin(guildID, channelID string, mute, deaf bool) (*discordgo.VoiceConnection, error) {
	m.VoiceJoinCalled = true
	if m.VoiceJoinError != nil {
		return nil, m.VoiceJoinError
	}

	// Create a minimal mock voice connection if none provided
	if m.VoiceJoinReturn == nil {
		m.VoiceJoinReturn = &discordgo.VoiceConnection{
			GuildID:   guildID,
			ChannelID: channelID,
			Ready:     true,
		}
	}

	return m.VoiceJoinReturn, nil
}

// GetVoiceConnection mocks getting voice connections
func (m *MockMusicSession) GetVoiceConnection(guildID string) *discordgo.VoiceConnection {
	m.GetVoiceConnCalled = true
	return m.GetVoiceConnReturn
}

// Reset resets all mock state including voice-specific state
func (m *MockMusicSession) Reset() {
	m.MockSession.Reset()
	m.VoiceJoinCalled = false
	m.VoiceJoinError = nil
	m.VoiceJoinReturn = nil
	m.GetVoiceConnCalled = false
	m.GetVoiceConnReturn = nil
}

// MockMusicManager implements the MusicManager interface for testing
type MockMusicManager struct {
	JoinChannelCalled     bool
	JoinChannelError      error
	JoinChannelGuildID    string
	JoinChannelChannelID  string
	LeaveChannelCalled    bool
	LeaveChannelError     error
	LeaveChannelGuildID   string
	IsConnectedCalled     bool
	IsConnectedReturn     bool
	PlayCalled            bool
	PlayError             error
	PauseCalled           bool
	PauseError            error
	ResumeCalled          bool
	ResumeError           error
	StopCalled            bool
	StopError             error
	SkipCalled            bool
	SkipError             error
	AddToQueueCalled      bool
	AddToQueueError       error
	GetQueueCalled        bool
	GetQueueReturn        []types.AudioSource
	GetQueueError         error
	RemoveFromQueueCalled bool
	RemoveFromQueueError  error
	ClearQueueCalled      bool
	ClearQueueError       error
	ShuffleQueueCalled    bool
	ShuffleQueueError     error
	SetVolumeCalled       bool
	SetVolumeError        error
	GetVolumeCalled       bool
	GetVolumeReturn       int
	GetVolumeError        error
	GetNowPlayingCalled   bool
	GetNowPlayingReturn   *types.AudioSource
	GetNowPlayingError    error
	GetPlayerStatusCalled bool
	GetPlayerStatusReturn types.PlayerStatus
	GetPlayerStatusError  error
	CleanupCalled         bool
	CleanupError          error
	GetProvidersCalled    bool
	GetProvidersReturn    []types.AudioProvider
}

// JoinChannel mocks joining a voice channel
func (m *MockMusicManager) JoinChannel(ctx context.Context, guildID, channelID string) error {
	m.JoinChannelCalled = true
	m.JoinChannelGuildID = guildID
	m.JoinChannelChannelID = channelID
	return m.JoinChannelError
}

// LeaveChannel mocks leaving a voice channel
func (m *MockMusicManager) LeaveChannel(ctx context.Context, guildID string) error {
	m.LeaveChannelCalled = true
	m.LeaveChannelGuildID = guildID
	return m.LeaveChannelError
}

// IsConnected mocks connection status check
func (m *MockMusicManager) IsConnected(guildID string) bool {
	m.IsConnectedCalled = true
	return m.IsConnectedReturn
}

// Play mocks playing audio
func (m *MockMusicManager) Play(ctx context.Context, guildID string, audioSource types.AudioSource) error {
	m.PlayCalled = true
	return m.PlayError
}

// Pause mocks pausing playback
func (m *MockMusicManager) Pause(ctx context.Context, guildID string) error {
	m.PauseCalled = true
	return m.PauseError
}

// Resume mocks resuming playback
func (m *MockMusicManager) Resume(ctx context.Context, guildID string) error {
	m.ResumeCalled = true
	return m.ResumeError
}

// Stop mocks stopping playback
func (m *MockMusicManager) Stop(ctx context.Context, guildID string) error {
	m.StopCalled = true
	return m.StopError
}

// Skip mocks skipping current song
func (m *MockMusicManager) Skip(ctx context.Context, guildID string) error {
	m.SkipCalled = true
	return m.SkipError
}

// AddToQueue mocks adding to queue
func (m *MockMusicManager) AddToQueue(ctx context.Context, guildID string, audioSource types.AudioSource) error {
	m.AddToQueueCalled = true
	return m.AddToQueueError
}

// GetQueue mocks getting the queue
func (m *MockMusicManager) GetQueue(ctx context.Context, guildID string) ([]types.AudioSource, error) {
	m.GetQueueCalled = true
	return m.GetQueueReturn, m.GetQueueError
}

// RemoveFromQueue mocks removing from queue
func (m *MockMusicManager) RemoveFromQueue(ctx context.Context, guildID string, position int) error {
	m.RemoveFromQueueCalled = true
	return m.RemoveFromQueueError
}

// ClearQueue mocks clearing the queue
func (m *MockMusicManager) ClearQueue(ctx context.Context, guildID string) error {
	m.ClearQueueCalled = true
	return m.ClearQueueError
}

// ShuffleQueue mocks shuffling the queue
func (m *MockMusicManager) ShuffleQueue(ctx context.Context, guildID string) error {
	m.ShuffleQueueCalled = true
	return m.ShuffleQueueError
}

// SetVolume mocks setting volume
func (m *MockMusicManager) SetVolume(ctx context.Context, guildID string, volume int) error {
	m.SetVolumeCalled = true
	return m.SetVolumeError
}

// GetVolume mocks getting volume
func (m *MockMusicManager) GetVolume(ctx context.Context, guildID string) (int, error) {
	m.GetVolumeCalled = true
	return m.GetVolumeReturn, m.GetVolumeError
}

// GetNowPlaying mocks getting current song
func (m *MockMusicManager) GetNowPlaying(ctx context.Context, guildID string) (*types.AudioSource, error) {
	m.GetNowPlayingCalled = true
	return m.GetNowPlayingReturn, m.GetNowPlayingError
}

// GetPlayerStatus mocks getting player status
func (m *MockMusicManager) GetPlayerStatus(ctx context.Context, guildID string) (types.PlayerStatus, error) {
	m.GetPlayerStatusCalled = true
	return m.GetPlayerStatusReturn, m.GetPlayerStatusError
}

// Cleanup mocks cleanup
func (m *MockMusicManager) Cleanup(ctx context.Context) error {
	m.CleanupCalled = true
	return m.CleanupError
}

// RegisterProvider mocks registering a provider
func (m *MockMusicManager) RegisterProvider(provider types.AudioProvider) {
	// For mock, we don't need to do anything
}

// GetProviders mocks getting all providers
func (m *MockMusicManager) GetProviders() []types.AudioProvider {
	m.GetProvidersCalled = true
	return m.GetProvidersReturn
}

// Reset resets all mock state
func (m *MockMusicManager) Reset() {
	m.JoinChannelCalled = false
	m.JoinChannelError = nil
	m.JoinChannelGuildID = ""
	m.JoinChannelChannelID = ""
	m.LeaveChannelCalled = false
	m.LeaveChannelError = nil
	m.LeaveChannelGuildID = ""
	m.IsConnectedCalled = false
	m.IsConnectedReturn = false
	m.PlayCalled = false
	m.PlayError = nil
	m.PauseCalled = false
	m.PauseError = nil
	m.ResumeCalled = false
	m.ResumeError = nil
	m.StopCalled = false
	m.StopError = nil
	m.SkipCalled = false
	m.SkipError = nil
	m.AddToQueueCalled = false
	m.AddToQueueError = nil
	m.GetQueueCalled = false
	m.GetQueueReturn = nil
	m.GetQueueError = nil
	m.RemoveFromQueueCalled = false
	m.RemoveFromQueueError = nil
	m.ClearQueueCalled = false
	m.ClearQueueError = nil
	m.ShuffleQueueCalled = false
	m.ShuffleQueueError = nil
	m.SetVolumeCalled = false
	m.SetVolumeError = nil
	m.GetVolumeCalled = false
	m.GetVolumeReturn = 0
	m.GetVolumeError = nil
	m.GetNowPlayingCalled = false
	m.GetNowPlayingReturn = nil
	m.GetNowPlayingError = nil
	m.GetPlayerStatusCalled = false
	m.GetPlayerStatusReturn = types.StatusIdle
	m.GetPlayerStatusError = nil
	m.CleanupCalled = false
	m.CleanupError = nil
	m.GetProvidersCalled = false
	m.GetProvidersReturn = nil
}
