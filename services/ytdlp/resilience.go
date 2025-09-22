package ytdlp

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// RetryConfig defines configuration for retry operations
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`
	InitialDelay    time.Duration `json:"initial_delay"`
	MaxDelay        time.Duration `json:"max_delay"`
	BackoffFactor   float64       `json:"backoff_factor"`
	RandomJitter    bool          `json:"random_jitter"`
	RetryableErrors []string      `json:"retryable_errors"`
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RandomJitter:  true,
		RetryableErrors: []string{
			"connection refused",
			"timeout",
			"temporary failure",
			"network unreachable",
			"connection reset",
			"service unavailable",
			"too many requests",
		},
	}
}

// CircuitBreakerConfig defines configuration for circuit breaker
type CircuitBreakerConfig struct {
	FailureThreshold   int           `json:"failure_threshold"`
	SuccessThreshold   int           `json:"success_threshold"`
	Timeout           time.Duration `json:"timeout"`
	ResetTimeout      time.Duration `json:"reset_timeout"`
	MaxConcurrentRequests int       `json:"max_concurrent_requests"`
}

// DefaultCircuitBreakerConfig returns a default circuit breaker configuration
func DefaultCircuitBreakerConfig() *CircuitBreakerConfig {
	return &CircuitBreakerConfig{
		FailureThreshold:      5,
		SuccessThreshold:      3,
		Timeout:              30 * time.Second,
		ResetTimeout:         60 * time.Second,
		MaxConcurrentRequests: 10,
	}
}

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateOpen
	StateHalfOpen
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

// CircuitBreaker implements the circuit breaker pattern for service resilience
type CircuitBreaker struct {
	config           *CircuitBreakerConfig
	state            CircuitBreakerState
	failures         int
	successes        int
	lastFailure      time.Time
	nextAttempt      time.Time
	activeRequests   int
	mu               sync.RWMutex
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig()
	}

	return &CircuitBreaker{
		config: config,
		state:  StateClosed,
	}
}

// Execute executes a function with circuit breaker protection
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	if !cb.canExecute() {
		return fmt.Errorf("circuit breaker is open")
	}

	cb.beforeRequest()
	defer cb.afterRequest()

	// Create a timeout context
	reqCtx, cancel := context.WithTimeout(ctx, cb.config.Timeout)
	defer cancel()

	err := fn(reqCtx)

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

// canExecute checks if a request can be executed
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return cb.activeRequests < cb.config.MaxConcurrentRequests
	case StateOpen:
		if time.Now().After(cb.nextAttempt) {
			cb.state = StateHalfOpen
			cb.successes = 0
			return true
		}
		return false
	case StateHalfOpen:
		return cb.activeRequests < 1 // Only allow one request in half-open
	default:
		return false
	}
}

// beforeRequest is called before executing a request
func (cb *CircuitBreaker) beforeRequest() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.activeRequests++
}

// afterRequest is called after executing a request
func (cb *CircuitBreaker) afterRequest() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.activeRequests--
}

// onSuccess is called when a request succeeds
func (cb *CircuitBreaker) onSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures = 0

	switch cb.state {
	case StateHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.state = StateClosed
		}
	}
}

// onFailure is called when a request fails
func (cb *CircuitBreaker) onFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failures++
	cb.lastFailure = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failures >= cb.config.FailureThreshold {
			cb.state = StateOpen
			cb.nextAttempt = time.Now().Add(cb.config.ResetTimeout)
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.nextAttempt = time.Now().Add(cb.config.ResetTimeout)
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// GetMetrics returns circuit breaker metrics
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":            cb.state.String(),
		"failures":         cb.failures,
		"successes":        cb.successes,
		"active_requests":  cb.activeRequests,
		"last_failure":     cb.lastFailure,
		"next_attempt":     cb.nextAttempt,
	}
}

// ResilientClient wraps the yt-dlp client with resilience patterns
type ResilientClient struct {
	client         *Client
	circuitBreaker *CircuitBreaker
	retryConfig    *RetryConfig
}

// NewResilientClient creates a new resilient client
func NewResilientClient(config *ServiceConfig) *ResilientClient {
	client := NewClient(config)
	circuitBreaker := NewCircuitBreaker(DefaultCircuitBreakerConfig())
	retryConfig := DefaultRetryConfig()

	return &ResilientClient{
		client:         client,
		circuitBreaker: circuitBreaker,
		retryConfig:    retryConfig,
	}
}

// NewResilientClientWithConfigs creates a resilient client with custom configurations
func NewResilientClientWithConfigs(
	serviceConfig *ServiceConfig,
	cbConfig *CircuitBreakerConfig,
	retryConfig *RetryConfig,
) *ResilientClient {
	client := NewClient(serviceConfig)
	circuitBreaker := NewCircuitBreaker(cbConfig)

	if retryConfig == nil {
		retryConfig = DefaultRetryConfig()
	}

	return &ResilientClient{
		client:         client,
		circuitBreaker: circuitBreaker,
		retryConfig:    retryConfig,
	}
}

// HealthCheck performs a health check with resilience
func (rc *ResilientClient) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	var result *HealthStatus
	var err error

	retryErr := rc.withRetry(ctx, func(ctx context.Context) error {
		result, err = rc.client.HealthCheck(ctx)
		return err
	})

	if retryErr != nil {
		return nil, retryErr
	}

	return result, nil
}

// ExtractInfo extracts video information with resilience
func (rc *ResilientClient) ExtractInfo(ctx context.Context, url string) (*VideoInfo, error) {
	var result *VideoInfo
	var err error

	retryErr := rc.withRetry(ctx, func(ctx context.Context) error {
		result, err = rc.client.ExtractInfo(ctx, url)
		return err
	})

	if retryErr != nil {
		return nil, retryErr
	}

	return result, nil
}

// ExtractInfoWithFormat extracts video information with format and resilience
func (rc *ResilientClient) ExtractInfoWithFormat(ctx context.Context, url, format string) (*VideoInfo, error) {
	var result *VideoInfo
	var err error

	retryErr := rc.withRetry(ctx, func(ctx context.Context) error {
		result, err = rc.client.ExtractInfoWithFormat(ctx, url, format)
		return err
	})

	if retryErr != nil {
		return nil, retryErr
	}

	return result, nil
}

// Search searches for videos with resilience
func (rc *ResilientClient) Search(ctx context.Context, query string, maxResults int) (*SearchResult, error) {
	var result *SearchResult
	var err error

	retryErr := rc.withRetry(ctx, func(ctx context.Context) error {
		result, err = rc.client.Search(ctx, query, maxResults)
		return err
	})

	if retryErr != nil {
		return nil, retryErr
	}

	return result, nil
}

// ClearCache clears the cache with resilience
func (rc *ResilientClient) ClearCache(ctx context.Context) error {
	return rc.withRetry(ctx, func(ctx context.Context) error {
		return rc.client.ClearCache(ctx)
	})
}

// withRetry executes a function with retry logic and circuit breaker
func (rc *ResilientClient) withRetry(ctx context.Context, fn func(context.Context) error) error {
	return rc.circuitBreaker.Execute(ctx, func(ctx context.Context) error {
		return rc.executeWithRetry(ctx, fn)
	})
}

// executeWithRetry executes a function with exponential backoff retry
func (rc *ResilientClient) executeWithRetry(ctx context.Context, fn func(context.Context) error) error {
	var lastErr error

	for attempt := 0; attempt <= rc.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := rc.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !rc.isRetryableError(err) {
			return err
		}

		// Don't retry on the last attempt
		if attempt == rc.retryConfig.MaxRetries {
			break
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// calculateDelay calculates the delay for exponential backoff
func (rc *ResilientClient) calculateDelay(attempt int) time.Duration {
	delay := float64(rc.retryConfig.InitialDelay) * math.Pow(rc.retryConfig.BackoffFactor, float64(attempt-1))

	if rc.retryConfig.RandomJitter {
		// Add random jitter (±25%)
		jitter := delay * 0.25
		randomFactor := float64(time.Now().UnixNano()%1000) / 1000.0
		delay += (2*jitter*randomFactor - jitter)
	}

	delayDuration := time.Duration(delay)
	if delayDuration > rc.retryConfig.MaxDelay {
		delayDuration = rc.retryConfig.MaxDelay
	}

	return delayDuration
}

// isRetryableError checks if an error is retryable
func (rc *ResilientClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errorMsg := err.Error()
	for _, retryableError := range rc.retryConfig.RetryableErrors {
		if containsIgnoreCase(errorMsg, retryableError) {
			return true
		}
	}

	// Check for specific service error codes
	if serviceErr, ok := err.(*ServiceError); ok {
		switch serviceErr.Code {
		case 429: // Too Many Requests
			return true
		case 502, 503, 504: // Bad Gateway, Service Unavailable, Gateway Timeout
			return true
		case 500: // Internal Server Error (sometimes temporary)
			return true
		}
	}

	return false
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		   (s == substr ||
		    (len(s) > len(substr) &&
		     (s[:len(substr)] == substr ||
		      s[len(s)-len(substr):] == substr ||
		      containsIgnoreCase(s[1:], substr))))
}

// GetCircuitBreakerMetrics returns circuit breaker metrics
func (rc *ResilientClient) GetCircuitBreakerMetrics() map[string]interface{} {
	return rc.circuitBreaker.GetMetrics()
}

// GetCircuitBreakerState returns the current circuit breaker state
func (rc *ResilientClient) GetCircuitBreakerState() CircuitBreakerState {
	return rc.circuitBreaker.GetState()
}

// Close closes the resilient client
func (rc *ResilientClient) Close() error {
	return rc.client.Close()
}