package ytdlp

import (
	"time"
)

// ServiceConfig holds configuration for the yt-dlp service
type ServiceConfig struct {
	// Service settings
	Host        string        `json:"host"`
	Port        int           `json:"port"`
	MaxWorkers  int           `json:"max_workers"`
	Timeout     time.Duration `json:"timeout"`
	MaxRetries  int           `json:"max_retries"`

	// yt-dlp settings
	Format      string `json:"format"`
	AudioFormat string `json:"audio_format"`
	AudioQuality string `json:"audio_quality"`

	// Rate limiting
	RateLimit      string `json:"rate_limit"`
	SleepInterval  string `json:"sleep_interval"`

	// Cache settings
	CacheDir       string        `json:"cache_dir"`
	CacheTTL       time.Duration `json:"cache_ttl"`
	MaxCacheSize   int64         `json:"max_cache_size"`

	// Health check settings
	HealthCheckInterval time.Duration `json:"health_check_interval"`
}

// DefaultServiceConfig returns a default configuration
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		Host:        "localhost",
		Port:        8080,
		MaxWorkers:  4,
		Timeout:     30 * time.Second,
		MaxRetries:  3,

		Format:      "bestaudio/best",
		AudioFormat: "opus",
		AudioQuality: "128K",

		RateLimit:     "1M",
		SleepInterval: "0",

		CacheDir:     "/tmp/ytdlp-cache",
		CacheTTL:     24 * time.Hour,
		MaxCacheSize: 1024 * 1024 * 1024, // 1GB

		HealthCheckInterval: 30 * time.Second,
	}
}

// SearchRequest represents a search request to yt-dlp
type SearchRequest struct {
	Query      string `json:"query"`
	MaxResults int    `json:"max_results,omitempty"`
	Type       string `json:"type,omitempty"` // "video", "playlist", etc.
}

// ExtractRequest represents an extraction request for a specific URL
type ExtractRequest struct {
	URL    string `json:"url"`
	Format string `json:"format,omitempty"`
}

// VideoInfo represents video information returned by yt-dlp
type VideoInfo struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description,omitempty"`
	Duration    float64           `json:"duration,omitempty"`
	URL         string            `json:"webpage_url"`
	Thumbnail   string            `json:"thumbnail,omitempty"`
	Uploader    string            `json:"uploader,omitempty"`
	UploadDate  string            `json:"upload_date,omitempty"`
	ViewCount   int64             `json:"view_count,omitempty"`
	Formats     []FormatInfo      `json:"formats,omitempty"`
	Thumbnails  []ThumbnailInfo   `json:"thumbnails,omitempty"`
	Extractor   string            `json:"extractor"`
	ExtractorKey string           `json:"extractor_key"`
	Available   bool              `json:"available"`
	LiveStatus  string            `json:"live_status,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Categories  []string          `json:"categories,omitempty"`
}

// FormatInfo represents format information for a video
type FormatInfo struct {
	FormatID   string  `json:"format_id"`
	URL        string  `json:"url"`
	Ext        string  `json:"ext"`
	Format     string  `json:"format"`
	Protocol   string  `json:"protocol,omitempty"`
	VCodec     string  `json:"vcodec,omitempty"`
	ACodec     string  `json:"acodec,omitempty"`
	Width      int     `json:"width,omitempty"`
	Height     int     `json:"height,omitempty"`
	FPS        float64 `json:"fps,omitempty"`
	TBR        float64 `json:"tbr,omitempty"`
	VBR        float64 `json:"vbr,omitempty"`
	ABR        float64 `json:"abr,omitempty"`
	ASR        int     `json:"asr,omitempty"`
	Filesize   int64   `json:"filesize,omitempty"`
	Quality    int     `json:"quality,omitempty"`
	Language   string  `json:"language,omitempty"`
	Preference int     `json:"preference,omitempty"`
}

// ThumbnailInfo represents thumbnail information
type ThumbnailInfo struct {
	ID         string `json:"id,omitempty"`
	URL        string `json:"url"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	Resolution string `json:"resolution,omitempty"`
}

// SearchResult represents a search result from yt-dlp
type SearchResult struct {
	Videos     []VideoInfo `json:"videos"`
	TotalCount int         `json:"total_count"`
	Query      string      `json:"query"`
}

// ServiceResponse represents the response from yt-dlp service
type ServiceResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Code    int         `json:"code,omitempty"`
}

// HealthStatus represents the health status of the yt-dlp service
type HealthStatus struct {
	Status      string    `json:"status"`
	Version     string    `json:"version,omitempty"`
	Uptime      string    `json:"uptime,omitempty"`
	LastCheck   time.Time `json:"last_check"`
	WorkerCount int       `json:"worker_count,omitempty"`
	QueueSize   int       `json:"queue_size,omitempty"`
}

// ServiceError represents an error from the yt-dlp service
type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type,omitempty"`
	Details string `json:"details,omitempty"`
}

func (e *ServiceError) Error() string {
	if e.Details != "" {
		return e.Message + ": " + e.Details
	}
	return e.Message
}

// ServiceStatus represents the current status of the service
type ServiceStatus int

const (
	StatusStopped ServiceStatus = iota
	StatusStarting
	StatusRunning
	StatusStopping
	StatusError
)

func (s ServiceStatus) String() string {
	switch s {
	case StatusStopped:
		return "stopped"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusStopping:
		return "stopping"
	case StatusError:
		return "error"
	default:
		return "unknown"
	}
}