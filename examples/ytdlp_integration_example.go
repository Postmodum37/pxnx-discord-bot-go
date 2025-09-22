package examples

import (
	"context"
	"fmt"
	"log"
	"time"

	"pxnx-discord-bot/music/manager"
	"pxnx-discord-bot/music/providers"
	"pxnx-discord-bot/services/ytdlp"
)

// YTDLPIntegrationExample demonstrates how to integrate yt-dlp service with the music system
func YTDLPIntegrationExample() {
	fmt.Println("🎵 yt-dlp Service Integration Example")
	fmt.Println("====================================")

	ctx := context.Background()

	// Example 1: Using yt-dlp provider directly
	fmt.Println("\n1. Direct yt-dlp Provider Usage:")
	directProviderExample(ctx)

	// Example 2: Using music manager with yt-dlp integration
	fmt.Println("\n2. Music Manager with yt-dlp Integration:")
	musicManagerExample(ctx)

	// Example 3: Service management and monitoring
	fmt.Println("\n3. Service Management and Monitoring:")
	serviceManagementExample(ctx)

	// Example 4: Error handling and resilience
	fmt.Println("\n4. Error Handling and Resilience:")
	resilienceExample(ctx)
}

// directProviderExample shows how to use the yt-dlp provider directly
func directProviderExample(ctx context.Context) {
	// Create a custom configuration
	config := ytdlp.DefaultServiceConfig()
	config.Port = 8081 // Use different port to avoid conflicts
	config.MaxWorkers = 1
	config.Format = "bestaudio[ext=webm][acodec=opus]/bestaudio/best"

	// Create yt-dlp provider
	provider := providers.NewYouTubeYTDLPProviderWithConfig(config)

	// Start the service
	if err := provider.Start(ctx); err != nil {
		log.Printf("Failed to start yt-dlp service: %v", err)
		return
	}
	defer func() {
		if err := provider.Stop(ctx); err != nil {
			log.Printf("Failed to stop yt-dlp service: %v", err)
		}
	}()

	// Test video extraction
	fmt.Println("  Extracting video info...")
	videoURL := "https://www.youtube.com/watch?v=dQw4w9WgXcQ" // Rick Roll for testing
	audioSource, err := provider.GetAudioSource(ctx, videoURL)
	if err != nil {
		log.Printf("  ❌ Failed to extract video: %v", err)
		return
	}

	fmt.Printf("  ✅ Title: %s\n", audioSource.Title)
	fmt.Printf("  ✅ Duration: %s\n", audioSource.Duration)
	fmt.Printf("  ✅ Provider: %s\n", audioSource.Provider)

	// Test search
	fmt.Println("  Searching for videos...")
	searchResults, err := provider.Search(ctx, "lofi hip hop", 3)
	if err != nil {
		log.Printf("  ❌ Search failed: %v", err)
		return
	}

	fmt.Printf("  ✅ Found %d results:\n", len(searchResults))
	for i, result := range searchResults {
		fmt.Printf("    %d. %s (%s)\n", i+1, result.Title, result.Duration)
	}
}

// musicManagerExample shows how to use yt-dlp with the music manager
func musicManagerExample(ctx context.Context) {
	// Note: This would normally use a real Discord session
	// For this example, we'll use a mock session (not shown here)

	// Create music manager (with mock session)
	// musicManager := manager.NewManager(mockSession)

	// For demonstration, we'll show the integration setup
	fmt.Println("  Setting up music manager with yt-dlp integration...")

	// This is how you would integrate yt-dlp with the music manager:
	/*
		// Create yt-dlp integration
		ytdlpIntegration := manager.NewYTDLPIntegration(musicManager)

		// Enable with custom config
		config := ytdlp.DefaultServiceConfig()
		config.MaxWorkers = 2
		config.Port = 8082

		if err := ytdlpIntegration.EnableWithConfig(ctx, config); err != nil {
			log.Printf("Failed to enable yt-dlp integration: %v", err)
			return
		}
		defer ytdlpIntegration.Disable(ctx)

		// Now you can use the music manager with yt-dlp support
		guildID := "123456789"
		channelID := "987654321"

		// Join voice channel
		if err := musicManager.JoinChannel(ctx, guildID, channelID); err != nil {
			log.Printf("Failed to join channel: %v", err)
			return
		}

		// Search and play
		audioSource, err := ytdlpIntegration.GetAudioSourceWithYTDLP(ctx, "relaxing music")
		if err != nil {
			log.Printf("Failed to get audio source: %v", err)
			return
		}

		if err := musicManager.Play(ctx, guildID, *audioSource); err != nil {
			log.Printf("Failed to play audio: %v", err)
			return
		}

		fmt.Printf("  ✅ Now playing: %s\n", audioSource.Title)
	*/

	fmt.Println("  ✅ Music manager integration setup complete")
}

// serviceManagementExample shows service management capabilities
func serviceManagementExample(ctx context.Context) {
	// Create service manager
	config := ytdlp.DefaultServiceConfig()
	config.Port = 8083
	serviceManager := ytdlp.NewServiceManager(config)

	// Start service
	fmt.Println("  Starting yt-dlp service...")
	if err := serviceManager.Start(ctx); err != nil {
		log.Printf("  ❌ Failed to start service: %v", err)
		return
	}
	defer serviceManager.Stop(ctx)

	fmt.Printf("  ✅ Service status: %s\n", serviceManager.GetStatus())

	// Check service health
	fmt.Println("  Checking service health...")
	client := serviceManager.GetClient()
	health, err := client.HealthCheck(ctx)
	if err != nil {
		log.Printf("  ❌ Health check failed: %v", err)
		return
	}

	fmt.Printf("  ✅ Service health: %s\n", health.Status)
	fmt.Printf("  ✅ Version: %s\n", health.Version)
	fmt.Printf("  ✅ Uptime: %s\n", health.Uptime)

	// Monitor service errors
	fmt.Println("  Monitoring service errors...")
	go func() {
		errorChan := serviceManager.GetErrors()
		select {
		case err := <-errorChan:
			fmt.Printf("  ⚠️  Service error detected: %v\n", err)
		case <-time.After(2 * time.Second):
			fmt.Println("  ✅ No errors detected in monitoring period")
		}
	}()

	time.Sleep(3 * time.Second)
}

// resilienceExample shows error handling and resilience features
func resilienceExample(ctx context.Context) {
	// Create resilient client
	config := ytdlp.DefaultServiceConfig()
	config.Port = 8084
	resilientClient := ytdlp.NewResilientClient(config)
	defer resilientClient.Close()

	// Note: For this example, we're not starting the actual service
	// This will demonstrate the retry and circuit breaker behavior

	fmt.Println("  Testing resilience features...")

	// This will fail because no service is running, demonstrating error handling
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := resilientClient.HealthCheck(ctx)
	if err != nil {
		fmt.Printf("  ✅ Expected error handled gracefully: %v\n", err)
	}

	// Check circuit breaker metrics
	metrics := resilientClient.GetCircuitBreakerMetrics()
	fmt.Printf("  ✅ Circuit breaker state: %v\n", metrics["state"])
	fmt.Printf("  ✅ Failure count: %v\n", metrics["failures"])

	fmt.Println("  ✅ Resilience testing complete")
}

// ProductionExample shows a complete production-ready setup
func ProductionExample() {
	fmt.Println("\n🏭 Production Setup Example")
	fmt.Println("===========================")

	ctx := context.Background()

	// Production configuration
	config := &ytdlp.ServiceConfig{
		Host:        "localhost",
		Port:        8080,
		MaxWorkers:  4,
		Timeout:     45 * time.Second,
		MaxRetries:  3,
		Format:      "bestaudio[ext=webm][acodec=opus]/bestaudio[ext=m4a]/bestaudio/best",
		AudioFormat: "opus",
		AudioQuality: "128K",
		RateLimit:   "1M",
		CacheDir:    "/var/cache/ytdlp",
		CacheTTL:    24 * time.Hour,
		HealthCheckInterval: 30 * time.Second,
	}

	// Circuit breaker configuration
	cbConfig := &ytdlp.CircuitBreakerConfig{
		FailureThreshold:      5,
		SuccessThreshold:      3,
		Timeout:              30 * time.Second,
		ResetTimeout:         60 * time.Second,
		MaxConcurrentRequests: 10,
	}

	// Retry configuration
	retryConfig := &ytdlp.RetryConfig{
		MaxRetries:    3,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      30 * time.Second,
		BackoffFactor: 2.0,
		RandomJitter:  true,
	}

	// Create resilient client with production configurations
	resilientClient := ytdlp.NewResilientClientWithConfigs(config, cbConfig, retryConfig)
	defer resilientClient.Close()

	fmt.Println("  ✅ Production configuration created")
	fmt.Printf("  ✅ Service endpoint: http://%s:%d\n", config.Host, config.Port)
	fmt.Printf("  ✅ Cache directory: %s\n", config.CacheDir)
	fmt.Printf("  ✅ Max workers: %d\n", config.MaxWorkers)
	fmt.Printf("  ✅ Preferred format: %s\n", config.Format)

	// Service utilities example
	utils := ytdlp.NewServiceUtils()

	// Check dependencies
	fmt.Println("  Checking dependencies...")
	if err := utils.CheckDependencies(); err != nil {
		fmt.Printf("  ⚠️  Dependency check: %v\n", err)
		fmt.Println("  💡 Run: pip install yt-dlp aiohttp")
	} else {
		fmt.Println("  ✅ All dependencies satisfied")
	}

	fmt.Println("\n  🎯 Production setup complete!")
	fmt.Println("     Ready for Discord music bot integration")
}