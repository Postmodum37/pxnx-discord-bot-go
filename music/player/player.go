package player

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"

	"pxnx-discord-bot/music/types"
	"pxnx-discord-bot/utils"
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
	testMode      bool // Set to true for testing to skip actual encoding/streaming
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
		testMode:   false,
	}
}

// NewDCAAudioPlayerForTesting creates a new DCA-based audio player for testing
// This version skips actual audio encoding and streaming
func NewDCAAudioPlayerForTesting(guildID string, voiceConn *discordgo.VoiceConnection) *DCAAudioPlayer {
	return &DCAAudioPlayer{
		guildID:    guildID,
		voiceConn:  voiceConn,
		volume:     50, // Default volume 50%
		stopChan:   make(chan struct{}),
		pauseChan:  make(chan struct{}),
		resumeChan: make(chan struct{}),
		testMode:   true,
	}
}

// Play starts playing an audio source
func (p *DCAAudioPlayer) Play(ctx context.Context, source types.AudioSource) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	utils.LogInfo("Starting playback: %s (Guild: %s)", source.Title, p.guildID)

	// Validate voice connection before playing
	if err := p.validateVoiceConnection(); err != nil {
		utils.LogError("Voice connection validation failed: %v", err)
		return fmt.Errorf("voice connection validation failed: %w", err)
	}

	// Stop any current playback
	if err := p.stopInternal(); err != nil {
		utils.LogWarn("Error stopping previous playback: %v", err)
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

	utils.LogInfo("Playback started: %s", source.Title)
	return nil
}

// playbackLoop handles the actual audio playback
func (p *DCAAudioPlayer) playbackLoop(ctx context.Context, source types.AudioSource) {
	defer func() {
		utils.LogInfo("Playback finished: %s", source.Title)
		atomic.StoreInt32(&p.playing, 0)
		atomic.StoreInt32(&p.paused, 0)
		p.currentSource.Store(nil)
		p.cleanupEncoder()
	}()

	utils.LogInfo("Playback started: %s", source.Title)

	// In test mode, simulate playback without actual encoding/streaming
	if p.testMode {
		p.testModePlayback(ctx)
		return
	}

	// Create DCA encoder for the audio source
	utils.LogDebug("Creating encoder for: %s", source.Title)
	encoder, err := p.createEncoder(source)
	if err != nil {
		utils.LogError("Failed to create encoder: %v", err)
		return
	}
	utils.LogDebug("Encoder created for: %s", source.Title)
	defer encoder.Cleanup()

	// Store encoder for potential cleanup
	p.encoder = encoder

	// Start streaming in a separate goroutine
	utils.LogDebug("Starting streaming for: %s", source.Title)
	streamDone := make(chan error, 1)
	go func() {
		utils.LogDebug("Streaming started for: %s", source.Title)
		streamDone <- p.streamToVoice(encoder)
		utils.LogDebug("Streaming finished for: %s", source.Title)
	}()

	// Main control loop - handle pause/resume/stop signals while streaming
	for {
		select {
		case <-ctx.Done():
			utils.LogDebug("Context cancelled for: %s", source.Title)
			return
		case <-p.stopChan:
			utils.LogDebug("Stop signal received for: %s", source.Title)
			return
		case err := <-streamDone:
			if err != nil {
				utils.LogError("Streaming error for %s: %v", source.Title, err)
			} else {
				utils.LogInfo("Streaming completed: %s", source.Title)
			}
			// Stream completed (success or error)
			return
		case <-p.pauseChan:
			atomic.StoreInt32(&p.paused, 1)
			// Set voice to not speaking when paused
			if p.voiceConn != nil {
				p.voiceConn.Speaking(false)
			}
			// Wait for resume or stop
			select {
			case <-p.resumeChan:
				atomic.StoreInt32(&p.paused, 0)
				// Resume speaking when unpaused
				if p.voiceConn != nil {
					p.voiceConn.Speaking(true)
				}
				continue
			case <-p.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	}
}

// testModePlayback simulates playback for testing without actual audio processing
func (p *DCAAudioPlayer) testModePlayback(ctx context.Context) {
	// Simulate playback with pause/resume/stop support
	ticker := time.NewTicker(100 * time.Millisecond)
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

	// Note: Volume changes during active playback require stopping and restarting
	// the stream with new encoder settings. DCA encoder volume is set at creation time.
	// For real-time volume control, you would need to:
	// 1. Stop current playback
	// 2. Create new encoder with updated volume
	// 3. Restart playback from current position
	// This is a limitation of the DCA library design.

	// For now, volume changes take effect on the next track
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

// createEncoder creates a DCA encoder for the given audio source
func (p *DCAAudioPlayer) createEncoder(source types.AudioSource) (*dca.EncodeSession, error) {
	utils.LogDebug("Creating encoder for: %s", source.Title)
	utils.LogDebug("Stream URL: %s", source.StreamURL)

	if source.StreamURL == "" {
		utils.LogError("Stream URL is empty for: %s", source.Title)
		return nil, fmt.Errorf("stream URL is empty")
	}

	// Validate stream URL accessibility before DCA processing
	utils.LogDebug("Testing stream URL accessibility...")
	if err := p.validateStreamURL(source.StreamURL); err != nil {
		utils.LogError("Stream URL validation failed: %v", err)
		return nil, fmt.Errorf("stream URL validation failed: %w", err)
	}
	utils.LogDebug("Stream URL is accessible")

	// Create DCA encoding options optimized for Discord and YouTube streams
	utils.LogDebug("Setting up DCA encoder options")
	options := dca.StdEncodeOptions

	// Basic audio settings for Discord compatibility
	options.Volume = p.GetVolume()
	options.Channels = 2          // Stereo
	options.FrameRate = 48000     // Discord requires 48kHz
	options.FrameDuration = 20    // 20ms frames
	options.Bitrate = 96          // Higher bitrate for better quality
	options.Application = "audio" // Optimized for music

	// Enhanced settings for better opus/webm compatibility
	options.CompressionLevel = 3  // Less compression for compatibility
	options.PacketLoss = 1        // Allow some packet loss tolerance
	options.BufferedFrames = 200  // Larger buffer for stability (increased from 100)
	options.VBR = false           // Constant bitrate for stability

	// Disable raw output to let DCA handle format conversion
	options.RawOutput = false
	options.Threads = 2           // Use more threads for better performance

	// Set start time to beginning of stream
	options.StartTime = 0

	// Add comprehensive audio filter for format normalization and stability
	// Includes error resilience and format conversion
	options.AudioFilter = "aformat=sample_fmts=s16:channel_layouts=stereo:sample_rates=48000,aresample=48000"

	// Note: The DCA library automatically includes FFmpeg reconnect options:
	// -reconnect 1, -reconnect_at_eof 1, -reconnect_streamed 1, -reconnect_delay_max 2
	// These built-in options help prevent EOF errors from network interruptions

	utils.LogDebug("DCA encoder options: bitrate=%d, channels=%d, framerate=%d (with built-in reconnect)",
		options.Bitrate, options.Channels, options.FrameRate)

	// Add detailed encoder creation monitoring
	utils.LogDebug("Creating DCA encoder with options: bitrate=%d, channels=%d, framerate=%d", options.Bitrate, options.Channels, options.FrameRate)

	// Create encoder session from the stream URL with error handling and fallback
	utils.LogDebug("Creating DCA encoder with URL: %s", source.StreamURL)
	encoder, err := dca.EncodeFile(source.StreamURL, options)
	if err != nil {
		utils.LogError("DCA encoder creation failed: %v", err)
		utils.LogDebug("DCA error type: %T", err)
		if encoder != nil {
			utils.LogDebug("Cleaning up encoder returned despite error")
			encoder.Cleanup()
		}

		// Try fallback with alternative format if possible
		if alternativeSource := p.tryAlternativeFormat(source, err); alternativeSource != nil {
			utils.LogInfo("Retrying with alternative format: %s", alternativeSource.StreamURL)
			return p.createEncoder(*alternativeSource)
		}

		// Provide more specific error messages based on common failure types
		errorMsg := err.Error()
		switch {
		case contains(errorMsg, "No such file or directory"):
			return nil, fmt.Errorf("audio stream URL is not accessible: %w", err)
		case contains(errorMsg, "Invalid data found"):
			return nil, fmt.Errorf("invalid or corrupted audio format: %w", err)
		case contains(errorMsg, "Connection refused") || contains(errorMsg, "timeout"):
			return nil, fmt.Errorf("network connection failed - please check your internet connection: %w", err)
		case contains(errorMsg, "403") || contains(errorMsg, "Forbidden"):
			return nil, fmt.Errorf("access denied to audio stream - the video may be private or region-locked: %w", err)
		case contains(errorMsg, "404") || contains(errorMsg, "Not Found"):
			return nil, fmt.Errorf("audio stream not found - the video may have been deleted: %w", err)
		case contains(errorMsg, "429") || contains(errorMsg, "Too Many Requests"):
			return nil, fmt.Errorf("rate limited by video provider - please try again later: %w", err)
		default:
			return nil, fmt.Errorf("failed to create audio encoder: %w", err)
		}
	}

	utils.LogDebug("DCA encoder created successfully")
	if encoder == nil {
		utils.LogError("Encoder is nil despite successful creation")
		return nil, fmt.Errorf("encoder is nil despite successful creation")
	}

	// DEBUGGING: Test encoder readiness (disabled by default to avoid consuming frames)
	// Uncomment the following lines to enable encoder reading test:
	/*
	log.Printf("[PLAYER] DEBUG: Encoder created, checking if it's reading data...")
	if err := p.testEncoderReading(encoder); err != nil {
		log.Printf("[PLAYER] DEBUG: Encoder reading test failed: %v", err)
		encoder.Cleanup()
		return nil, fmt.Errorf("encoder reading test failed: %w", err)
	}
	log.Printf("[PLAYER] DEBUG: Encoder is reading data successfully")
	*/

	utils.LogDebug("Encoder creation completed")
	return encoder, nil
}

// contains is a helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// validateStreamURL validates that a stream URL is accessible and contains audio data
func (p *DCAAudioPlayer) validateStreamURL(url string) error {
	utils.LogDebug("Validating stream URL: %s", url)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	// Create HEAD request to check availability without downloading
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add User-Agent to avoid blocking
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Discord Music Bot)")

	utils.LogDebug("Sending HEAD request to: %s", url)
	resp, err := client.Do(req)
	if err != nil {
		utils.LogDebug("HEAD request failed: %v", err)
		return fmt.Errorf("failed to access stream URL: %w", err)
	}
	defer resp.Body.Close()

	utils.LogDebug("HEAD response status: %s", resp.Status)
	utils.LogDebug("Content-Type: %s", resp.Header.Get("Content-Type"))
	utils.LogDebug("Content-Length: %s", resp.Header.Get("Content-Length"))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("stream URL returned non-success status: %s", resp.Status)
	}

	// Check if content type indicates audio
	contentType := resp.Header.Get("Content-Type")
	if contentType != "" && !strings.Contains(contentType, "audio") && !strings.Contains(contentType, "video") {
		utils.LogWarn("Content-Type doesn't indicate audio/video: %s", contentType)
	}

	utils.LogDebug("Stream URL validation successful")
	return nil
}

// testEncoderReading attempts to read a small amount of data from the encoder to verify it's working
func (p *DCAAudioPlayer) testEncoderReading(encoder *dca.EncodeSession) error {
	utils.LogDebug("Testing encoder reading capability...")

	// Set a timeout for the test
	testCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Try to read some data from the encoder
	done := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("encoder test panicked: %v", r)
			}
		}()

		// Try to read a few frames using ReadFrame method
		frameCount := 0
		maxFrames := 5 // Just test reading a few frames

		for frameCount < maxFrames {
			select {
			case <-testCtx.Done():
				done <- fmt.Errorf("encoder test timeout after reading %d frames", frameCount)
				return
			default:
				// Try to read from encoder using ReadFrame
				frame, err := encoder.ReadFrame()
				if err != nil {
					done <- fmt.Errorf("encoder ReadFrame failed at frame %d: %w", frameCount+1, err)
					return
				}
				if len(frame) == 0 {
					done <- fmt.Errorf("encoder returned empty frame at frame %d", frameCount)
					return
				}
				frameCount++
				utils.LogDebug("Successfully read frame %d, size: %d bytes", frameCount, len(frame))

				// Small delay between reads
				time.Sleep(100 * time.Millisecond)
			}
		}

		utils.LogDebug("Successfully read %d frames from encoder", frameCount)
		done <- nil
	}()

	select {
	case err := <-done:
		return err
	case <-testCtx.Done():
		return fmt.Errorf("encoder test timed out")
	}
}

// streamToVoice streams encoded audio to Discord voice connection with enhanced error recovery
func (p *DCAAudioPlayer) streamToVoice(encoder *dca.EncodeSession) error {
	utils.LogDebug("Starting enhanced stream to voice with error recovery")

	// Try the robust streaming method first
	if err := p.streamToVoiceRobust(encoder); err != nil {
		utils.LogWarn("Robust streaming failed, attempting fallback: %v", err)
		// Fallback to basic streaming if robust method fails
		return p.streamToVoiceBasic(encoder)
	}

	return nil
}

// streamToVoiceRobust implements enhanced streaming with circuit breaker patterns and error recovery
func (p *DCAAudioPlayer) streamToVoiceRobust(encoder *dca.EncodeSession) error {
	utils.LogDebug("Starting robust streaming with error recovery patterns")

	// Comprehensive voice connection validation with retry
	if err := p.validateVoiceConnectionWithRetry(3); err != nil {
		utils.LogError("Voice connection validation failed after retries: %v", err)
		return err
	}

	// Start speaking with retry logic
	speakingRetries := 3
	for attempt := 1; attempt <= speakingRetries; attempt++ {
		if err := p.voiceConn.Speaking(true); err != nil {
			utils.LogWarn("Failed to start speaking (attempt %d/%d): %v", attempt, speakingRetries, err)
			if attempt < speakingRetries {
				time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
				continue
			}
			return fmt.Errorf("failed to start speaking after %d attempts: %w", speakingRetries, err)
		}
		utils.LogDebug("Speaking started successfully on attempt %d", attempt)
		break
	}

	// Ensure we stop speaking when done
	defer func() {
		if p.voiceConn != nil {
			if err := p.voiceConn.Speaking(false); err != nil {
				utils.LogWarn("Failed to stop speaking: %v", err)
			}
		}
	}()

	// Enhanced streaming with frame-by-frame processing and error recovery
	return p.streamFrameByFrame(encoder)
}

// streamFrameByFrame implements frame-by-frame streaming with comprehensive error handling
func (p *DCAAudioPlayer) streamFrameByFrame(encoder *dca.EncodeSession) error {
	utils.LogDebug("Starting frame-by-frame streaming with error recovery")

	frameCount := 0
	consecutiveErrors := 0
	maxConsecutiveErrors := 5

	for {
		// Check for stop/pause signals
		select {
		case <-p.stopChan:
			utils.LogDebug("Stop signal received during frame streaming")
			return nil
		default:
		}

		// Handle pause state
		if atomic.LoadInt32(&p.paused) == 1 {
			// Wait for resume or stop while paused
			for atomic.LoadInt32(&p.paused) == 1 {
				select {
				case <-p.stopChan:
					return nil
				case <-time.After(50 * time.Millisecond):
					continue
				}
			}
		}

		// Validate encoder and voice connection health periodically
		if frameCount%100 == 0 { // Check every 100 frames (~2 seconds)
			if !encoder.Running() {
				utils.LogDebug("Encoder stopped running after %d frames", frameCount)
				return nil // Normal completion
			}

			if err := encoder.Error(); err != nil {
				utils.LogError("Encoder error detected after %d frames: %v", frameCount, err)
				return fmt.Errorf("encoder error: %w", err)
			}

			if err := p.validateVoiceConnection(); err != nil {
				utils.LogError("Voice connection became invalid after %d frames: %v", frameCount, err)
				return fmt.Errorf("voice connection lost: %w", err)
			}
		}

		// Read frame from encoder
		frame, err := encoder.ReadFrame()
		if err != nil {
			if err == io.EOF {
				utils.LogDebug("Reached end of stream after %d frames", frameCount)
				return nil // Normal completion
			}

			consecutiveErrors++
			utils.LogWarn("Frame read error %d (consecutive: %d): %v", frameCount, consecutiveErrors, err)

			if consecutiveErrors >= maxConsecutiveErrors {
				utils.LogError("Too many consecutive errors (%d), aborting stream", consecutiveErrors)
				return fmt.Errorf("too many consecutive frame read errors: %w", err)
			}

			// Exponential backoff for error recovery
			backoff := time.Duration(consecutiveErrors*consecutiveErrors) * 100 * time.Millisecond
			time.Sleep(backoff)
			continue
		}

		// Reset error counter on successful read
		if consecutiveErrors > 0 {
			utils.LogDebug("Frame read recovered after %d consecutive errors", consecutiveErrors)
			consecutiveErrors = 0
		}

		// Send frame to Discord voice connection
		if err := p.sendFrameWithRetry(frame, 3); err != nil {
			utils.LogError("Failed to send frame %d: %v", frameCount, err)
			return fmt.Errorf("frame send failed: %w", err)
		}

		frameCount++

		// Log progress periodically
		if frameCount%500 == 0 { // Every 500 frames (~10 seconds)
			utils.LogDebug("Streaming progress: %d frames sent", frameCount)
		}
	}
}

// sendFrameWithRetry sends a frame to Discord voice connection with retry logic
func (p *DCAAudioPlayer) sendFrameWithRetry(frame []byte, maxRetries int) error {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case p.voiceConn.OpusSend <- frame:
			return nil // Success
		case <-time.After(time.Second):
			lastErr = fmt.Errorf("frame send timeout on attempt %d", attempt)
			utils.LogWarn("Frame send timeout (attempt %d/%d)", attempt, maxRetries)

			// Check if voice connection is still valid
			if err := p.validateVoiceConnection(); err != nil {
				return fmt.Errorf("voice connection lost during frame send: %w", err)
			}

			if attempt < maxRetries {
				// Wait before retry
				time.Sleep(time.Duration(attempt) * 100 * time.Millisecond)
			}
		}
	}

	return fmt.Errorf("failed to send frame after %d attempts: %w", maxRetries, lastErr)
}

// streamToVoiceBasic implements basic streaming as fallback
func (p *DCAAudioPlayer) streamToVoiceBasic(encoder *dca.EncodeSession) error {
	utils.LogDebug("Using basic streaming as fallback")

	// Comprehensive voice connection validation
	if err := p.validateVoiceConnection(); err != nil {
		utils.LogError("Voice connection validation failed: %v", err)
		return err
	}

	// Start speaking
	if err := p.voiceConn.Speaking(true); err != nil {
		utils.LogError("Failed to start speaking: %v", err)
		return fmt.Errorf("failed to start speaking: %w", err)
	}

	// Ensure we stop speaking when done
	defer func() {
		if p.voiceConn != nil {
			if err := p.voiceConn.Speaking(false); err != nil {
				utils.LogWarn("Failed to stop speaking: %v", err)
			}
		}
	}()

	// Pre-streaming encoder validation
	if err := p.validateEncoderForStreaming(encoder); err != nil {
		utils.LogError("Encoder validation failed: %v", err)
		return fmt.Errorf("encoder validation failed: %w", err)
	}

	// Create streaming session with proper error handling
	utils.LogDebug("Creating DCA streaming session")
	done := make(chan error, 1)
	streamSession := dca.NewStream(encoder, p.voiceConn, done)
	p.streamSession = streamSession

	// Monitor the streaming with pause/resume support and error recovery
	for {
		select {
		case err := <-done:
			if err != nil {
				// Enhanced error analysis and recovery
				errorMsg := err.Error()
				utils.LogDebug("Analyzing streaming error: %s", errorMsg)
				switch {
				case contains(errorMsg, "connection reset") || contains(errorMsg, "broken pipe"):
					return fmt.Errorf("network connection interrupted during streaming: %w", err)
				case contains(errorMsg, "no route to host") || contains(errorMsg, "network unreachable"):
					return fmt.Errorf("network connectivity lost during streaming: %w", err)
				case contains(errorMsg, "context deadline exceeded") || contains(errorMsg, "timeout"):
					return fmt.Errorf("streaming timeout - connection too slow: %w", err)
				case contains(errorMsg, "EOF"):
					utils.LogDebug("EOF detected - attempting stream recovery...")
					p.debugEncoderState(encoder)

					// Try to recover from EOF
					if recoveryErr := p.attemptStreamRecovery(); recoveryErr != nil {
						utils.LogWarn("Stream recovery failed: %v", recoveryErr)
						return fmt.Errorf("stream ended unexpectedly and recovery failed - source may be unavailable: %w", err)
					}

					utils.LogInfo("Stream recovery succeeded, treating EOF as completion")
					return nil
				default:
					return fmt.Errorf("streaming error: %w", err)
				}
			}
			utils.LogInfo("Stream completed successfully")
			return nil

		case <-p.stopChan:
			utils.LogDebug("Stop signal received in streaming loop")
			return nil

		default:
			// Handle pause/resume
			if atomic.LoadInt32(&p.paused) == 1 {
				streamSession.SetPaused(true)
				for atomic.LoadInt32(&p.paused) == 1 {
					select {
					case <-p.stopChan:
						return nil
					case <-time.After(50 * time.Millisecond):
						continue
					}
				}
				streamSession.SetPaused(false)
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// validateVoiceConnection performs comprehensive voice connection validation with health monitoring
func (p *DCAAudioPlayer) validateVoiceConnection() error {
	return p.validateVoiceConnectionWithRetry(1)
}

// validateVoiceConnectionWithRetry performs voice connection validation with retry logic
func (p *DCAAudioPlayer) validateVoiceConnectionWithRetry(maxRetries int) error {
	utils.LogDebug("Validating voice connection with retry (max %d)...", maxRetries)

	var lastErr error
	for attempt := 1; attempt <= maxRetries; attempt++ {
		utils.LogDebug("Voice connection validation attempt %d/%d", attempt, maxRetries)

		if p.voiceConn == nil {
			lastErr = fmt.Errorf("voice connection is nil")
			utils.LogDebug("Voice connection is nil (attempt %d)", attempt)
			continue
		}

		if !p.voiceConn.Ready {
			lastErr = fmt.Errorf("voice connection is not ready")
			utils.LogDebug("Voice connection is not ready (attempt %d)", attempt)

			// Wait a bit for connection to become ready
			if attempt < maxRetries {
				utils.LogDebug("Waiting for voice connection to become ready...")
				time.Sleep(time.Duration(attempt) * time.Second)
			}
			continue
		}

		// Check if the connection is still active
		if p.voiceConn.OpusSend == nil {
			lastErr = fmt.Errorf("voice connection OpusSend channel is nil")
			utils.LogDebug("OpusSend channel is nil (attempt %d)", attempt)
			continue
		}

		// Enhanced voice connection health checks
		if p.voiceConn.GuildID == "" {
			lastErr = fmt.Errorf("voice connection has empty guild ID")
			utils.LogDebug("Voice connection has empty guild ID (attempt %d)", attempt)
			continue
		}

		if p.voiceConn.ChannelID == "" {
			lastErr = fmt.Errorf("voice connection has empty channel ID")
			utils.LogDebug("Voice connection has empty channel ID (attempt %d)", attempt)
			continue
		}

		utils.LogDebug("Voice connection state - GuildID: %s, ChannelID: %s, Ready: %v",
			p.voiceConn.GuildID, p.voiceConn.ChannelID, p.voiceConn.Ready)

		utils.LogDebug("Voice connection validation successful on attempt %d", attempt)
		return nil
	}

	return fmt.Errorf("voice connection validation failed after %d attempts: %w", maxRetries, lastErr)
}

// attemptStreamRecovery attempts to recover from streaming errors
func (p *DCAAudioPlayer) attemptStreamRecovery() error {
	utils.LogDebug("Attempting stream recovery...")

	// Validate voice connection health
	if err := p.validateVoiceConnectionWithRetry(2); err != nil {
		return fmt.Errorf("voice connection health check failed during recovery: %w", err)
	}

	// Check if we can still send to the voice connection
	select {
	case p.voiceConn.OpusSend <- []byte{}:
		utils.LogDebug("Voice connection OpusSend channel is responsive")
	default:
		return fmt.Errorf("voice connection OpusSend channel is blocked or closed")
	}

	utils.LogDebug("Stream recovery successful")
	return nil
}

// validateEncoderForStreaming validates encoder state before streaming
func (p *DCAAudioPlayer) validateEncoderForStreaming(encoder *dca.EncodeSession) error {
	utils.LogDebug("Validating encoder for streaming...")

	if encoder == nil {
		return fmt.Errorf("encoder is nil")
	}

	// Test if we can read from the encoder (using ReadFrame method)
	utils.LogDebug("Testing encoder readability...")
	frame, err := encoder.ReadFrame()
	if err != nil {
		utils.LogDebug("Could not read test frame from encoder: %v", err)
		// Don't fail validation on read errors as the encoder might not have data ready yet
	} else {
		utils.LogDebug("Successfully read test frame from encoder, size: %d bytes", len(frame))
	}

	utils.LogDebug("Encoder validation for streaming successful")
	return nil
}

// debugEncoderState provides detailed debugging information about encoder state
func (p *DCAAudioPlayer) debugEncoderState(encoder *dca.EncodeSession) {
	utils.LogDebug("=== ENCODER STATE DEBUGGING ===")

	if encoder == nil {
		utils.LogDebug("Encoder is nil")
		return
	}

	// Check encoder state
	utils.LogDebug("Checking encoder state...")

	// Check if encoder is still running
	if encoder.Running() {
		utils.LogDebug("Encoder is still running")
	} else {
		utils.LogDebug("Encoder has stopped running")
	}

	// Check for encoder errors
	if err := encoder.Error(); err != nil {
		utils.LogDebug("Encoder has error: %v", err)
	} else {
		utils.LogDebug("Encoder has no errors")
	}

	// Try to read a frame to test encoder state
	frame, err := encoder.ReadFrame()
	if err != nil {
		utils.LogDebug("Cannot read frame from encoder: %v", err)
	} else {
		utils.LogDebug("Successfully read frame, size: %d bytes", len(frame))
	}

	// Add more encoder state debugging as needed
	utils.LogDebug("=== END ENCODER STATE DEBUGGING ===")
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

// tryAlternativeFormat attempts to get an alternative audio format when encoding fails
func (p *DCAAudioPlayer) tryAlternativeFormat(source types.AudioSource, originalErr error) *types.AudioSource {
	utils.LogInfo("Attempting to get alternative format due to encoding failure: %v", originalErr)

	// Only try alternatives for YouTube provider
	if source.Provider != "youtube-ytdlp" {
		utils.LogDebug("Provider %s doesn't support alternative formats", source.Provider)
		return nil
	}

	if source.Metadata == nil {
		utils.LogDebug("No metadata available for alternative format")
		return nil
	}

	// Try to get the selected format to determine codec to avoid
	_, ok := source.Metadata["selectedFormat"]
	if !ok {
		utils.LogDebug("No selected format info in metadata")
		return nil
	}

	// Import the youtube provider to use GetAlternativeFormat
	// Note: This is a bit hacky, in a real implementation we'd pass the provider through
	// For now, let's just try different approaches based on error type

	// Determine codec to avoid based on the original error and current format
	avoidCodec := "opus" // Default to avoiding opus since that's the most common issue

	errorMsg := originalErr.Error()
	if strings.Contains(errorMsg, "opus") || strings.Contains(errorMsg, "webm") {
		avoidCodec = "opus"
	} else if strings.Contains(errorMsg, "mp4") || strings.Contains(errorMsg, "m4a") {
		avoidCodec = "mp4a"
	}

	utils.LogDebug("Will avoid codec: %s for alternative format", avoidCodec)

	// This is a simplified approach - in a full implementation, we'd need access to the provider
	// For now, we'll return nil and log that alternative formats aren't available in this context
	utils.LogDebug("Alternative format system needs provider integration - skipping for now")

	return nil
}

// Ensure DCAAudioPlayer implements the AudioPlayer interface
var _ types.AudioPlayer = (*DCAAudioPlayer)(nil)
