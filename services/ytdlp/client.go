package ytdlp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"sync"
	"time"
)

// Client represents a client for the yt-dlp service
type Client struct {
	baseURL    string
	httpClient *http.Client
	config     *ServiceConfig
	mu         sync.RWMutex
	lastHealth *HealthStatus
}

// NewClient creates a new yt-dlp service client
func NewClient(config *ServiceConfig) *Client {
	if config == nil {
		config = DefaultServiceConfig()
	}

	baseURL := fmt.Sprintf("http://%s:%d", config.Host, config.Port)

	// Create HTTP client with appropriate timeouts
	httpClient := &http.Client{
		Timeout: config.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   10 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:          10,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			MaxIdleConnsPerHost:   5,
		},
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		config:     config,
	}
}

// HealthCheck checks if the yt-dlp service is healthy
func (c *Client) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/health", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create health check request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("health check request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read health check response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("health check failed with status %d: %s", resp.StatusCode, string(body))
	}

	var serviceResp ServiceResponse
	if err := json.Unmarshal(body, &serviceResp); err != nil {
		return nil, fmt.Errorf("failed to parse health check response: %w", err)
	}

	if !serviceResp.Success {
		return nil, fmt.Errorf("health check failed: %s", serviceResp.Error)
	}

	healthData, ok := serviceResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid health check response format")
	}

	health := &HealthStatus{
		Status:    getStringFromMap(healthData, "status"),
		Version:   getStringFromMap(healthData, "version"),
		Uptime:    getStringFromMap(healthData, "uptime"),
		LastCheck: time.Now(),
	}

	if workerCount, ok := healthData["worker_count"].(float64); ok {
		health.WorkerCount = int(workerCount)
	}

	if queueSize, ok := healthData["queue_size"].(float64); ok {
		health.QueueSize = int(queueSize)
	}

	// Cache the health status
	c.mu.Lock()
	c.lastHealth = health
	c.mu.Unlock()

	return health, nil
}

// IsHealthy returns true if the service was healthy on the last check
func (c *Client) IsHealthy() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.lastHealth == nil {
		return false
	}

	// Consider healthy if the last check was within the health check interval
	return time.Since(c.lastHealth.LastCheck) < c.config.HealthCheckInterval*2
}

// GetLastHealthStatus returns the last known health status
func (c *Client) GetLastHealthStatus() *HealthStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.lastHealth == nil {
		return nil
	}

	// Return a copy to prevent race conditions
	healthCopy := *c.lastHealth
	return &healthCopy
}

// ExtractInfo extracts video information from a URL
func (c *Client) ExtractInfo(ctx context.Context, url string) (*VideoInfo, error) {
	return c.ExtractInfoWithFormat(ctx, url, "")
}

// ExtractInfoWithFormat extracts video information with a specific format
func (c *Client) ExtractInfoWithFormat(ctx context.Context, url, format string) (*VideoInfo, error) {
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	req := &ExtractRequest{
		URL:    url,
		Format: format,
	}

	resp, err := c.makeRequest(ctx, "POST", "/extract", req)
	if err != nil {
		return nil, fmt.Errorf("extract request failed: %w", err)
	}

	if !resp.Success {
		return nil, &ServiceError{
			Code:    resp.Code,
			Message: resp.Error,
			Type:    "extraction_failed",
		}
	}

	videoData, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response format from yt-dlp service")
	}

	return c.parseVideoInfo(videoData)
}

// Search searches for videos using the provided query
func (c *Client) Search(ctx context.Context, query string, maxResults int) (*SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query cannot be empty")
	}

	if maxResults <= 0 {
		maxResults = 10
	}

	if maxResults > 50 {
		maxResults = 50 // Cap at 50 results
	}

	req := &SearchRequest{
		Query:      query,
		MaxResults: maxResults,
	}

	resp, err := c.makeRequest(ctx, "POST", "/search", req)
	if err != nil {
		return nil, fmt.Errorf("search request failed: %w", err)
	}

	if !resp.Success {
		return nil, &ServiceError{
			Code:    resp.Code,
			Message: resp.Error,
			Type:    "search_failed",
		}
	}

	searchData, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid search response format")
	}

	return c.parseSearchResult(searchData)
}

// ClearCache clears the service cache
func (c *Client) ClearCache(ctx context.Context) error {
	resp, err := c.makeRequest(ctx, "POST", "/cache/clear", nil)
	if err != nil {
		return fmt.Errorf("clear cache request failed: %w", err)
	}

	if !resp.Success {
		return &ServiceError{
			Code:    resp.Code,
			Message: resp.Error,
			Type:    "cache_clear_failed",
		}
	}

	return nil
}

// makeRequest makes an HTTP request to the yt-dlp service
func (c *Client) makeRequest(ctx context.Context, method, endpoint string, payload interface{}) (*ServiceResponse, error) {
	var body io.Reader

	if payload != nil {
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request payload: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var serviceResp ServiceResponse
	if err := json.Unmarshal(respBody, &serviceResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Set the HTTP status code if not set by the service
	if serviceResp.Code == 0 {
		serviceResp.Code = resp.StatusCode
	}

	return &serviceResp, nil
}

// parseVideoInfo parses video information from the service response
func (c *Client) parseVideoInfo(data map[string]interface{}) (*VideoInfo, error) {
	video := &VideoInfo{
		ID:           getStringFromMap(data, "id"),
		Title:        getStringFromMap(data, "title"),
		Description:  getStringFromMap(data, "description"),
		URL:          getStringFromMap(data, "webpage_url"),
		Thumbnail:    getStringFromMap(data, "thumbnail"),
		Uploader:     getStringFromMap(data, "uploader"),
		UploadDate:   getStringFromMap(data, "upload_date"),
		Extractor:    getStringFromMap(data, "extractor"),
		ExtractorKey: getStringFromMap(data, "extractor_key"),
		LiveStatus:   getStringFromMap(data, "live_status"),
		Available:    getBoolFromMap(data, "available"),
	}

	if duration, ok := data["duration"].(float64); ok {
		video.Duration = duration
	}

	if viewCount, ok := data["view_count"].(float64); ok {
		video.ViewCount = int64(viewCount)
	}

	// Parse tags
	if tagsInterface, ok := data["tags"].([]interface{}); ok {
		tags := make([]string, 0, len(tagsInterface))
		for _, tag := range tagsInterface {
			if tagStr, ok := tag.(string); ok {
				tags = append(tags, tagStr)
			}
		}
		video.Tags = tags
	}

	// Parse categories
	if categoriesInterface, ok := data["categories"].([]interface{}); ok {
		categories := make([]string, 0, len(categoriesInterface))
		for _, category := range categoriesInterface {
			if categoryStr, ok := category.(string); ok {
				categories = append(categories, categoryStr)
			}
		}
		video.Categories = categories
	}

	// Parse formats
	if formatsInterface, ok := data["formats"].([]interface{}); ok {
		formats := make([]FormatInfo, 0, len(formatsInterface))
		for _, formatInterface := range formatsInterface {
			if formatData, ok := formatInterface.(map[string]interface{}); ok {
				format := FormatInfo{
					FormatID: getStringFromMap(formatData, "format_id"),
					URL:      getStringFromMap(formatData, "url"),
					Ext:      getStringFromMap(formatData, "ext"),
					Format:   getStringFromMap(formatData, "format"),
					Protocol: getStringFromMap(formatData, "protocol"),
					VCodec:   getStringFromMap(formatData, "vcodec"),
					ACodec:   getStringFromMap(formatData, "acodec"),
					Language: getStringFromMap(formatData, "language"),
				}

				if width, ok := formatData["width"].(float64); ok {
					format.Width = int(width)
				}
				if height, ok := formatData["height"].(float64); ok {
					format.Height = int(height)
				}
				if fps, ok := formatData["fps"].(float64); ok {
					format.FPS = fps
				}
				if tbr, ok := formatData["tbr"].(float64); ok {
					format.TBR = tbr
				}
				if vbr, ok := formatData["vbr"].(float64); ok {
					format.VBR = vbr
				}
				if abr, ok := formatData["abr"].(float64); ok {
					format.ABR = abr
				}
				if asr, ok := formatData["asr"].(float64); ok {
					format.ASR = int(asr)
				}
				if filesize, ok := formatData["filesize"].(float64); ok {
					format.Filesize = int64(filesize)
				}
				if quality, ok := formatData["quality"].(float64); ok {
					format.Quality = int(quality)
				}
				if preference, ok := formatData["preference"].(float64); ok {
					format.Preference = int(preference)
				}

				formats = append(formats, format)
			}
		}
		video.Formats = formats
	}

	// Parse thumbnails
	if thumbnailsInterface, ok := data["thumbnails"].([]interface{}); ok {
		thumbnails := make([]ThumbnailInfo, 0, len(thumbnailsInterface))
		for _, thumbnailInterface := range thumbnailsInterface {
			if thumbnailData, ok := thumbnailInterface.(map[string]interface{}); ok {
				thumbnail := ThumbnailInfo{
					ID:         getStringFromMap(thumbnailData, "id"),
					URL:        getStringFromMap(thumbnailData, "url"),
					Resolution: getStringFromMap(thumbnailData, "resolution"),
				}

				if width, ok := thumbnailData["width"].(float64); ok {
					thumbnail.Width = int(width)
				}
				if height, ok := thumbnailData["height"].(float64); ok {
					thumbnail.Height = int(height)
				}

				thumbnails = append(thumbnails, thumbnail)
			}
		}
		video.Thumbnails = thumbnails
	}

	return video, nil
}

// parseSearchResult parses search results from the service response
func (c *Client) parseSearchResult(data map[string]interface{}) (*SearchResult, error) {
	result := &SearchResult{
		Query: getStringFromMap(data, "query"),
	}

	if totalCount, ok := data["total_count"].(float64); ok {
		result.TotalCount = int(totalCount)
	}

	if videosInterface, ok := data["videos"].([]interface{}); ok {
		videos := make([]VideoInfo, 0, len(videosInterface))
		for _, videoInterface := range videosInterface {
			if videoData, ok := videoInterface.(map[string]interface{}); ok {
				video, err := c.parseVideoInfo(videoData)
				if err != nil {
					continue // Skip invalid videos
				}
				videos = append(videos, *video)
			}
		}
		result.Videos = videos
	}

	return result, nil
}

// Helper functions for safely extracting values from maps

func getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

// Close closes the client and cleans up resources
func (c *Client) Close() error {
	// Close idle connections
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		transport.CloseIdleConnections()
	}
	return nil
}