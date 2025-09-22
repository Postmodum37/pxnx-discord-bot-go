package ytdlp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// ServiceUtils provides utility functions for the yt-dlp service
type ServiceUtils struct{}

// NewServiceUtils creates a new ServiceUtils instance
func NewServiceUtils() *ServiceUtils {
	return &ServiceUtils{}
}

// CheckDependencies checks if all required dependencies are available
func (su *ServiceUtils) CheckDependencies() error {
	// Check Python 3
	if err := su.checkPython(); err != nil {
		return fmt.Errorf("python dependency check failed: %w", err)
	}

	// Check yt-dlp
	if err := su.checkYTDLP(); err != nil {
		return fmt.Errorf("yt-dlp dependency check failed: %w", err)
	}

	// Check aiohttp
	if err := su.checkAiohttp(); err != nil {
		return fmt.Errorf("aiohttp dependency check failed: %w", err)
	}

	return nil
}

// checkPython checks if Python 3 is available and gets version info
func (su *ServiceUtils) checkPython() error {
	cmd := exec.Command("python3", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("python3 is not available - please install Python 3.7+: %w", err)
	}

	fmt.Printf("✓ Python found: %s", string(output))
	return nil
}

// checkYTDLP checks if yt-dlp is available and gets version info
func (su *ServiceUtils) checkYTDLP() error {
	cmd := exec.Command("python3", "-c", "import yt_dlp; print(f'yt-dlp {yt_dlp.version.__version__}')")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp is not available - install with: pip install yt-dlp")
	}

	fmt.Printf("✓ %s\n", string(output))
	return nil
}

// checkAiohttp checks if aiohttp is available
func (su *ServiceUtils) checkAiohttp() error {
	cmd := exec.Command("python3", "-c", "import aiohttp; print(f'aiohttp {aiohttp.__version__}')")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("aiohttp is not available - install with: pip install aiohttp")
	}

	fmt.Printf("✓ %s\n", string(output))
	return nil
}

// InstallDependencies attempts to install required Python dependencies
func (su *ServiceUtils) InstallDependencies() error {
	fmt.Println("Installing yt-dlp service dependencies...")

	packages := []string{"yt-dlp", "aiohttp"}

	for _, pkg := range packages {
		fmt.Printf("Installing %s...\n", pkg)
		cmd := exec.Command("pip", "install", "--upgrade", pkg)
		if err := cmd.Run(); err != nil {
			// Try pip3 if pip fails
			cmd = exec.Command("pip3", "install", "--upgrade", pkg)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to install %s: %w", pkg, err)
			}
		}
		fmt.Printf("✓ %s installed successfully\n", pkg)
	}

	return nil
}

// CreateConfigFile creates a configuration file for the yt-dlp service
func (su *ServiceUtils) CreateConfigFile(config *ServiceConfig, filePath string) error {
	configDir := filepath.Dir(filePath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Write a JSON configuration
	configJSON := fmt.Sprintf(`{
  "host": "%s",
  "port": %d,
  "max_workers": %d,
  "timeout": "%s",
  "max_retries": %d,
  "format": "%s",
  "audio_format": "%s",
  "audio_quality": "%s",
  "rate_limit": "%s",
  "sleep_interval": "%s",
  "cache_dir": "%s",
  "cache_ttl": "%s",
  "max_cache_size": %d,
  "health_check_interval": "%s"
}`,
		config.Host,
		config.Port,
		config.MaxWorkers,
		config.Timeout.String(),
		config.MaxRetries,
		config.Format,
		config.AudioFormat,
		config.AudioQuality,
		config.RateLimit,
		config.SleepInterval,
		config.CacheDir,
		config.CacheTTL.String(),
		config.MaxCacheSize,
		config.HealthCheckInterval.String(),
	)

	if _, err := file.WriteString(configJSON); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Printf("✓ Configuration file created: %s\n", filePath)
	return nil
}

// CleanupCache removes old cache files
func (su *ServiceUtils) CleanupCache(cacheDir string, maxAge time.Duration) error {
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		return nil // Cache directory doesn't exist
	}

	cutoff := time.Now().Add(-maxAge)

	err := filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.ModTime().Before(cutoff) {
			if err := os.Remove(path); err != nil {
				fmt.Printf("Warning: failed to remove cache file %s: %v\n", path, err)
			} else {
				fmt.Printf("Removed old cache file: %s\n", path)
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("cache cleanup failed: %w", err)
	}

	return nil
}

// GetServiceInfo retrieves information about the running service
func (su *ServiceUtils) GetServiceInfo(client *Client) (*ServiceInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	health, err := client.HealthCheck(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get service health: %w", err)
	}

	return &ServiceInfo{
		Status:      "running",
		Version:     health.Version,
		Uptime:      health.Uptime,
		LastCheck:   health.LastCheck,
		WorkerCount: health.WorkerCount,
	}, nil
}

// ServiceInfo contains information about the service
type ServiceInfo struct {
	Status      string    `json:"status"`
	Version     string    `json:"version"`
	Uptime      string    `json:"uptime"`
	LastCheck   time.Time `json:"last_check"`
	WorkerCount int       `json:"worker_count"`
}

// ValidateURL validates if a URL is supported by yt-dlp
func (su *ServiceUtils) ValidateURL(url string) error {
	if url == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	// Basic URL validation
	if len(url) < 10 {
		return fmt.Errorf("URL too short")
	}

	if !containsAny(url, []string{"http://", "https://"}) {
		return fmt.Errorf("URL must start with http:// or https://")
	}

	return nil
}

// containsAny checks if a string contains any of the provided substrings
func containsAny(s string, substrings []string) bool {
	for _, substr := range substrings {
		if len(s) >= len(substr) {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
		}
	}
	return false
}

// FormatBytes formats byte size into human-readable format
func (su *ServiceUtils) FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB", "PB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// FormatDuration formats duration into human-readable format
func (su *ServiceUtils) FormatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("%.1f seconds", duration.Seconds())
	}
	if duration < time.Hour {
		return fmt.Sprintf("%.1f minutes", duration.Minutes())
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%.1f hours", duration.Hours())
	}
	return fmt.Sprintf("%.1f days", duration.Hours()/24)
}

// WaitForPort waits for a port to become available
func (su *ServiceUtils) WaitForPort(host string, port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		cmd := exec.Command("nc", "-z", host, fmt.Sprintf("%d", port))
		if err := cmd.Run(); err == nil {
			return nil
		}

		time.Sleep(100 * time.Millisecond)
	}

	return fmt.Errorf("timeout waiting for %s:%d to become available", host, port)
}

// GetSystemInfo returns system information relevant to the service
func (su *ServiceUtils) GetSystemInfo() map[string]string {
	info := make(map[string]string)

	// Get Python version
	if cmd := exec.Command("python3", "--version"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			info["python_version"] = string(output)
		}
	}

	// Get yt-dlp version
	if cmd := exec.Command("python3", "-c", "import yt_dlp; print(yt_dlp.version.__version__)"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			info["ytdlp_version"] = string(output)
		}
	}

	// Get system information
	if cmd := exec.Command("uname", "-a"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			info["system"] = string(output)
		}
	}

	return info
}