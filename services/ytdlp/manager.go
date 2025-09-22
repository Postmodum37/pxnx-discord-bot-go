package ytdlp

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

// ServiceManager manages the lifecycle of the yt-dlp service
type ServiceManager struct {
	config       *ServiceConfig
	client       *Client
	cmd          *exec.Cmd
	status       int32 // Use atomic for thread-safe status updates
	mu           sync.RWMutex
	healthTicker *time.Ticker
	stopChan     chan struct{}
	errorChan    chan error
	logFile      *os.File
}

// NewServiceManager creates a new service manager
func NewServiceManager(config *ServiceConfig) *ServiceManager {
	if config == nil {
		config = DefaultServiceConfig()
	}

	client := NewClient(config)

	return &ServiceManager{
		config:    config,
		client:    client,
		status:    int32(StatusStopped),
		stopChan:  make(chan struct{}),
		errorChan: make(chan error, 10),
	}
}

// Start starts the yt-dlp service
func (sm *ServiceManager) Start(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	log.Printf("[SERVICE] Starting yt-dlp service manager")

	currentStatus := ServiceStatus(atomic.LoadInt32(&sm.status))
	if currentStatus == StatusRunning || currentStatus == StatusStarting {
		log.Printf("[SERVICE] Service already %s", currentStatus)
		return fmt.Errorf("service is already running or starting")
	}

	atomic.StoreInt32(&sm.status, int32(StatusStarting))
	log.Printf("[SERVICE] Status set to starting")

	// Check if Python is available
	log.Printf("[SERVICE] Checking Python availability...")
	if err := sm.checkPythonAvailability(); err != nil {
		log.Printf("[SERVICE] Python check failed: %v", err)
		atomic.StoreInt32(&sm.status, int32(StatusError))
		return fmt.Errorf("python check failed: %w", err)
	}
	log.Printf("[SERVICE] Python check passed")

	// Check if yt-dlp is available
	log.Printf("[SERVICE] Checking yt-dlp availability...")
	if err := sm.checkYTDLPAvailability(); err != nil {
		log.Printf("[SERVICE] yt-dlp check failed: %v", err)
		atomic.StoreInt32(&sm.status, int32(StatusError))
		return fmt.Errorf("yt-dlp check failed: %w", err)
	}
	log.Printf("[SERVICE] yt-dlp check passed")

	// Setup logging
	log.Printf("[SERVICE] Setting up logging...")
	if err := sm.setupLogging(); err != nil {
		log.Printf("[SERVICE] Logging setup failed: %v", err)
		atomic.StoreInt32(&sm.status, int32(StatusError))
		return fmt.Errorf("logging setup failed: %w", err)
	}
	log.Printf("[SERVICE] Logging setup complete")

	// Get the path to the Python server script
	log.Printf("[SERVICE] Locating server script...")
	serverPath, err := sm.getServerScriptPath()
	if err != nil {
		log.Printf("[SERVICE] Failed to locate server script: %v", err)
		atomic.StoreInt32(&sm.status, int32(StatusError))
		return fmt.Errorf("failed to locate server script: %w", err)
	}
	log.Printf("[SERVICE] Server script found at: %s", serverPath)

	// Prepare command arguments
	args := []string{
		serverPath,
		"--host", sm.config.Host,
		"--port", fmt.Sprintf("%d", sm.config.Port),
		"--workers", fmt.Sprintf("%d", sm.config.MaxWorkers),
	}
	log.Printf("[SERVICE] Command: python3 %v", args)

	// Create the command
	sm.cmd = exec.CommandContext(ctx, "python3", args...)

	// Set up environment - preserve current PATH to ensure mise Python is available
	sm.cmd.Env = append(os.Environ(),
		"PYTHONUNBUFFERED=1",
		"PYTHONPATH="+filepath.Dir(serverPath),
	)

	// Set up stdout/stderr redirection
	if sm.logFile != nil {
		sm.cmd.Stdout = sm.logFile
		sm.cmd.Stderr = sm.logFile
	}

	// Set process group for proper cleanup
	sm.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Start the service
	log.Printf("[SERVICE] Starting Python process...")
	if err := sm.cmd.Start(); err != nil {
		log.Printf("[SERVICE] Failed to start process: %v", err)
		atomic.StoreInt32(&sm.status, int32(StatusError))
		return fmt.Errorf("failed to start yt-dlp service: %w", err)
	}
	log.Printf("[SERVICE] Python process started with PID: %d", sm.cmd.Process.Pid)

	// Wait for the service to be ready
	log.Printf("[SERVICE] Waiting for service to become ready...")
	if err := sm.waitForService(ctx); err != nil {
		log.Printf("[SERVICE] Service failed to become ready: %v", err)
		atomic.StoreInt32(&sm.status, int32(StatusError))
		sm.stopProcess()
		return fmt.Errorf("service failed to become ready: %w", err)
	}
	log.Printf("[SERVICE] Service is ready!")

	atomic.StoreInt32(&sm.status, int32(StatusRunning))

	// Start monitoring
	go sm.monitorService()
	sm.startHealthChecks()

	log.Printf("[SERVICE] Service startup complete")
	return nil
}

// Stop stops the yt-dlp service
func (sm *ServiceManager) Stop(ctx context.Context) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	currentStatus := ServiceStatus(atomic.LoadInt32(&sm.status))
	if currentStatus == StatusStopped || currentStatus == StatusStopping {
		return nil
	}

	atomic.StoreInt32(&sm.status, int32(StatusStopping))

	// Stop health checks
	sm.stopHealthChecks()

	// Signal stop
	close(sm.stopChan)

	// Stop the process
	if err := sm.stopProcess(); err != nil {
		atomic.StoreInt32(&sm.status, int32(StatusError))
		return fmt.Errorf("failed to stop service process: %w", err)
	}

	// Close log file
	if sm.logFile != nil {
		sm.logFile.Close()
		sm.logFile = nil
	}

	// Close client
	if err := sm.client.Close(); err != nil {
		return fmt.Errorf("failed to close client: %w", err)
	}

	atomic.StoreInt32(&sm.status, int32(StatusStopped))
	return nil
}

// Restart restarts the yt-dlp service
func (sm *ServiceManager) Restart(ctx context.Context) error {
	if err := sm.Stop(ctx); err != nil {
		return fmt.Errorf("failed to stop service for restart: %w", err)
	}

	// Wait a moment for cleanup
	time.Sleep(1 * time.Second)

	return sm.Start(ctx)
}

// GetStatus returns the current service status
func (sm *ServiceManager) GetStatus() ServiceStatus {
	return ServiceStatus(atomic.LoadInt32(&sm.status))
}

// IsRunning returns true if the service is running
func (sm *ServiceManager) IsRunning() bool {
	return sm.GetStatus() == StatusRunning
}

// GetClient returns the service client
func (sm *ServiceManager) GetClient() *Client {
	return sm.client
}

// GetErrors returns a channel for receiving service errors
func (sm *ServiceManager) GetErrors() <-chan error {
	return sm.errorChan
}

// checkPythonAvailability checks if Python 3 is available
func (sm *ServiceManager) checkPythonAvailability() error {
	cmd := exec.Command("python3", "--version")
	cmd.Env = os.Environ() // Preserve current environment including PATH
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("python3 is not available: %w", err)
	}
	return nil
}

// checkYTDLPAvailability checks if yt-dlp is available
func (sm *ServiceManager) checkYTDLPAvailability() error {
	cmd := exec.Command("python3", "-c", "import yt_dlp; print(yt_dlp.version.__version__)")
	cmd.Env = os.Environ() // Preserve current environment including PATH
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("yt-dlp is not available - install with: pip install yt-dlp aiohttp")
	}
	return nil
}

// getServerScriptPath returns the path to the Python server script
func (sm *ServiceManager) getServerScriptPath() (string, error) {
	// Try to find the script relative to the current executable or working directory
	possiblePaths := []string{
		"services/ytdlp/server.py",
		"./services/ytdlp/server.py",
		"../services/ytdlp/server.py",
		"../../services/ytdlp/server.py",
	}

	for _, path := range possiblePaths {
		if absPath, err := filepath.Abs(path); err == nil {
			if _, err := os.Stat(absPath); err == nil {
				return absPath, nil
			}
		}
	}

	return "", fmt.Errorf("server.py script not found in expected locations")
}

// setupLogging sets up logging for the service
func (sm *ServiceManager) setupLogging() error {
	// Create logs directory if it doesn't exist
	logDir := filepath.Join(sm.config.CacheDir, "logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file
	logPath := filepath.Join(logDir, "ytdlp-service.log")
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	sm.logFile = logFile
	return nil
}

// waitForService waits for the service to become ready
func (sm *ServiceManager) waitForService(ctx context.Context) error {
	timeout := 30 * time.Second
	interval := 500 * time.Millisecond

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for service to become ready")
		case <-ticker.C:
			if _, err := sm.client.HealthCheck(context.Background()); err == nil {
				return nil
			}
		}
	}
}

// stopProcess stops the service process
func (sm *ServiceManager) stopProcess() error {
	if sm.cmd == nil || sm.cmd.Process == nil {
		return nil
	}

	// Try graceful shutdown first
	if err := sm.cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// If graceful shutdown fails, force kill
		if killErr := sm.cmd.Process.Kill(); killErr != nil {
			return fmt.Errorf("failed to kill process: %w", killErr)
		}
	}

	// Wait for process to exit
	done := make(chan error, 1)
	go func() {
		done <- sm.cmd.Wait()
	}()

	select {
	case <-done:
		return nil
	case <-time.After(10 * time.Second):
		// Force kill if process doesn't exit gracefully
		if err := sm.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to force kill process: %w", err)
		}
		return nil
	}
}

// monitorService monitors the service process
func (sm *ServiceManager) monitorService() {
	if sm.cmd == nil {
		return
	}

	err := sm.cmd.Wait()

	// Process has exited
	if atomic.LoadInt32(&sm.status) == int32(StatusRunning) {
		atomic.StoreInt32(&sm.status, int32(StatusError))

		// Send error to error channel
		select {
		case sm.errorChan <- fmt.Errorf("service process exited unexpectedly: %w", err):
		default:
			// Channel is full, skip
		}
	}
}

// startHealthChecks starts periodic health checks
func (sm *ServiceManager) startHealthChecks() {
	if sm.config.HealthCheckInterval <= 0 {
		return
	}

	sm.healthTicker = time.NewTicker(sm.config.HealthCheckInterval)

	go func() {
		for {
			select {
			case <-sm.healthTicker.C:
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				_, err := sm.client.HealthCheck(ctx)
				cancel()

				if err != nil {
					// Health check failed
					if atomic.LoadInt32(&sm.status) == int32(StatusRunning) {
						select {
						case sm.errorChan <- fmt.Errorf("health check failed: %w", err):
						default:
							// Channel is full, skip
						}
					}
				}

			case <-sm.stopChan:
				return
			}
		}
	}()
}

// stopHealthChecks stops periodic health checks
func (sm *ServiceManager) stopHealthChecks() {
	if sm.healthTicker != nil {
		sm.healthTicker.Stop()
		sm.healthTicker = nil
	}
}