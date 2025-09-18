package providers

import (
	"context"
	"testing"
	"time"

	"github.com/kkdai/youtube/v2"
	"github.com/stretchr/testify/assert"

	"pxnx-discord-bot/music/types"
)

func TestNewYouTubeProvider(t *testing.T) {
	provider := NewYouTubeProvider()
	assert.NotNil(t, provider)
	assert.NotNil(t, provider.client)
	assert.Equal(t, "youtube", provider.GetProviderName())
}

func TestYouTubeProvider_SupportsURL(t *testing.T) {
	provider := NewYouTubeProvider()

	testCases := []struct {
		url      string
		expected bool
		name     string
	}{
		{
			url:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: true,
			name:     "Standard YouTube URL",
		},
		{
			url:      "https://youtu.be/dQw4w9WgXcQ",
			expected: true,
			name:     "Short YouTube URL",
		},
		{
			url:      "https://youtube.com/watch?v=dQw4w9WgXcQ",
			expected: true,
			name:     "YouTube without www",
		},
		{
			url:      "https://m.youtube.com/watch?v=dQw4w9WgXcQ",
			expected: true,
			name:     "Mobile YouTube URL",
		},
		{
			url:      "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=30s",
			expected: true,
			name:     "YouTube URL with timestamp",
		},
		{
			url:      "https://www.youtube.com/watch?list=PLx&v=dQw4w9WgXcQ",
			expected: true,
			name:     "YouTube URL with playlist",
		},
		{
			url:      "https://spotify.com/track/123",
			expected: false,
			name:     "Spotify URL",
		},
		{
			url:      "https://soundcloud.com/user/track",
			expected: false,
			name:     "SoundCloud URL",
		},
		{
			url:      "not-a-url",
			expected: false,
			name:     "Invalid URL",
		},
		{
			url:      "https://www.youtube.com/playlist?list=PLx",
			expected: false,
			name:     "YouTube playlist URL (no video ID)",
		},
		{
			url:      "https://www.youtube.com/watch?v=invalid",
			expected: false,
			name:     "YouTube URL with invalid video ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := provider.SupportsURL(tc.url)
			assert.Equal(t, tc.expected, result, "URL: %s", tc.url)
		})
	}
}

func TestYouTubeProvider_ExtractVideoID(t *testing.T) {
	provider := NewYouTubeProvider()

	testCases := []struct {
		url         string
		expectedID  string
		expectError bool
		name        string
	}{
		{
			url:         "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			expectedID:  "dQw4w9WgXcQ",
			expectError: false,
			name:        "Standard YouTube URL",
		},
		{
			url:         "https://youtu.be/dQw4w9WgXcQ",
			expectedID:  "dQw4w9WgXcQ",
			expectError: false,
			name:        "Short YouTube URL",
		},
		{
			url:         "https://youtube.com/watch?v=dQw4w9WgXcQ",
			expectedID:  "dQw4w9WgXcQ",
			expectError: false,
			name:        "YouTube without www",
		},
		{
			url:         "https://m.youtube.com/watch?v=dQw4w9WgXcQ",
			expectedID:  "dQw4w9WgXcQ",
			expectError: false,
			name:        "Mobile YouTube URL",
		},
		{
			url:         "https://www.youtube.com/watch?v=dQw4w9WgXcQ&t=30s",
			expectedID:  "dQw4w9WgXcQ",
			expectError: false,
			name:        "YouTube URL with parameters",
		},
		{
			url:         "https://spotify.com/track/123",
			expectedID:  "",
			expectError: true,
			name:        "Non-YouTube URL",
		},
		{
			url:         "invalid-url",
			expectedID:  "",
			expectError: true,
			name:        "Invalid URL format",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			id, err := provider.extractVideoID(tc.url)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedID, id)
			}
		})
	}
}

func TestYouTubeProvider_ValidateURL(t *testing.T) {
	provider := NewYouTubeProvider()

	// Valid YouTube URL
	validURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"
	parsedURL, err := provider.ValidateURL(validURL)
	assert.NoError(t, err)
	assert.NotNil(t, parsedURL)
	assert.Equal(t, "www.youtube.com", parsedURL.Host)

	// Invalid YouTube URL
	invalidURL := "https://spotify.com/track/123"
	_, err = provider.ValidateURL(invalidURL)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not a valid YouTube URL")

	// Malformed URL
	malformedURL := "not-a-url"
	_, err = provider.ValidateURL(malformedURL)
	assert.Error(t, err)
}

func TestYouTubeProvider_FormatDuration(t *testing.T) {
	provider := NewYouTubeProvider()

	testCases := []struct {
		duration time.Duration
		expected string
		name     string
	}{
		{
			duration: 0,
			expected: "0:00",
			name:     "Zero duration",
		},
		{
			duration: 30 * time.Second,
			expected: "0:30",
			name:     "30 seconds",
		},
		{
			duration: 2*time.Minute + 15*time.Second,
			expected: "2:15",
			name:     "2 minutes 15 seconds",
		},
		{
			duration: 1*time.Hour + 30*time.Minute + 45*time.Second,
			expected: "1:30:45",
			name:     "1 hour 30 minutes 45 seconds",
		},
		{
			duration: 10*time.Hour + 5*time.Minute + 3*time.Second,
			expected: "10:05:03",
			name:     "10 hours with leading zeros",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := provider.formatDuration(tc.duration)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestYouTubeProvider_VideoToAudioSource(t *testing.T) {
	provider := NewYouTubeProvider()

	// Test with nil video
	_, err := provider.videoToAudioSource(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "video is nil")

	// Create a mock video with formats
	mockVideo := &youtube.Video{
		ID:       "dQw4w9WgXcQ",
		Title:    "Test Video",
		Duration: 3*time.Minute + 45*time.Second,
		Formats: []youtube.Format{
			{
				ItagNo:   140,
				MimeType: "audio/mp4; codecs=\"mp4a.40.2\"",
				Bitrate:  128,
				URL:      "https://example.com/audio.mp4",
			},
			{
				ItagNo:        251,
				MimeType:      "audio/webm; codecs=\"opus\"",
				Bitrate:       160,
				AudioChannels: 2,
				URL:           "https://example.com/audio.webm",
			},
		},
		Thumbnails: []youtube.Thumbnail{
			{
				URL:    "https://example.com/thumb_small.jpg",
				Width:  120,
				Height: 90,
			},
			{
				URL:    "https://example.com/thumb_large.jpg",
				Width:  1280,
				Height: 720,
			},
		},
	}

	audioSource, err := provider.videoToAudioSource(mockVideo)
	assert.NoError(t, err)
	assert.NotNil(t, audioSource)

	assert.Equal(t, "Test Video", audioSource.Title)
	assert.Equal(t, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", audioSource.URL)
	assert.Equal(t, "3:45", audioSource.Duration)
	assert.Equal(t, "youtube", audioSource.Provider)
	assert.Equal(t, "https://example.com/thumb_large.jpg", audioSource.Thumbnail)
	assert.Equal(t, "https://example.com/audio.webm", audioSource.StreamURL) // Should prefer opus
}

func TestYouTubeProvider_GetBestAudioFormat(t *testing.T) {
	provider := NewYouTubeProvider()

	testCases := []struct {
		name           string
		formats        []youtube.Format
		expectedFormat *youtube.Format
		expectError    bool
		expectedItagNo int
	}{
		{
			name:        "No formats",
			formats:     []youtube.Format{},
			expectError: true,
		},
		{
			name: "Audio-only formats - prefer opus",
			formats: []youtube.Format{
				{
					ItagNo:   140,
					MimeType: "audio/mp4; codecs=\"mp4a.40.2\"",
					Bitrate:  128,
				},
				{
					ItagNo:   251,
					MimeType: "audio/webm; codecs=\"opus\"",
					Bitrate:  160,
				},
			},
			expectedItagNo: 251, // Should prefer opus
		},
		{
			name: "Audio-only formats - prefer higher bitrate",
			formats: []youtube.Format{
				{
					ItagNo:   140,
					MimeType: "audio/mp4; codecs=\"mp4a.40.2\"",
					Bitrate:  128,
				},
				{
					ItagNo:   141,
					MimeType: "audio/mp4; codecs=\"mp4a.40.2\"",
					Bitrate:  256,
				},
			},
			expectedItagNo: 141, // Should prefer higher bitrate
		},
		{
			name: "Mixed formats - prefer audio-only",
			formats: []youtube.Format{
				{
					ItagNo:        18,
					MimeType:      "video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"",
					Bitrate:       512,
					AudioChannels: 2,
				},
				{
					ItagNo:   140,
					MimeType: "audio/mp4; codecs=\"mp4a.40.2\"",
					Bitrate:  128,
				},
			},
			expectedItagNo: 140, // Should prefer audio-only even with lower bitrate
		},
		{
			name: "Video formats with audio",
			formats: []youtube.Format{
				{
					ItagNo:        18,
					MimeType:      "video/mp4; codecs=\"avc1.42001E, mp4a.40.2\"",
					Bitrate:       512,
					AudioChannels: 2,
				},
				{
					ItagNo:        22,
					MimeType:      "video/mp4; codecs=\"avc1.64001F, mp4a.40.2\"",
					Bitrate:       1024,
					AudioChannels: 2,
				},
			},
			expectedItagNo: 22, // Should prefer higher bitrate video with audio
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			video := &youtube.Video{
				Formats: tc.formats,
			}

			format, err := provider.getBestAudioFormat(video)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, format)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, format)
				assert.Equal(t, tc.expectedItagNo, format.ItagNo)
			}
		})
	}
}

func TestYouTubeProvider_GetBestThumbnail(t *testing.T) {
	provider := NewYouTubeProvider()

	// Test with no thumbnails
	video := &youtube.Video{Thumbnails: []youtube.Thumbnail{}}
	thumbnail := provider.getBestThumbnail(video)
	assert.Empty(t, thumbnail)

	// Test with multiple thumbnails
	video = &youtube.Video{
		Thumbnails: []youtube.Thumbnail{
			{
				URL:    "https://example.com/thumb_small.jpg",
				Width:  120,
				Height: 90,
			},
			{
				URL:    "https://example.com/thumb_medium.jpg",
				Width:  320,
				Height: 180,
			},
			{
				URL:    "https://example.com/thumb_large.jpg",
				Width:  1280,
				Height: 720,
			},
		},
	}

	thumbnail = provider.getBestThumbnail(video)
	assert.Equal(t, "https://example.com/thumb_large.jpg", thumbnail)
}

func TestYouTubeProvider_Search(t *testing.T) {
	provider := NewYouTubeProvider()
	ctx := context.Background()

	// Search is not implemented in this demo version
	results, err := provider.Search(ctx, "test query", 5)
	assert.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "search functionality requires YouTube Data API v3 key")
}

func TestYouTubeProvider_InterfaceCompliance(t *testing.T) {
	// Compile-time check that YouTubeProvider implements AudioProvider interface
	var _ types.AudioProvider = (*YouTubeProvider)(nil)

	// Runtime check with actual instance
	provider := NewYouTubeProvider()
	var audioProvider types.AudioProvider = provider

	assert.NotNil(t, audioProvider)
	assert.Equal(t, "youtube", audioProvider.GetProviderName())

	// Test URL support
	assert.True(t, audioProvider.SupportsURL("https://www.youtube.com/watch?v=dQw4w9WgXcQ"))
	assert.False(t, audioProvider.SupportsURL("https://spotify.com/track/123"))
}

// BenchmarkYouTubeProvider_SupportsURL benchmarks URL validation
func BenchmarkYouTubeProvider_SupportsURL(b *testing.B) {
	provider := NewYouTubeProvider()
	testURLs := []string{
		"https://www.youtube.com/watch?v=dQw4w9WgXcQ",
		"https://youtu.be/dQw4w9WgXcQ",
		"https://spotify.com/track/123",
		"not-a-url",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		url := testURLs[i%len(testURLs)]
		provider.SupportsURL(url)
	}
}

// BenchmarkYouTubeProvider_ExtractVideoID benchmarks video ID extraction
func BenchmarkYouTubeProvider_ExtractVideoID(b *testing.B) {
	provider := NewYouTubeProvider()
	testURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		provider.extractVideoID(testURL)
	}
}
