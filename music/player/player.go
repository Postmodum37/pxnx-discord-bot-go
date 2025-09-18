package player

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"

	"pxnx-discord-bot/music/types"
)

// DCAAudioPlayer implements the AudioPlayer interface for Discord voice connections
type DCAAudioPlayer struct {
	guildID       string
	voiceConn     *discordgo.VoiceConnection
	currentSource atomic.Pointer[types.AudioSource]
	volume        int64 // Use atomic for thread-safe access
	playing       int32 // Use atomic: 0 = stopped, 1 = playing
	paused        int32 // Use atomic: 0 = not paused, 1 = paused
	stopChan      chan struct{}
	pauseChan     chan struct{}
	resumeChan    chan struct{}
	mu            sync.RWMutex
	encoder       *dca.EncodeSession
	streamSession *dca.StreamingSession
}

// NewDCAAudioPlayer creates a new DCA-based audio player
func NewDCAAudioPlayer(guildID string, voiceConn *discordgo.VoiceConnection) *DCAAudioPlayer {
	return &DCAAudioPlayer{
		guildID:    guildID,
		voiceConn:  voiceConn,
		volume:     50, // Default volume 50%
		stopChan:   make(chan struct{}),
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
	}
}

// Play starts playing an audio source
func (p *DCAAudioPlayer) Play(ctx context.Context, source types.AudioSource) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop any current playback
	if err := p.stopInternal(); err != nil {
		// Continue with new playback even if stop failed
	}

	// Give a moment for the previous playback to fully stop
	time.Sleep(10 * time.Millisecond)

	// Reset channels
	p.stopChan = make(chan struct{})
	p.pauseChan = make(chan struct{})
	p.resumeChan = make(chan struct{})

	// Set the current source before marking as playing
	p.currentSource.Store(&source)

	// Mark as playing
	atomic.StoreInt32(&p.playing, 1)
	atomic.StoreInt32(&p.paused, 0)

	// Start playback in a goroutine
	go p.playbackLoop(ctx, source)

	return nil
}

// playbackLoop handles the actual audio playback
func (p *DCAAudioPlayer) playbackLoop(ctx context.Context, source types.AudioSource) {
	defer func() {
		atomic.StoreInt32(&p.playing, 0)
		atomic.StoreInt32(&p.paused, 0)
		p.currentSource.Store(nil)
		p.cleanupEncoder()
	}()

	// For now, we'll simulate playback since we need the full audio processing pipeline
	// In a real implementation, this would:
	// 1. Get audio stream from source.StreamURL
	// 2. Create DCA encoder with proper options
	// 3. Stream encoded audio to Discord voice connection
	// 4. Handle pause/resume/stop signals

	// Placeholder implementation that respects stop/pause signals
	ticker := time.NewTicker(100 * time.Millisecond) // Simulate audio frames
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-p.stopChan:
			return
		case <-p.pauseChan:
			atomic.StoreInt32(&p.paused, 1)
			// Wait for resume or stop
			select {
			case <-p.resumeChan:
				atomic.StoreInt32(&p.paused, 0)
				continue
			case <-p.stopChan:
				return
			case <-ctx.Done():
				return
			}
		case <-ticker.C:
			// Simulate processing audio frames
			if atomic.LoadInt32(&p.playing) == 0 {
				return
			}
		}
	}
}

// Pause pauses the current playback
func (p *DCAAudioPlayer) Pause() error {
	if !p.IsPlaying() {
		return fmt.Errorf("no audio is currently playing")
	}

	if p.IsPaused() {
		return fmt.Errorf("audio is already paused")
	}

	select {
	case p.pauseChan <- struct{}{}:
		return nil
	default:
		return fmt.Errorf("failed to pause audio")
	}
}

// Resume resumes paused playback
func (p *DCAAudioPlayer) Resume() error {
	if !p.IsPlaying() {
		return fmt.Errorf("no audio is currently playing")
	}

	if !p.IsPaused() {
		return fmt.Errorf("audio is not paused")
	}

	select {
	case p.resumeChan <- struct{}{}:
		return nil
	default:
		return fmt.Errorf("failed to resume audio")
	}
}

// Stop stops the current playback
func (p *DCAAudioPlayer) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.stopInternal()
}

// stopInternal stops playback without acquiring the mutex (internal use only)
func (p *DCAAudioPlayer) stopInternal() error {
	if atomic.LoadInt32(&p.playing) == 0 {
		return nil // Already stopped
	}

	// Mark as stopped first to prevent race conditions
	atomic.StoreInt32(&p.playing, 0)
	atomic.StoreInt32(&p.paused, 0)

	// Signal stop to playback loop (non-blocking)
	select {
	case p.stopChan <- struct{}{}:
	default:
		// Channel might be closed or full, that's okay
	}

	// Clear current source
	p.currentSource.Store(nil)

	// Cleanup resources
	p.cleanupEncoder()

	return nil
}

// SetVolume sets the playback volume (0-100)
func (p *DCAAudioPlayer) SetVolume(volume int) error {
	if volume < 0 || volume > 100 {
		return fmt.Errorf("volume must be between 0 and 100, got %d", volume)
	}

	atomic.StoreInt64(&p.volume, int64(volume))

	// In a real implementation, this would adjust the DCA encoder volume settings
	// or apply volume transformation to the audio stream
	return nil
}

// GetVolume returns the current volume (0-100)
func (p *DCAAudioPlayer) GetVolume() int {
	return int(atomic.LoadInt64(&p.volume))
}

// IsPlaying checks if audio is currently playing
func (p *DCAAudioPlayer) IsPlaying() bool {
	return atomic.LoadInt32(&p.playing) == 1
}

// IsPaused checks if audio is currently paused
func (p *DCAAudioPlayer) IsPaused() bool {
	return atomic.LoadInt32(&p.paused) == 1
}

// GetCurrentSource returns the currently playing audio source
func (p *DCAAudioPlayer) GetCurrentSource() *types.AudioSource {
	return p.currentSource.Load()
}

// cleanupEncoder cleans up DCA encoder resources
func (p *DCAAudioPlayer) cleanupEncoder() {
	if p.encoder != nil {
		p.encoder.Cleanup()
		p.encoder = nil
	}
	if p.streamSession != nil {
		p.streamSession = nil
	}
}

// TODO: Implement createEncoder for actual DCA encoding when audio streaming is added
// createEncoder creates a DCA encoder for the given audio source
// This is a placeholder for the full implementation
func (p *DCAAudioPlayer) createEncoder(source types.AudioSource) (*dca.EncodeSession, error) {
	// In a real implementation, this would:
	// 1. Create appropriate DCA encoding options based on Discord requirements
	// 2. Set up FFmpeg options for audio conversion
	// 3. Apply volume settings
	// 4. Return configured encoder session

	// Placeholder options for reference
	options := dca.StdEncodeOptions
	options.Volume = p.GetVolume()
	options.Channels = 2
	options.FrameRate = 48000
	options.FrameDuration = 20
	options.Bitrate = 128

	// This would normally use source.StreamURL as input
	return dca.EncodeFile(source.StreamURL, options)
}

// TODO: Implement streamToVoice for actual audio streaming when needed
// streamToVoice streams encoded audio to Discord voice connection
// This is a placeholder for the full implementation
func (p *DCAAudioPlayer) streamToVoice(encoder *dca.EncodeSession) error {
	// In a real implementation, this would:
	// 1. Create streaming session from encoder
	// 2. Handle voice connection speaking state
	// 3. Stream audio data with proper timing
	// 4. Handle interruptions and errors

	if p.voiceConn == nil {
		return fmt.Errorf("voice connection is nil")
	}

	// Start speaking
	if err := p.voiceConn.Speaking(true); err != nil {
		return fmt.Errorf("failed to start speaking: %w", err)
	}
	defer func() {
		if err := p.voiceConn.Speaking(false); err != nil {
			// Log error but don't fail
		}
	}()

	// Create streaming session
	done := make(chan error)
	_ = dca.NewStream(encoder, p.voiceConn, done)

	// Note: Volume control would be handled in the encoder options
	// The DCA streaming session doesn't have runtime volume control

	// Wait for completion or error
	return <-done
}

// Cleanup cleans up all player resources
func (p *DCAAudioPlayer) Cleanup() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Stop any current playback
	if err := p.stopInternal(); err != nil {
		return fmt.Errorf("failed to stop playback during cleanup: %w", err)
	}

	// Give a moment for playback loop to finish
	time.Sleep(50 * time.Millisecond)

	// Close channels safely
	select {
	case <-p.stopChan:
	default:
		close(p.stopChan)
	}

	select {
	case <-p.pauseChan:
	default:
		close(p.pauseChan)
	}

	select {
	case <-p.resumeChan:
	default:
		close(p.resumeChan)
	}

	// Clear voice connection reference
	p.voiceConn = nil

	return nil
}

// GetGuildID returns the guild ID this player is associated with
func (p *DCAAudioPlayer) GetGuildID() string {
	return p.guildID
}

// IsReady checks if the player is ready for playback
func (p *DCAAudioPlayer) IsReady() bool {
	return p.voiceConn != nil && p.voiceConn.Ready
}

// GetPlaybackDuration returns the duration of current playback (placeholder)
func (p *DCAAudioPlayer) GetPlaybackDuration() (current, total string) {
	// In a real implementation, this would track playback position
	// and return formatted duration strings
	return "0:00", "0:00"
}

// Ensure DCAAudioPlayer implements the AudioPlayer interface
var _ types.AudioPlayer = (*DCAAudioPlayer)(nil)
