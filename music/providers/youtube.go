package providers

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"

	"pxnx-discord-bot/music/types"
)

// YouTubeProvider implements the AudioProvider interface for YouTube
type YouTubeProvider struct {
	client *youtube.Client
}

// NewYouTubeProvider creates a new YouTube provider
func NewYouTubeProvider() *YouTubeProvider {
	return &YouTubeProvider{
		client: &youtube.Client{},
	}
}

// GetProviderName returns the name of this provider
func (yt *YouTubeProvider) GetProviderName() string {
	return "youtube"
}

// SupportsURL checks if this provider can handle the given URL
func (yt *YouTubeProvider) SupportsURL(url string) bool {
	return yt.isYouTubeURL(url)
}

// isYouTubeURL checks if the URL is a valid YouTube URL
func (yt *YouTubeProvider) isYouTubeURL(urlStr string) bool {
	// Supported YouTube URL patterns:
	// https://www.youtube.com/watch?v=VIDEO_ID
	// https://youtu.be/VIDEO_ID
	// https://youtube.com/watch?v=VIDEO_ID
	// https://m.youtube.com/watch?v=VIDEO_ID

	patterns := []string{
		`^https?://(www\.)?youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`,
		`^https?://youtu\.be/([a-zA-Z0-9_-]{11})`,
		`^https?://m\.youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, urlStr)
		if matched {
			return true
		}
	}

	return false
}

// extractVideoID extracts the video ID from a YouTube URL
func (yt *YouTubeProvider) extractVideoID(urlStr string) (string, error) {
	patterns := []struct {
		regex *regexp.Regexp
		group int // Which capture group contains the video ID
	}{
		{regexp.MustCompile(`^https?://(www\.)?youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`), 2},
		{regexp.MustCompile(`^https?://youtu\.be/([a-zA-Z0-9_-]{11})`), 1},
		{regexp.MustCompile(`^https?://m\.youtube\.com/watch\?.*v=([a-zA-Z0-9_-]{11})`), 1},
	}

	for _, pattern := range patterns {
		matches := pattern.regex.FindStringSubmatch(urlStr)
		if len(matches) > pattern.group {
			return matches[pattern.group], nil
		}
	}

	return "", fmt.Errorf("could not extract video ID from URL: %s", urlStr)
}

// GetAudioSource gets audio source information for a YouTube URL or search query
func (yt *YouTubeProvider) GetAudioSource(ctx context.Context, query string) (*types.AudioSource, error) {
	var video *youtube.Video
	var err error

	if yt.isYouTubeURL(query) {
		// It's a URL, extract video ID and get video info
		videoID, extractErr := yt.extractVideoID(query)
		if extractErr != nil {
			return nil, fmt.Errorf("failed to extract video ID: %w", extractErr)
		}

		video, err = yt.client.GetVideoContext(ctx, videoID)
		if err != nil {
			return nil, fmt.Errorf("failed to get video info for %s: %w", videoID, err)
		}
	} else {
		// It's a search query, search for videos and take the first result
		searchResults, searchErr := yt.Search(ctx, query, 1)
		if searchErr != nil {
			return nil, fmt.Errorf("search failed: %w", searchErr)
		}

		if len(searchResults) == 0 {
			return nil, fmt.Errorf("no results found for query: %s", query)
		}

		return &searchResults[0], nil
	}

	// Convert YouTube video to AudioSource
	return yt.videoToAudioSource(video)
}

// Search searches YouTube for videos matching the query
func (yt *YouTubeProvider) Search(ctx context.Context, query string, maxResults int) ([]types.AudioSource, error) {
	if maxResults <= 0 || maxResults > 50 {
		maxResults = 10 // Default to 10, cap at 50
	}
	_ = maxResults // Used for future implementation

	// Note: The kkdai/youtube library doesn't have built-in search functionality
	// In a real implementation, you would either:
	// 1. Use YouTube Data API v3 for search (requires API key)
	// 2. Use a scraping approach (less reliable)
	// 3. Use a different library that supports search

	// For now, return an error indicating search is not implemented
	return nil, fmt.Errorf("search functionality requires YouTube Data API v3 key - not implemented in this demo")
}

// videoToAudioSource converts a YouTube video to an AudioSource
func (yt *YouTubeProvider) videoToAudioSource(video *youtube.Video) (*types.AudioSource, error) {
	if video == nil {
		return nil, fmt.Errorf("video is nil")
	}

	// Find the best audio format
	audioFormat, err := yt.getBestAudioFormat(video)
	if err != nil {
		return nil, fmt.Errorf("failed to find audio format: %w", err)
	}

	// Format duration
	duration := yt.formatDuration(video.Duration)

	// Get thumbnail URL
	thumbnail := yt.getBestThumbnail(video)

	return &types.AudioSource{
		Title:     video.Title,
		URL:       fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID),
		Duration:  duration,
		Thumbnail: thumbnail,
		Provider:  "youtube",
		StreamURL: audioFormat.URL,
	}, nil
}

// getBestAudioFormat finds the best audio format from available formats
func (yt *YouTubeProvider) getBestAudioFormat(video *youtube.Video) (*youtube.Format, error) {
	var bestFormat *youtube.Format

	// Prefer audio-only formats first
	for _, format := range video.Formats {
		if format.MimeType != "" && strings.Contains(format.MimeType, "audio") {
			// Prefer opus codec for Discord
			if strings.Contains(format.MimeType, "opus") {
				bestFormat = &format
				break
			}
			// Otherwise, prefer higher quality audio
			if bestFormat == nil || format.Bitrate > bestFormat.Bitrate {
				bestFormat = &format
			}
		}
	}

	// If no audio-only format found, look for formats with audio
	if bestFormat == nil {
		for _, format := range video.Formats {
			if format.AudioChannels > 0 {
				if bestFormat == nil || format.Bitrate > bestFormat.Bitrate {
					bestFormat = &format
				}
			}
		}
	}

	if bestFormat == nil {
		return nil, fmt.Errorf("no suitable audio format found")
	}

	return bestFormat, nil
}

// getBestThumbnail gets the best quality thumbnail URL
func (yt *YouTubeProvider) getBestThumbnail(video *youtube.Video) string {
	if len(video.Thumbnails) == 0 {
		return ""
	}

	// Find the highest quality thumbnail
	var bestThumbnail youtube.Thumbnail
	for _, thumbnail := range video.Thumbnails {
		if thumbnail.Width > bestThumbnail.Width {
			bestThumbnail = thumbnail
		}
	}

	return bestThumbnail.URL
}

// formatDuration formats duration from time.Duration to MM:SS or HH:MM:SS format
func (yt *YouTubeProvider) formatDuration(duration time.Duration) string {
	if duration == 0 {
		return "0:00"
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// ValidateURL validates and normalizes a YouTube URL
func (yt *YouTubeProvider) ValidateURL(urlStr string) (*url.URL, error) {
	if !yt.isYouTubeURL(urlStr) {
		return nil, fmt.Errorf("not a valid YouTube URL: %s", urlStr)
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	return parsedURL, nil
}

// GetVideoInfo gets basic video information without formats (faster)
func (yt *YouTubeProvider) GetVideoInfo(ctx context.Context, videoID string) (*youtube.Video, error) {
	return yt.client.GetVideoContext(ctx, videoID)
}

// IsVideoAvailable checks if a video is available and not private/deleted
func (yt *YouTubeProvider) IsVideoAvailable(ctx context.Context, videoID string) (bool, error) {
	video, err := yt.GetVideoInfo(ctx, videoID)
	if err != nil {
		// If we can't get video info, it's likely unavailable
		return false, err
	}

	// Check for common unavailability indicators
	if video == nil {
		return false, fmt.Errorf("video not found")
	}

	if video.Title == "" {
		return false, fmt.Errorf("video has no title (likely private or deleted)")
	}

	return true, nil
}

// Ensure YouTubeProvider implements the AudioProvider interface
var _ types.AudioProvider = (*YouTubeProvider)(nil)
