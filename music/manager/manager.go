package manager

import (
	"context"
	"fmt"
	"sync"
	"time"

	"pxnx-discord-bot/music/player"
	"pxnx-discord-bot/music/queue"
	"pxnx-discord-bot/music/types"
)

// Manager implements the MusicManager interface
type Manager struct {
	session   types.SessionInterface
	players   map[string]types.AudioPlayer
	queues    map[string]types.Queue
	providers map[string]types.AudioProvider
	mu        sync.RWMutex

	// Auto-disconnect timers
	aloneTimers map[string]*time.Timer
	timerMu     sync.RWMutex

	// Configuration
	defaultVolume int
	maxQueueSize  int
	aloneTimeout  time.Duration
}

// NewManager creates a new music manager
func NewManager(session types.SessionInterface) *Manager {
	return &Manager{
		session:       session,
		players:       make(map[string]types.AudioPlayer),
		queues:        make(map[string]types.Queue),
		providers:     make(map[string]types.AudioProvider),
		aloneTimers:   make(map[string]*time.Timer),
		defaultVolume: 50,
		maxQueueSize:  100,
		aloneTimeout:  15 * time.Second, // 15 seconds timeout
	}
}

// RegisterProvider registers an audio provider
func (m *Manager) RegisterProvider(provider types.AudioProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[provider.GetProviderName()] = provider
}

// GetProviders returns a slice of all registered providers
func (m *Manager) GetProviders() []types.AudioProvider {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]types.AudioProvider, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, provider)
	}
	return providers
}

// JoinChannel joins a voice channel
func (m *Manager) JoinChannel(ctx context.Context, guildID, channelID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if already connected to the same channel
	if vc := m.session.GetVoiceConnection(guildID); vc != nil {
		if vc.ChannelID == channelID {
			return nil // Already connected to the same channel
		}
		// Disconnect from current channel first
		_ = vc.Disconnect() // Ignore error, continue with connection attempt
		// Give a moment for cleanup
		time.Sleep(500 * time.Millisecond)
	}

	// Connect to the new channel with simplified approach
	vc, err := m.session.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return &types.VoiceChannelError{
			Type:    "connection_failed",
			Message: fmt.Sprintf("Failed to connect to voice channel: %v", err),
			GuildID: guildID,
		}
	}

	// Simple ready check with shorter timeout
	maxWait := 5 * time.Second
	start := time.Now()

	for time.Since(start) < maxWait {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			_ = vc.Disconnect() // Try to disconnect cleanly
			return ctx.Err()
		default:
		}

		// Check if connection is ready
		if vc.Ready {
			// Start monitoring for alone status after a short delay
			// to allow Discord voice state to propagate properly
			go func() {
				time.Sleep(2 * time.Second)
				m.checkAndStartAloneTimer(guildID)
			}()
			return nil
		}

		// Short sleep to avoid busy waiting
		time.Sleep(200 * time.Millisecond)
	}

	// Connection timed out
	_ = vc.Disconnect() // Try to disconnect cleanly
	return &types.VoiceChannelError{
		Type:    "connection_timeout",
		Message: "Voice connection took too long to establish",
		GuildID: guildID,
	}
}

// LeaveChannel leaves the voice channel
func (m *Manager) LeaveChannel(ctx context.Context, guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	vc := m.session.GetVoiceConnection(guildID)
	if vc == nil {
		return nil // Not connected
	}

	// Stop any ongoing playback
	if audioPlayer, exists := m.players[guildID]; exists {
		if err := audioPlayer.Stop(); err != nil {
			// Log error but continue with cleanup
		}
		delete(m.players, guildID)
	}

	// Clear the queue
	if guildQueue, exists := m.queues[guildID]; exists {
		guildQueue.Clear()
		delete(m.queues, guildID)
	}

	// Stop alone timer if running
	m.timerMu.Lock()
	if timer, exists := m.aloneTimers[guildID]; exists {
		timer.Stop()
		delete(m.aloneTimers, guildID)
	}
	m.timerMu.Unlock()

	// Disconnect from voice channel
	return vc.Disconnect()
}

// IsConnected checks if connected to a voice channel
func (m *Manager) IsConnected(guildID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	vc := m.session.GetVoiceConnection(guildID)
	return vc != nil && vc.Ready
}

// Play starts playing an audio source
func (m *Manager) Play(ctx context.Context, guildID string, audioSource types.AudioSource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.IsConnected(guildID) {
		return &types.VoiceChannelError{
			Type:    "not_connected",
			Message: "Bot is not connected to a voice channel",
			GuildID: guildID,
		}
	}

	// Get or create player for this guild
	audioPlayer, exists := m.players[guildID]
	if !exists {
		// Get voice connection for this guild
		vc := m.session.GetVoiceConnection(guildID)
		if vc == nil {
			return &types.VoiceChannelError{
				Type:    "no_voice_connection",
				Message: "No voice connection found for guild",
				GuildID: guildID,
			}
		}

		// Create a new DCA audio player
		audioPlayer = player.NewDCAAudioPlayer(guildID, vc)
		m.players[guildID] = audioPlayer
	}

	// If something is already playing, add to queue
	if audioPlayer.IsPlaying() {
		return m.AddToQueue(ctx, guildID, audioSource)
	}

	// Start playing
	return audioPlayer.Play(ctx, audioSource)
}

// Pause pauses playback
func (m *Manager) Pause(ctx context.Context, guildID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	audioPlayer, exists := m.players[guildID]
	if !exists {
		return &types.MusicError{
			Type:    "no_player",
			Message: "No active player for this guild",
		}
	}

	return audioPlayer.Pause()
}

// Resume resumes playback
func (m *Manager) Resume(ctx context.Context, guildID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	audioPlayer, exists := m.players[guildID]
	if !exists {
		return &types.MusicError{
			Type:    "no_player",
			Message: "No active player for this guild",
		}
	}

	return audioPlayer.Resume()
}

// Stop stops playback
func (m *Manager) Stop(ctx context.Context, guildID string) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	audioPlayer, exists := m.players[guildID]
	if !exists {
		return nil // Nothing to stop
	}

	return audioPlayer.Stop()
}

// Skip skips the current song
func (m *Manager) Skip(ctx context.Context, guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	audioPlayer, exists := m.players[guildID]
	if !exists {
		return &types.MusicError{
			Type:    "no_player",
			Message: "No active player for this guild",
		}
	}

	guildQueue, queueExists := m.queues[guildID]
	if !queueExists {
		return audioPlayer.Stop()
	}

	// Stop current playback
	if err := audioPlayer.Stop(); err != nil {
		return err
	}

	// Play next song from queue
	nextSource, hasNext := guildQueue.Next()
	if !hasNext {
		return nil // Queue is empty
	}

	return audioPlayer.Play(ctx, *nextSource)
}

// AddToQueue adds a song to the queue
func (m *Manager) AddToQueue(ctx context.Context, guildID string, audioSource types.AudioSource) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get or create queue for this guild
	guildQueue, exists := m.queues[guildID]
	if !exists {
		// Create a new queue
		guildQueue = queue.NewQueue()
		m.queues[guildID] = guildQueue
	}

	if guildQueue.Size() >= m.maxQueueSize {
		return &types.MusicError{
			Type:    "queue_full",
			Message: fmt.Sprintf("Queue is full (max %d songs)", m.maxQueueSize),
		}
	}

	guildQueue.Add(audioSource)
	return nil
}

// GetQueue returns the current queue
func (m *Manager) GetQueue(ctx context.Context, guildID string) ([]types.AudioSource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	queue, exists := m.queues[guildID]
	if !exists {
		return []types.AudioSource{}, nil
	}

	return queue.GetAll(), nil
}

// RemoveFromQueue removes a song from the queue
func (m *Manager) RemoveFromQueue(ctx context.Context, guildID string, position int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.queues[guildID]
	if !exists {
		return &types.MusicError{
			Type:    "no_queue",
			Message: "No queue found for this guild",
		}
	}

	return queue.Remove(position)
}

// ClearQueue clears the queue
func (m *Manager) ClearQueue(ctx context.Context, guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.queues[guildID]
	if !exists {
		return nil // Nothing to clear
	}

	queue.Clear()
	return nil
}

// ShuffleQueue shuffles the queue
func (m *Manager) ShuffleQueue(ctx context.Context, guildID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	queue, exists := m.queues[guildID]
	if !exists {
		return &types.MusicError{
			Type:    "no_queue",
			Message: "No queue found for this guild",
		}
	}

	queue.Shuffle()
	return nil
}

// SetVolume sets the playback volume
func (m *Manager) SetVolume(ctx context.Context, guildID string, volume int) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if volume < 0 || volume > 100 {
		return &types.MusicError{
			Type:    "invalid_volume",
			Message: "Volume must be between 0 and 100",
		}
	}

	audioPlayer, exists := m.players[guildID]
	if !exists {
		return &types.MusicError{
			Type:    "no_player",
			Message: "No active player for this guild",
		}
	}

	return audioPlayer.SetVolume(volume)
}

// GetVolume returns the current volume
func (m *Manager) GetVolume(ctx context.Context, guildID string) (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	audioPlayer, exists := m.players[guildID]
	if !exists {
		return m.defaultVolume, nil
	}

	return audioPlayer.GetVolume(), nil
}

// GetNowPlaying returns the currently playing song
func (m *Manager) GetNowPlaying(ctx context.Context, guildID string) (*types.AudioSource, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	audioPlayer, exists := m.players[guildID]
	if !exists {
		return nil, nil
	}

	return audioPlayer.GetCurrentSource(), nil
}

// GetPlayerStatus returns the current player status
func (m *Manager) GetPlayerStatus(ctx context.Context, guildID string) (types.PlayerStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	audioPlayer, exists := m.players[guildID]
	if !exists {
		return types.StatusIdle, nil
	}

	if audioPlayer.IsPlaying() {
		return types.StatusPlaying, nil
	} else if audioPlayer.IsPaused() {
		return types.StatusPaused, nil
	}

	return types.StatusStopped, nil
}

// Cleanup cleans up resources
func (m *Manager) Cleanup(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error

	// Stop all players and disconnect from voice channels
	for guildID := range m.players {
		if err := m.LeaveChannel(ctx, guildID); err != nil {
			errors = append(errors, err)
		}
	}

	// Clear all data
	m.players = make(map[string]types.AudioPlayer)
	m.queues = make(map[string]types.Queue)

	if len(errors) > 0 {
		return &types.MusicError{
			Type:    "cleanup_failed",
			Message: fmt.Sprintf("Failed to cleanup %d resources", len(errors)),
			Err:     errors[0],
		}
	}

	return nil
}

// checkAndStartAloneTimer checks if bot is alone and starts/stops timer accordingly
func (m *Manager) checkAndStartAloneTimer(guildID string) {
	m.timerMu.Lock()
	defer m.timerMu.Unlock()

	// Check if bot is still connected (might have disconnected between calls)
	if !m.IsConnected(guildID) {
		// Clean up any existing timer for disconnected guild
		if timer, exists := m.aloneTimers[guildID]; exists {
			timer.Stop()
			delete(m.aloneTimers, guildID)
		}
		return
	}

	// Check if bot is alone in voice channel
	alone := m.isBotAloneInVoice(guildID)

	if alone {
		// Start timer if not already running
		if _, exists := m.aloneTimers[guildID]; !exists {
			m.aloneTimers[guildID] = time.AfterFunc(m.aloneTimeout, func() {
				// Double-check if still alone and connected before auto-disconnect
				// Use a separate lock to avoid deadlock
				var shouldDisconnect bool
				func() {
					m.timerMu.Lock()
					defer m.timerMu.Unlock()
					_, timerExists := m.aloneTimers[guildID]
					shouldDisconnect = timerExists
				}()

				// Final check: ensure we're still connected and alone
				if shouldDisconnect && m.IsConnected(guildID) {
					// Do a final check with a small delay to ensure accuracy
					time.Sleep(500 * time.Millisecond)
					if m.isBotAloneInVoice(guildID) {
						// Auto-disconnect after timeout
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						if err := m.LeaveChannel(ctx, guildID); err != nil {
							// Auto-disconnect failed, but timer already fired
						}
					} else {
						// Someone joined before disconnect, restart timer check
						m.checkAndStartAloneTimer(guildID)
					}
				}
			})
		}
	} else {
		// Stop timer if running - users are present
		if timer, exists := m.aloneTimers[guildID]; exists {
			timer.Stop()
			delete(m.aloneTimers, guildID)
		}
	}
}

// isBotAloneInVoice checks if the bot is alone in the voice channel
func (m *Manager) isBotAloneInVoice(guildID string) bool {
	vc := m.session.GetVoiceConnection(guildID)
	if vc == nil || !vc.Ready {
		return false // Not connected
	}

	// Get session state for more reliable voice state information
	sessionWrapper, ok := m.session.(*SessionWrapper)
	if !ok {
		// Fallback to API call if not using SessionWrapper
		return m.isBotAloneInVoiceAPI(guildID, vc.ChannelID)
	}

	state := sessionWrapper.session.State
	if state == nil {
		// Fallback to API call if state not available
		return m.isBotAloneInVoiceAPI(guildID, vc.ChannelID)
	}

	// Try to get guild from state first (more reliable)
	state.RLock()
	guild, err := state.Guild(guildID)
	state.RUnlock()

	if err != nil {
		// Fallback to API call
		return m.isBotAloneInVoiceAPI(guildID, vc.ChannelID)
	}

	// Count non-bot users in the same voice channel
	userCount := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == vc.ChannelID && vs.UserID != sessionWrapper.session.State.User.ID {
			// Get user info from state or API
			user, err := sessionWrapper.session.User(vs.UserID)
			if err != nil {
				// If we can't get user info, assume it's a user (safer)
				userCount++
				continue
			}

			// Only count non-bot users
			if !user.Bot {
				userCount++
			}
		}
	}

	return userCount == 0
}

// isBotAloneInVoiceAPI is a fallback method using API calls
func (m *Manager) isBotAloneInVoiceAPI(guildID, channelID string) bool {
	// Try to get voice states from guild via API
	guild, err := m.session.Guild(guildID)
	if err != nil {
		return false // Can't determine, assume not alone (safer)
	}

	// Count non-bot users in the same voice channel
	userCount := 0
	for _, vs := range guild.VoiceStates {
		if vs.ChannelID == channelID {
			// Try to get user information
			if vs.Member != nil && vs.Member.User != nil {
				// Use member information if available
				if !vs.Member.User.Bot {
					userCount++
				}
			} else {
				// If we can't determine bot status, assume it's a user (safer)
				userCount++
			}
		}
	}

	return userCount == 0
}

// OnVoiceStateUpdate should be called when voice states change to update alone timers
func (m *Manager) OnVoiceStateUpdate(guildID string) {
	// Only check if we're connected to this guild
	if m.IsConnected(guildID) {
		// Add a small delay to allow voice state to propagate
		go func() {
			time.Sleep(1 * time.Second)
			m.checkAndStartAloneTimer(guildID)
		}()
	}
}
