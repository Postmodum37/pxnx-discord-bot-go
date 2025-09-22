package music

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
)

// SimplePlayer provides a simplified, reliable Discord music player
// that replaces the complex DCA-based implementation with direct FFmpeg streaming
type SimplePlayer struct {
	session     *discordgo.Session
	connections map[string]*VoicePlayer
	mu          sync.RWMutex
}

// VoicePlayer handles audio playback for a single Discord server
type VoicePlayer struct {
	guildID    string
	conn       *discordgo.VoiceConnection
	queue      []AudioTrack
	current    *AudioTrack
	playing    bool
	stopChan   chan struct{}
	skipChan   chan struct{}
	mu         sync.RWMutex
	ffmpegCmd  *exec.Cmd
}

// AudioTrack represents a playable audio track
type AudioTrack struct {
	Title    string `json:"title"`
	URL      string `json:"url"`
	Duration string `json:"duration"`
	Uploader string `json:"uploader"`
}

// NewSimplePlayer creates a new simplified music player
func NewSimplePlayer(session *discordgo.Session) *SimplePlayer {
	return &SimplePlayer{
		session:     session,
		connections: make(map[string]*VoicePlayer),
	}
}

// JoinChannel connects to a voice channel
func (sp *SimplePlayer) JoinChannel(guildID, channelID string) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	// Check if already connected
	if player, exists := sp.connections[guildID]; exists {
		if player.conn != nil && player.conn.ChannelID == channelID {
			return nil // Already connected to the same channel
		}
		// Disconnect from current channel
		if player.conn != nil {
			player.conn.Disconnect()
		}
	}

	// Connect to voice channel
	conn, err := sp.session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return fmt.Errorf("failed to join voice channel: %w", err)
	}

	// Wait for connection to be ready
	for i := 0; i < 50; i++ { // 5 second timeout
		if conn.Ready {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !conn.Ready {
		conn.Disconnect()
		return fmt.Errorf("voice connection timeout")
	}

	// Create voice player
	player := &VoicePlayer{
		guildID:  guildID,
		conn:     conn,
		queue:    make([]AudioTrack, 0),
		stopChan: make(chan struct{}),
		skipChan: make(chan struct{}),
	}

	sp.connections[guildID] = player
	return nil
}

// LeaveChannel disconnects from voice channel
func (sp *SimplePlayer) LeaveChannel(guildID string) error {
	sp.mu.Lock()
	defer sp.mu.Unlock()

	player, exists := sp.connections[guildID]
	if !exists {
		return nil
	}

	// Stop current playback
	player.Stop()

	// Disconnect voice connection
	if player.conn != nil {
		player.conn.Disconnect()
	}

	// Remove from connections
	delete(sp.connections, guildID)
	return nil
}

// Play adds a track to the queue and starts playback if not already playing
func (sp *SimplePlayer) Play(guildID string, query string) (*AudioTrack, error) {
	sp.mu.RLock()
	player, exists := sp.connections[guildID]
	sp.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("not connected to voice channel")
	}

	// Extract track information using yt-dlp
	track, err := sp.extractTrackInfo(query)
	if err != nil {
		return nil, fmt.Errorf("failed to extract track info: %w", err)
	}

	player.mu.Lock()
	defer player.mu.Unlock()

	// Add to queue
	player.queue = append(player.queue, *track)

	// Start playback if not already playing
	if !player.playing {
		go player.playNext()
	}

	return track, nil
}

// extractTrackInfo uses yt-dlp to extract track information and stream URL
func (sp *SimplePlayer) extractTrackInfo(query string) (*AudioTrack, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use yt-dlp to extract information
	cmd := exec.CommandContext(ctx, "yt-dlp",
		"--default-search", "ytsearch",
		"--format", "bestaudio[ext=webm]/bestaudio",
		"--get-title",
		"--get-url",
		"--get-duration",
		"--get-description",
		query,
	)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("yt-dlp extraction failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("invalid yt-dlp output")
	}

	track := &AudioTrack{
		Title: lines[0],
		URL:   lines[1],
	}

	if len(lines) > 2 {
		track.Duration = lines[2]
	}

	return track, nil
}

// playNext plays the next track in the queue
func (vp *VoicePlayer) playNext() {
	vp.mu.Lock()
	if len(vp.queue) == 0 {
		vp.playing = false
		vp.mu.Unlock()
		return
	}

	track := vp.queue[0]
	vp.queue = vp.queue[1:]
	vp.current = &track
	vp.playing = true
	vp.mu.Unlock()

	// Play the track
	err := vp.playTrack(track)
	if err != nil {
		fmt.Printf("Failed to play track %s: %v\n", track.Title, err)
	}

	// Continue with next track
	go vp.playNext()
}

// playTrack streams audio using FFmpeg directly to Discord
func (vp *VoicePlayer) playTrack(track AudioTrack) error {
	// Start speaking
	err := vp.conn.Speaking(true)
	if err != nil {
		return fmt.Errorf("failed to start speaking: %w", err)
	}
	defer vp.conn.Speaking(false)

	// Create FFmpeg command for direct streaming
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Enhanced FFmpeg command with better error handling and reconnection
	vp.ffmpegCmd = exec.CommandContext(ctx, "ffmpeg",
		"-reconnect", "1",
		"-reconnect_streamed", "1",
		"-reconnect_delay_max", "2",
		"-i", track.URL,
		"-f", "s16le",
		"-ar", "48000",
		"-ac", "2",
		"-application", "audio",
		"-vn",
		"pipe:1",
	)

	stdout, err := vp.ffmpegCmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	err = vp.ffmpegCmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	// Stream audio to Discord
	go func() {
		defer stdout.Close()

		// Create a buffer for audio data
		buffer := make([]byte, 3840) // 20ms of 48kHz 16-bit stereo

		for {
			select {
			case <-vp.stopChan:
				cancel()
				return
			case <-vp.skipChan:
				cancel()
				return
			default:
				// Read audio data
				n, err := stdout.Read(buffer)
				if err != nil {
					if err != io.EOF {
						fmt.Printf("Error reading audio data: %v\n", err)
					}
					return
				}

				if n > 0 {
					// Send to Discord voice connection
					select {
					case vp.conn.OpusSend <- buffer[:n]:
					case <-time.After(time.Millisecond * 100):
						// Drop frame if channel is full
					}
				}
			}
		}
	}()

	// Wait for FFmpeg to complete or be cancelled
	err = vp.ffmpegCmd.Wait()
	if err != nil && ctx.Err() == nil {
		return fmt.Errorf("ffmpeg process failed: %w", err)
	}

	return nil
}

// Stop stops current playback
func (vp *VoicePlayer) Stop() {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	if vp.playing {
		close(vp.stopChan)
		vp.stopChan = make(chan struct{})
		vp.playing = false
		vp.current = nil
		vp.queue = vp.queue[:0] // Clear queue
	}

	// Kill FFmpeg process if running
	if vp.ffmpegCmd != nil && vp.ffmpegCmd.Process != nil {
		vp.ffmpegCmd.Process.Kill()
	}
}

// Skip skips current track
func (vp *VoicePlayer) Skip() {
	vp.mu.Lock()
	defer vp.mu.Unlock()

	if vp.playing {
		close(vp.skipChan)
		vp.skipChan = make(chan struct{})
	}
}

// GetQueue returns current queue
func (vp *VoicePlayer) GetQueue() []AudioTrack {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	queue := make([]AudioTrack, len(vp.queue))
	copy(queue, vp.queue)
	return queue
}

// GetCurrent returns currently playing track
func (vp *VoicePlayer) GetCurrent() *AudioTrack {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	if vp.current == nil {
		return nil
	}

	current := *vp.current
	return &current
}

// IsPlaying returns whether player is currently playing
func (vp *VoicePlayer) IsPlaying() bool {
	vp.mu.RLock()
	defer vp.mu.RUnlock()

	return vp.playing
}

// GetPlayer returns the voice player for a guild
func (sp *SimplePlayer) GetPlayer(guildID string) (*VoicePlayer, bool) {
	sp.mu.RLock()
	defer sp.mu.RUnlock()

	player, exists := sp.connections[guildID]
	return player, exists
}