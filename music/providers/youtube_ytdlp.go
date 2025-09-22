package providers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"pxnx-discord-bot/music/types"
	"pxnx-discord-bot/services/ytdlp"
)

// YouTubeYTDLPProvider implements the AudioProvider interface using yt-dlp service
type YouTubeYTDLPProvider struct {
	serviceManager *ytdlp.ServiceManager
	client         *ytdlp.Client
	mu             sync.RWMutex
	started        bool
}

// NewYouTubeYTDLPProvider creates a new YouTube provider using yt-dlp service
func NewYouTubeYTDLPProvider() *YouTubeYTDLPProvider {
	config := ytdlp.DefaultServiceConfig()

	// Customize config for better performance
	config.Format = "bestaudio[ext=webm][acodec=opus]/bestaudio[ext=m4a]/bestaudio/best"
	config.AudioFormat = "opus"
	config.AudioQuality = "128K"
	config.MaxWorkers = 2 // Conservative for Discord bot
	config.Timeout = 45 * time.Second // Longer timeout for yt-dlp operations
	config.MaxRetries = 2

	serviceManager := ytdlp.NewServiceManager(config)

	return &YouTubeYTDLPProvider{
		serviceManager: serviceManager,
		client:         serviceManager.GetClient(),
	}
}

// NewYouTubeYTDLPProviderWithConfig creates a new provider with custom configuration
func NewYouTubeYTDLPProviderWithConfig(config *ytdlp.ServiceConfig) *YouTubeYTDLPProvider {
	serviceManager := ytdlp.NewServiceManager(config)

	return &YouTubeYTDLPProvider{
		serviceManager: serviceManager,
		client:         serviceManager.GetClient(),
	}
}

// Start starts the yt-dlp service if not already running
func (yt *YouTubeYTDLPProvider) Start(ctx context.Context) error {
	yt.mu.Lock()
	defer yt.mu.Unlock()

	log.Printf("[YTDLP] Starting yt-dlp service, already started: %v", yt.started)

	if yt.started {
		log.Printf("[YTDLP] Service already marked as started, checking if actually running")
		if !yt.serviceManager.IsRunning() {
			log.Printf("[YTDLP] Service was marked as started but isn't running, restarting")
			yt.started = false
		} else {
			log.Printf("[YTDLP] Service is actually running")
			return nil
		}
	}

	log.Printf("[YTDLP] Starting service manager")
	if err := yt.serviceManager.Start(ctx); err != nil {
		log.Printf("[YTDLP] Failed to start service manager: %v", err)
		return fmt.Errorf("failed to start yt-dlp service: %w", err)
	}

	yt.started = true
	log.Printf("[YTDLP] Service started successfully, status: %s", yt.serviceManager.GetStatus())
	return nil
}

// Stop stops the yt-dlp service
func (yt *YouTubeYTDLPProvider) Stop(ctx context.Context) error {
	yt.mu.Lock()
	defer yt.mu.Unlock()

	if !yt.started {
		return nil
	}

	if err := yt.serviceManager.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop yt-dlp service: %w", err)
	}

	yt.started = false
	return nil
}

// IsServiceRunning checks if the yt-dlp service is running
func (yt *YouTubeYTDLPProvider) IsServiceRunning() bool {
	yt.mu.RLock()
	defer yt.mu.RUnlock()

	return yt.started && yt.serviceManager.IsRunning()
}

// GetProviderName returns the name of this provider
func (yt *YouTubeYTDLPProvider) GetProviderName() string {
	return "youtube-ytdlp"
}

// SupportsURL checks if this provider can handle the given URL
func (yt *YouTubeYTDLPProvider) SupportsURL(url string) bool {
	return yt.isYouTubeURL(url)
}

// isYouTubeURL checks if the URL is a valid YouTube URL
func (yt *YouTubeYTDLPProvider) isYouTubeURL(urlStr string) bool {
	// Supported YouTube URL patterns:
	// https://www.youtube.com/watch?v=VIDEO_ID
	// https://youtu.be/VIDEO_ID
	// https://youtube.com/watch?v=VIDEO_ID
	// https://m.youtube.com/watch?v=VIDEO_ID
	// https://music.youtube.com/watch?v=VIDEO_ID

	patterns := []string{
		`^https?://(www\.)?youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`,
		`^https?://youtu\.be/([a-zA-Z0-9_-]{11})`,
		`^https?://m\.youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`,
		`^https?://music\.youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`,
		`^https?://(www\.)?youtube\.com/shorts/([a-zA-Z0-9_-]{11})`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, urlStr)
		if matched {
			return true
		}
	}

	return false
}

// GetAudioSource gets audio source information for a YouTube URL or search query
func (yt *YouTubeYTDLPProvider) GetAudioSource(ctx context.Context, query string) (*types.AudioSource, error) {
	log.Printf("[YTDLP] GetAudioSource called with query: %s", query)

	if !yt.IsServiceRunning() {
		log.Printf("[YTDLP] Service not running, starting it")
		if err := yt.Start(ctx); err != nil {
			log.Printf("[YTDLP] Failed to start service: %v", err)
			return nil, fmt.Errorf("failed to start yt-dlp service: %w", err)
		}
	} else {
		log.Printf("[YTDLP] Service is already running")
	}

	var videoInfo *ytdlp.VideoInfo
	var err error

	if yt.isYouTubeURL(query) {
		log.Printf("[YTDLP] Query is a YouTube URL, extracting info directly")
		// It's a URL, extract video info directly
		videoInfo, err = yt.client.ExtractInfo(ctx, query)
		if err != nil {
			log.Printf("[YTDLP] Failed to extract video info: %v", err)
			return nil, fmt.Errorf("failed to extract video info: %w", err)
		}
	} else {
		log.Printf("[YTDLP] Query is a search term, performing search")
		// It's a search query, search for videos and take the first result
		searchResults, searchErr := yt.Search(ctx, query, 1)
		if searchErr != nil {
			log.Printf("[YTDLP] Search failed: %v", searchErr)
			return nil, fmt.Errorf("search failed: %w", searchErr)
		}

		log.Printf("[YTDLP] Search returned %d results", len(searchResults))
		if len(searchResults) == 0 {
			return nil, fmt.Errorf("no results found for query: %s", query)
		}

		log.Printf("[YTDLP] Using first search result: %s", searchResults[0].Title)
		return &searchResults[0], nil
	}

	// Convert yt-dlp VideoInfo to AudioSource
	log.Printf("[YTDLP] Converting video info to audio source")
	return yt.videoInfoToAudioSource(videoInfo)
}

// Search searches YouTube for videos matching the query
func (yt *YouTubeYTDLPProvider) Search(ctx context.Context, query string, maxResults int) ([]types.AudioSource, error) {
	log.Printf("[YTDLP] Search called with query: %s, maxResults: %d", query, maxResults)

	if !yt.IsServiceRunning() {
		log.Printf("[YTDLP] Service not running for search, starting it")
		if err := yt.Start(ctx); err != nil {
			log.Printf("[YTDLP] Failed to start service for search: %v", err)
			return nil, fmt.Errorf("failed to start yt-dlp service: %w", err)
		}
	}

	if maxResults <= 0 || maxResults > 50 {
		maxResults = 10 // Default to 10, cap at 50
	}

	log.Printf("[YTDLP] Making search request to client")
	searchResult, err := yt.client.Search(ctx, query, maxResults)
	if err != nil {
		log.Printf("[YTDLP] Client search failed: %v", err)
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	log.Printf("[YTDLP] Search result: %d videos found", len(searchResult.Videos))
	if len(searchResult.Videos) == 0 {
		return []types.AudioSource{}, nil
	}

	// Convert VideoInfo results to AudioSource
	audioSources := make([]types.AudioSource, 0, len(searchResult.Videos))
	for i, video := range searchResult.Videos {
		log.Printf("[YTDLP] Converting video %d: %s", i+1, video.Title)
		audioSource, err := yt.videoInfoToAudioSource(&video)
		if err != nil {
			log.Printf("[YTDLP] Failed to convert video %d: %v", i+1, err)
			// Log error but continue with other results
			continue
		}
		audioSources = append(audioSources, *audioSource)
	}

	log.Printf("[YTDLP] Successfully converted %d audio sources", len(audioSources))
	return audioSources, nil
}

// videoInfoToAudioSource converts yt-dlp VideoInfo to AudioSource
func (yt *YouTubeYTDLPProvider) videoInfoToAudioSource(video *ytdlp.VideoInfo) (*types.AudioSource, error) {
	return yt.videoInfoToAudioSourceWithCodec(video, "opus")
}

// videoInfoToAudioSourceWithCodec converts yt-dlp VideoInfo to AudioSource with specific codec preference
func (yt *YouTubeYTDLPProvider) videoInfoToAudioSourceWithCodec(video *ytdlp.VideoInfo, preferredCodec string) (*types.AudioSource, error) {
	if video == nil {
		return nil, fmt.Errorf("video info is nil")
	}

	log.Printf("[YTDLP] DEBUG: Converting video to audio source: %s (preferred codec: %s)", video.Title, preferredCodec)
	log.Printf("[YTDLP] DEBUG: Video has %d formats available", len(video.Formats))

	// DEBUG: Log all available formats for analysis
	yt.debugVideoFormats(video)

	// Find the best audio format with codec preference
	streamURL, selectedFormat, err := yt.getBestAudioStreamURLWithPreference(video, preferredCodec)
	if err != nil {
		return nil, fmt.Errorf("failed to find audio stream: %w", err)
	}

	log.Printf("[YTDLP] DEBUG: Selected format - Codec: %s, ABR: %f, URL length: %d",
		selectedFormat.ACodec, selectedFormat.ABR, len(streamURL))

	// Format duration
	duration := yt.formatDuration(video.Duration)

	audioSource := &types.AudioSource{
		Title:     video.Title,
		URL:       video.URL,
		Duration:  duration,
		Thumbnail: video.Thumbnail,
		Provider:  "youtube-ytdlp",
		StreamURL: streamURL,
		// Store the selected format info for potential fallback
		Metadata: map[string]interface{}{
			"selectedFormat": selectedFormat,
			"videoInfo":      video,
		},
	}

	// DEBUG: Validate the stream URL before returning
	if err := yt.validateStreamURLBasic(streamURL); err != nil {
		log.Printf("[YTDLP] DEBUG: Stream URL validation failed: %v", err)
		return nil, fmt.Errorf("stream URL validation failed: %w", err)
	}

	log.Printf("[YTDLP] DEBUG: Audio source created successfully")
	return audioSource, nil
}

// GetAlternativeFormat gets an alternative audio format for a source when the primary fails
func (yt *YouTubeYTDLPProvider) GetAlternativeFormat(source *types.AudioSource, avoidCodec string) (*types.AudioSource, error) {
	if source.Metadata == nil {
		return nil, fmt.Errorf("no metadata available for alternative format")
	}

	videoInfo, ok := source.Metadata["videoInfo"].(*ytdlp.VideoInfo)
	if !ok {
		return nil, fmt.Errorf("video info not available in metadata")
	}

	// Try alternative codecs (avoid the one that failed)
	alternativeCodecs := []string{"mp4a", "aac", "m4a"}
	if avoidCodec != "opus" {
		alternativeCodecs = append([]string{"opus"}, alternativeCodecs...)
	}

	for _, codec := range alternativeCodecs {
		if codec == avoidCodec {
			continue
		}

		log.Printf("[YTDLP] Trying alternative codec: %s", codec)
		altSource, err := yt.videoInfoToAudioSourceWithCodec(videoInfo, codec)
		if err != nil {
			log.Printf("[YTDLP] Alternative codec %s failed: %v", codec, err)
			continue
		}

		log.Printf("[YTDLP] Successfully created alternative format with codec: %s", codec)
		return altSource, nil
	}

	return nil, fmt.Errorf("no alternative formats available")
}

// getBestAudioStreamURL finds the best audio stream URL from video formats
func (yt *YouTubeYTDLPProvider) getBestAudioStreamURL(video *ytdlp.VideoInfo) (string, error) {
	url, _, err := yt.getBestAudioStreamURLWithFormat(video)
	return url, err
}

// getBestAudioStreamURLWithFormat finds the best audio stream URL and returns both URL and format info
func (yt *YouTubeYTDLPProvider) getBestAudioStreamURLWithFormat(video *ytdlp.VideoInfo) (string, *ytdlp.FormatInfo, error) {
	return yt.getBestAudioStreamURLWithPreference(video, "opus")
}

// getBestAudioStreamURLWithPreference finds the best audio stream URL with codec preference
func (yt *YouTubeYTDLPProvider) getBestAudioStreamURLWithPreference(video *ytdlp.VideoInfo, preferredCodec string) (string, *ytdlp.FormatInfo, error) {
	if len(video.Formats) == 0 {
		return "", nil, fmt.Errorf("no formats available")
	}

	log.Printf("[YTDLP] DEBUG: Analyzing %d formats for best audio stream (preferred: %s)", len(video.Formats), preferredCodec)

	var bestFormat *ytdlp.FormatInfo
	var fallbackFormat *ytdlp.FormatInfo

	// Define format preference order for fallback
	codecPreferences := []string{preferredCodec, "mp4a", "aac", "m4a"}
	if preferredCodec != "opus" {
		codecPreferences = []string{preferredCodec, "opus", "mp4a", "aac", "m4a"}
	}

	// First pass: look for preferred codec in audio-only formats
	for i := range video.Formats {
		format := &video.Formats[i]

		// Skip video-only formats
		if format.VCodec != "" && format.VCodec != "none" && format.ACodec == "none" {
			continue
		}

		// Prefer audio-only formats
		if format.VCodec == "none" || format.VCodec == "" {
			log.Printf("[YTDLP] DEBUG: Found audio-only format - Codec: %s, ABR: %f", format.ACodec, format.ABR)

			// Check if this matches our preferred codec
			for priority, codec := range codecPreferences {
				if strings.Contains(format.ACodec, codec) {
					log.Printf("[YTDLP] DEBUG: Found %s format (priority %d)", codec, priority)
					if priority == 0 { // Preferred codec
						bestFormat = format
						break
					} else if fallbackFormat == nil || format.ABR > fallbackFormat.ABR {
						fallbackFormat = format
					}
				}
			}

			if bestFormat != nil {
				break
			}
		}
	}

	// If no preferred codec found, use fallback or highest quality
	if bestFormat == nil && fallbackFormat != nil {
		log.Printf("[YTDLP] DEBUG: Using fallback format: %s", fallbackFormat.ACodec)
		bestFormat = fallbackFormat
	}

	// If no audio-only format found, look for combined formats with audio
	if bestFormat == nil {
		log.Printf("[YTDLP] DEBUG: No suitable audio-only format found, looking for combined formats")
		for i := range video.Formats {
			format := &video.Formats[i]

			if format.ACodec != "" && format.ACodec != "none" {
				log.Printf("[YTDLP] DEBUG: Found combined format - ACodec: %s, VCodec: %s, ABR: %f",
					format.ACodec, format.VCodec, format.ABR)
				if bestFormat == nil || format.ABR > bestFormat.ABR {
					bestFormat = format
				}
			}
		}
	}

	if bestFormat == nil {
		return "", nil, fmt.Errorf("no suitable audio format found")
	}

	if bestFormat.URL == "" {
		return "", nil, fmt.Errorf("selected format has no URL")
	}

	// Safely truncate URL for logging
	urlPreview := bestFormat.URL
	if len(urlPreview) > 100 {
		urlPreview = urlPreview[:100] + "..."
	}

	log.Printf("[YTDLP] DEBUG: Selected best format - ID: %s, Codec: %s, ABR: %f, URL: %s",
		bestFormat.FormatID, bestFormat.ACodec, bestFormat.ABR, urlPreview)

	return bestFormat.URL, bestFormat, nil
}

// formatDuration formats duration from seconds to MM:SS or HH:MM:SS format
func (yt *YouTubeYTDLPProvider) formatDuration(duration float64) string {
	if duration <= 0 {
		return "0:00"
	}

	totalSeconds := int(duration)
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// ValidateURL validates and normalizes a YouTube URL
func (yt *YouTubeYTDLPProvider) ValidateURL(urlStr string) (*url.URL, error) {
	if !yt.isYouTubeURL(urlStr) {
		return nil, fmt.Errorf("not a valid YouTube URL: %s", urlStr)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	return parsedURL, nil
}

// GetVideoInfo gets basic video information (wrapper around ExtractInfo)
func (yt *YouTubeYTDLPProvider) GetVideoInfo(ctx context.Context, url string) (*ytdlp.VideoInfo, error) {
	if !yt.IsServiceRunning() {
		if err := yt.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start yt-dlp service: %w", err)
		}
	}

	return yt.client.ExtractInfo(ctx, url)
}

// IsVideoAvailable checks if a video is available and not private/deleted
func (yt *YouTubeYTDLPProvider) IsVideoAvailable(ctx context.Context, url string) (bool, error) {
	video, err := yt.GetVideoInfo(ctx, url)
	if err != nil {
		return false, err
	}

	if video == nil {
		return false, fmt.Errorf("video not found")
	}

	if !video.Available {
		return false, fmt.Errorf("video is not available")
	}

	if video.Title == "" {
		return false, fmt.Errorf("video has no title (likely private or deleted)")
	}

	return true, nil
}

// GetServiceStatus returns the current status of the yt-dlp service
func (yt *YouTubeYTDLPProvider) GetServiceStatus() ytdlp.ServiceStatus {
	return yt.serviceManager.GetStatus()
}

// GetServiceHealth returns the health status of the yt-dlp service
func (yt *YouTubeYTDLPProvider) GetServiceHealth(ctx context.Context) (*ytdlp.HealthStatus, error) {
	if !yt.IsServiceRunning() {
		return nil, fmt.Errorf("service is not running")
	}

	return yt.client.HealthCheck(ctx)
}

// GetServiceErrors returns a channel for receiving service errors
func (yt *YouTubeYTDLPProvider) GetServiceErrors() <-chan error {
	return yt.serviceManager.GetErrors()
}

// ClearCache clears the yt-dlp service cache
func (yt *YouTubeYTDLPProvider) ClearCache(ctx context.Context) error {
	if !yt.IsServiceRunning() {
		return fmt.Errorf("service is not running")
	}

	return yt.client.ClearCache(ctx)
}

// RestartService restarts the yt-dlp service
func (yt *YouTubeYTDLPProvider) RestartService(ctx context.Context) error {
	yt.mu.Lock()
	defer yt.mu.Unlock()

	if err := yt.serviceManager.Restart(ctx); err != nil {
		yt.started = false
		return fmt.Errorf("failed to restart yt-dlp service: %w", err)
	}

	yt.started = true
	return nil
}

// Cleanup cleans up the provider and stops the service
func (yt *YouTubeYTDLPProvider) Cleanup(ctx context.Context) error {
	return yt.Stop(ctx)
}

// debugVideoFormats logs detailed information about all available formats
func (yt *YouTubeYTDLPProvider) debugVideoFormats(video *ytdlp.VideoInfo) {
	log.Printf("[YTDLP] DEBUG: === VIDEO FORMAT ANALYSIS ===")
	log.Printf("[YTDLP] DEBUG: Video: %s", video.Title)
	log.Printf("[YTDLP] DEBUG: Total formats: %d", len(video.Formats))

	audioOnlyCount := 0
	videoOnlyCount := 0
	combinedCount := 0

	for i, format := range video.Formats {
		formatType := "combined"
		if format.VCodec == "none" || format.VCodec == "" {
			formatType = "audio-only"
			audioOnlyCount++
		} else if format.ACodec == "none" || format.ACodec == "" {
			formatType = "video-only"
			videoOnlyCount++
		} else {
			combinedCount++
		}

		log.Printf("[YTDLP] DEBUG: Format %d - ID: %s, Type: %s, ACodec: %s, VCodec: %s, ABR: %f, FileSize: %d",
			i+1, format.FormatID, formatType, format.ACodec, format.VCodec, format.ABR, format.Filesize)
	}

	log.Printf("[YTDLP] DEBUG: Format summary - Audio-only: %d, Video-only: %d, Combined: %d",
		audioOnlyCount, videoOnlyCount, combinedCount)
	log.Printf("[YTDLP] DEBUG: === END FORMAT ANALYSIS ===")
}

// validateStreamURLBasic performs basic validation on a stream URL
func (yt *YouTubeYTDLPProvider) validateStreamURLBasic(streamURL string) error {
	log.Printf("[YTDLP] DEBUG: Validating stream URL (basic check)")

	if streamURL == "" {
		return fmt.Errorf("stream URL is empty")
	}

	// Parse URL to check format
	parsedURL, err := url.Parse(streamURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	if parsedURL.Host == "" {
		return fmt.Errorf("stream URL has no host")
	}

	log.Printf("[YTDLP] DEBUG: Stream URL host: %s", parsedURL.Host)
	log.Printf("[YTDLP] DEBUG: Stream URL scheme: %s", parsedURL.Scheme)

	// Check if it's a typical YouTube video URL
	if strings.Contains(parsedURL.Host, "googlevideo.com") {
		log.Printf("[YTDLP] DEBUG: Detected Google Video CDN URL")
	}

	// Perform a quick HEAD request to check accessibility
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, err := http.NewRequest("HEAD", streamURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; Discord Music Bot)")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[YTDLP] DEBUG: HEAD request failed (may be normal for some streams): %v", err)
		// Don't fail on HEAD request errors - some streams don't support HEAD
		return nil
	}
	defer resp.Body.Close()

	log.Printf("[YTDLP] DEBUG: Stream URL validation response: %s", resp.Status)
	log.Printf("[YTDLP] DEBUG: Content-Type: %s", resp.Header.Get("Content-Type"))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("stream URL returned error status: %s", resp.Status)
	}

	log.Printf("[YTDLP] DEBUG: Stream URL basic validation passed")
	return nil
}

// Ensure YouTubeYTDLPProvider implements the AudioProvider interface
var _ types.AudioProvider = (*YouTubeYTDLPProvider)(nil)