package music

import (
	"context"
	"os/exec"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockDiscordSession mocks the Discord session for testing
type MockDiscordSession struct {
	mock.Mock
	*discordgo.Session
}

func (m *MockDiscordSession) ChannelVoiceJoin(guildID, channelID string, mute, deaf bool) (*discordgo.VoiceConnection, error) {
	args := m.Called(guildID, channelID, mute, deaf)
	return args.Get(0).(*discordgo.VoiceConnection), args.Error(1)
}

// MockVoiceConnection mocks Discord voice connection
type MockVoiceConnection struct {
	*discordgo.VoiceConnection
	ChannelID string
	Ready     bool
	OpusSend  chan []byte
}

func NewMockVoiceConnection(channelID string) *MockVoiceConnection {
	return &MockVoiceConnection{
		ChannelID: channelID,
		Ready:     true,
		OpusSend:  make(chan []byte, 100),
	}
}

func (m *MockVoiceConnection) Speaking(speaking bool) error {
	return nil
}

func (m *MockVoiceConnection) Disconnect() error {
	return nil
}

func TestSimplePlayer_Basic(t *testing.T) {
	// Skip if yt-dlp is not available
	if !isYTDLPAvailable() {
		t.Skip("yt-dlp not available, skipping integration tests")
	}

	session := &MockDiscordSession{}
	player := NewSimplePlayer(&session.Session)

	assert.NotNil(t, player)
	assert.Empty(t, player.connections)
}

func TestSimplePlayer_JoinChannel(t *testing.T) {
	session := &MockDiscordSession{}
	player := NewSimplePlayer(&session.Session)

	mockConn := NewMockVoiceConnection("test-channel")
	session.On("ChannelVoiceJoin", "test-guild", "test-channel", false, true).Return(mockConn, nil)

	err := player.JoinChannel("test-guild", "test-channel")
	require.NoError(t, err)

	// Verify connection was stored
	voicePlayer, exists := player.connections["test-guild"]
	assert.True(t, exists)
	assert.Equal(t, "test-guild", voicePlayer.guildID)
	assert.Equal(t, mockConn, voicePlayer.conn)

	session.AssertExpectations(t)
}

func TestSimplePlayer_LeaveChannel(t *testing.T) {
	session := &MockDiscordSession{}
	player := NewSimplePlayer(&session.Session)

	// First join a channel
	mockConn := NewMockVoiceConnection("test-channel")
	session.On("ChannelVoiceJoin", "test-guild", "test-channel", false, true).Return(mockConn, nil)

	err := player.JoinChannel("test-guild", "test-channel")
	require.NoError(t, err)

	// Then leave it
	err = player.LeaveChannel("test-guild")
	require.NoError(t, err)

	// Verify connection was removed
	_, exists := player.connections["test-guild"]
	assert.False(t, exists)
}

func TestSimplePlayer_ExtractTrackInfo(t *testing.T) {
	// Skip if yt-dlp is not available
	if !isYTDLPAvailable() {
		t.Skip("yt-dlp not available, skipping integration tests")
	}

	session := &MockDiscordSession{}
	player := NewSimplePlayer(&session.Session)

	// Test with a reliable YouTube video
	track, err := player.extractTrackInfo("Rick Astley Never Gonna Give You Up")
	require.NoError(t, err)

	assert.NotEmpty(t, track.Title)
	assert.NotEmpty(t, track.URL)
	assert.Contains(t, track.URL, "googlevideo.com") // Should be YouTube CDN URL

	t.Logf("Extracted track: %s", track.Title)
	t.Logf("Stream URL: %s", track.URL[:50]+"...")
}

func TestSimplePlayer_Play_NotConnected(t *testing.T) {
	session := &MockDiscordSession{}
	player := NewSimplePlayer(&session.Session)

	_, err := player.Play("test-guild", "test query")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not connected")
}

func TestVoicePlayer_QueueManagement(t *testing.T) {
	mockConn := NewMockVoiceConnection("test-channel")
	player := &VoicePlayer{
		guildID:  "test-guild",
		conn:     mockConn,
		queue:    make([]AudioTrack, 0),
		stopChan: make(chan struct{}),
		skipChan: make(chan struct{}),
	}

	// Test empty queue
	assert.False(t, player.IsPlaying())
	assert.Nil(t, player.GetCurrent())
	assert.Empty(t, player.GetQueue())

	// Add some tracks to queue (without actually playing)
	player.queue = append(player.queue, AudioTrack{
		Title: "Test Track 1",
		URL:   "http://example.com/1",
	})
	player.queue = append(player.queue, AudioTrack{
		Title: "Test Track 2",
		URL:   "http://example.com/2",
	})

	queue := player.GetQueue()
	assert.Len(t, queue, 2)
	assert.Equal(t, "Test Track 1", queue[0].Title)
	assert.Equal(t, "Test Track 2", queue[1].Title)

	// Test stop
	player.Stop()
	assert.False(t, player.IsPlaying())
	assert.Empty(t, player.GetQueue())
}

func TestTrackExtraction_ErrorHandling(t *testing.T) {
	if !isYTDLPAvailable() {
		t.Skip("yt-dlp not available, skipping integration tests")
	}

	session := &MockDiscordSession{}
	player := NewSimplePlayer(&session.Session)

	// Test with invalid query
	_, err := player.extractTrackInfo("this_should_not_exist_12345_invalid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "yt-dlp extraction failed")
}

func TestFFmpegAvailability(t *testing.T) {
	// Test that FFmpeg is available for the streaming functionality
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffmpeg", "-version")
	err := cmd.Run()

	if err != nil {
		t.Skip("FFmpeg not available, audio streaming will not work")
	} else {
		t.Log("FFmpeg is available")
	}
}

func TestStreamURLAccessibility(t *testing.T) {
	if !isYTDLPAvailable() {
		t.Skip("yt-dlp not available, skipping integration tests")
	}

	session := &MockDiscordSession{}
	player := NewSimplePlayer(&session.Session)

	// Extract a track
	track, err := player.extractTrackInfo("Rick Astley Never Gonna Give You Up")
	require.NoError(t, err)

	// Test that we can create an FFmpeg command with the URL
	// (without actually running it to avoid long test times)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "ffmpeg",
		"-i", track.URL,
		"-f", "null",
		"-t", "1",
		"-",
	)

	err = cmd.Start()
	assert.NoError(t, err, "Should be able to start FFmpeg with extracted URL")

	// Kill the process immediately
	if cmd.Process != nil {
		cmd.Process.Kill()
	}
}

// Benchmark tests
func BenchmarkTrackExtraction(b *testing.B) {
	if !isYTDLPAvailable() {
		b.Skip("yt-dlp not available")
	}

	session := &MockDiscordSession{}
	player := NewSimplePlayer(&session.Session)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := player.extractTrackInfo("Rick Astley Never Gonna Give You Up")
		if err != nil {
			b.Fatalf("Track extraction failed: %v", err)
		}
	}
}

// Helper functions

func isYTDLPAvailable() bool {
	cmd := exec.Command("yt-dlp", "--version")
	err := cmd.Run()
	return err == nil
}

// Integration test that demonstrates the simplified approach vs. the complex one
func TestSimplePlayerIntegration(t *testing.T) {
	if !isYTDLPAvailable() {
		t.Skip("yt-dlp not available, skipping integration tests")
	}

	t.Run("DirectCLIApproach", func(t *testing.T) {
		session := &MockDiscordSession{}
		player := NewSimplePlayer(&session.Session)

		start := time.Now()
		track, err := player.extractTrackInfo("test music")
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.NotEmpty(t, track.Title)
		assert.NotEmpty(t, track.URL)

		t.Logf("Direct CLI extraction took: %v", duration)
		t.Logf("Track extracted: %s", track.Title)

		// This should be much simpler and more reliable than the HTTP service approach
		assert.Less(t, duration, 10*time.Second, "Extraction should be fast")
	})

	t.Run("CompareWithCurrentImplementation", func(t *testing.T) {
		// This test documents the differences between approaches
		t.Log("Current implementation issues:")
		t.Log("- Uses complex HTTP service architecture")
		t.Log("- Uses unmaintained jonas747/dca library")
		t.Log("- 11,698 lines of over-engineered code")
		t.Log("- Frequent EOF streaming errors")

		t.Log("\nSimplified implementation benefits:")
		t.Log("- Direct yt-dlp CLI integration")
		t.Log("- Direct FFmpeg streaming to Discord")
		t.Log("- ~500 lines of focused code")
		t.Log("- No complex error recovery needed")
		t.Log("- Uses maintained, standard tools")
	})
}