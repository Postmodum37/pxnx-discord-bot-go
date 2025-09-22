package player

import (
	"context"
	"testing"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"

	"pxnx-discord-bot/music/types"
)

// mockVoiceConnection creates a mock voice connection for testing
func mockVoiceConnection() *discordgo.VoiceConnection {
	return &discordgo.VoiceConnection{
		GuildID:   "test-guild-123",
		ChannelID: "test-channel-456",
		Ready:     true,
	}
}

// createTestAudioSource creates a test audio source
func createTestAudioSource(title string) types.AudioSource {
	return types.AudioSource{
		Title:       title,
		URL:         "https://youtube.com/watch?v=test",
		Duration:    "3:45",
		Thumbnail:   "https://example.com/thumb.jpg",
		Provider:    "youtube",
		RequestedBy: "user123",
		StreamURL:   "https://example.com/stream.opus",
	}
}

func TestNewDCAAudioPlayer(t *testing.T) {
	guildID := "test-guild-123"
	vc := mockVoiceConnection()

	player := NewDCAAudioPlayerForTesting(guildID, vc)

	assert.NotNil(t, player)
	assert.Equal(t, guildID, player.GetGuildID())
	assert.Equal(t, 50, player.GetVolume()) // Default volume
	assert.False(t, player.IsPlaying())
	assert.False(t, player.IsPaused())
	assert.Nil(t, player.GetCurrentSource())
	assert.True(t, player.IsReady())
}

func TestAudioPlayerVolume(t *testing.T) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())

	// Test default volume
	assert.Equal(t, 50, player.GetVolume())

	// Test setting valid volumes
	err := player.SetVolume(75)
	assert.NoError(t, err)
	assert.Equal(t, 75, player.GetVolume())

	err = player.SetVolume(0)
	assert.NoError(t, err)
	assert.Equal(t, 0, player.GetVolume())

	err = player.SetVolume(100)
	assert.NoError(t, err)
	assert.Equal(t, 100, player.GetVolume())

	// Test invalid volumes
	err = player.SetVolume(-1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "volume must be between 0 and 100")

	err = player.SetVolume(101)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "volume must be between 0 and 100")

	// Volume should remain unchanged after invalid attempts
	assert.Equal(t, 100, player.GetVolume())
}

func TestAudioPlayerPlay(t *testing.T) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())
	source := createTestAudioSource("test-song")
	ctx := context.Background()

	// Initially not playing
	assert.False(t, player.IsPlaying())
	assert.Nil(t, player.GetCurrentSource())

	// Start playing
	err := player.Play(ctx, source)
	assert.NoError(t, err)

	// Give a moment for the goroutine to start
	time.Sleep(10 * time.Millisecond)

	assert.True(t, player.IsPlaying())
	assert.False(t, player.IsPaused())

	currentSource := player.GetCurrentSource()
	assert.NotNil(t, currentSource)
	assert.Equal(t, "test-song", currentSource.Title)
}

func TestAudioPlayerStop(t *testing.T) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())
	source := createTestAudioSource("test-song")
	ctx := context.Background()

	// Start playing
	err := player.Play(ctx, source)
	assert.NoError(t, err)

	// Give a moment for playback to start
	time.Sleep(10 * time.Millisecond)
	assert.True(t, player.IsPlaying())

	// Stop playback
	err = player.Stop()
	assert.NoError(t, err)

	// Give a moment for stop to process
	time.Sleep(10 * time.Millisecond)

	assert.False(t, player.IsPlaying())
	assert.False(t, player.IsPaused())
	assert.Nil(t, player.GetCurrentSource())
}

func TestAudioPlayerPauseResume(t *testing.T) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())
	source := createTestAudioSource("test-song")
	ctx := context.Background()

	// Test pause without playing (should error)
	err := player.Pause()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no audio is currently playing")

	// Test resume without playing (should error)
	err = player.Resume()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no audio is currently playing")

	// Start playing
	err = player.Play(ctx, source)
	assert.NoError(t, err)

	// Give a moment for playback to start
	time.Sleep(10 * time.Millisecond)
	assert.True(t, player.IsPlaying())
	assert.False(t, player.IsPaused())

	// Pause playback
	err = player.Pause()
	assert.NoError(t, err)

	// Give a moment for pause to process
	time.Sleep(10 * time.Millisecond)
	assert.True(t, player.IsPlaying()) // Still "playing" but paused
	assert.True(t, player.IsPaused())

	// Test double pause (should error)
	err = player.Pause()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audio is already paused")

	// Resume playback
	err = player.Resume()
	assert.NoError(t, err)

	// Give a moment for resume to process
	time.Sleep(10 * time.Millisecond)
	assert.True(t, player.IsPlaying())
	assert.False(t, player.IsPaused())

	// Test double resume (should error)
	err = player.Resume()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "audio is not paused")
}

func TestAudioPlayerMultiplePlayCalls(t *testing.T) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())
	source1 := createTestAudioSource("song1")
	source2 := createTestAudioSource("song2")
	ctx := context.Background()

	// Play first song
	err := player.Play(ctx, source1)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	assert.True(t, player.IsPlaying())
	assert.Equal(t, "song1", player.GetCurrentSource().Title)

	// Play second song (should stop first and start second)
	err = player.Play(ctx, source2)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	assert.True(t, player.IsPlaying())
	assert.Equal(t, "song2", player.GetCurrentSource().Title)
}

func TestAudioPlayerCleanup(t *testing.T) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())
	source := createTestAudioSource("test-song")
	ctx := context.Background()

	// Start playing
	err := player.Play(ctx, source)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	assert.True(t, player.IsPlaying())

	// Cleanup should stop playback and clean resources
	err = player.Cleanup()
	assert.NoError(t, err)

	assert.False(t, player.IsPlaying())
	assert.False(t, player.IsPaused())
	assert.Nil(t, player.GetCurrentSource())
	assert.False(t, player.IsReady()) // Voice connection should be cleared
}

func TestAudioPlayerContextCancellation(t *testing.T) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())
	source := createTestAudioSource("test-song")

	// Create a context that we can cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Start playing
	err := player.Play(ctx, source)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)
	assert.True(t, player.IsPlaying())

	// Cancel the context
	cancel()

	// Give time for the cancellation to be processed
	time.Sleep(20 * time.Millisecond)

	// Playback should have stopped due to context cancellation
	assert.False(t, player.IsPlaying())
	assert.Nil(t, player.GetCurrentSource())
}

func TestAudioPlayerWithNilVoiceConnection(t *testing.T) {
	player := NewDCAAudioPlayerForTesting("test-guild", nil)

	assert.False(t, player.IsReady())
	assert.Equal(t, "test-guild", player.GetGuildID())

	// Should still be able to set volume and check state
	err := player.SetVolume(75)
	assert.NoError(t, err)
	assert.Equal(t, 75, player.GetVolume())

	assert.False(t, player.IsPlaying())
	assert.False(t, player.IsPaused())
}

func TestAudioPlayerConcurrentOperations(t *testing.T) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())
	source := createTestAudioSource("test-song")
	ctx := context.Background()

	// Start playing
	err := player.Play(ctx, source)
	assert.NoError(t, err)

	time.Sleep(10 * time.Millisecond)

	// Test concurrent volume changes
	for i := 0; i < 10; i++ {
		go func(vol int) {
			player.SetVolume(vol * 10)
		}(i)
	}

	// Test concurrent state checks
	for i := 0; i < 10; i++ {
		go func() {
			player.IsPlaying()
			player.IsPaused()
			player.GetCurrentSource()
			player.GetVolume()
		}()
	}

	time.Sleep(50 * time.Millisecond)

	// Should still be playing and responsive
	assert.True(t, player.IsPlaying())

	// Cleanup
	err = player.Stop()
	assert.NoError(t, err)
}

func TestAudioPlayerInterfaceCompliance(t *testing.T) {
	// Compile-time check that DCAAudioPlayer implements AudioPlayer interface
	var _ types.AudioPlayer = (*DCAAudioPlayer)(nil)

	// Runtime check with actual instance
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())
	var audioPlayer types.AudioPlayer = player

	assert.NotNil(t, audioPlayer)
	assert.Equal(t, 50, audioPlayer.GetVolume())
	assert.False(t, audioPlayer.IsPlaying())
	assert.False(t, audioPlayer.IsPaused())
}

// BenchmarkAudioPlayerVolumeOperations benchmarks volume get/set operations
func BenchmarkAudioPlayerVolumeOperations(b *testing.B) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())

	b.Run("SetVolume", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			player.SetVolume(i % 101) // 0-100
		}
	})

	b.Run("GetVolume", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			player.GetVolume()
		}
	})
}

// BenchmarkAudioPlayerStateChecks benchmarks state checking operations
func BenchmarkAudioPlayerStateChecks(b *testing.B) {
	player := NewDCAAudioPlayerForTesting("test-guild", mockVoiceConnection())

	b.Run("IsPlaying", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			player.IsPlaying()
		}
	})

	b.Run("IsPaused", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			player.IsPaused()
		}
	})

	b.Run("GetCurrentSource", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			player.GetCurrentSource()
		}
	})
}
