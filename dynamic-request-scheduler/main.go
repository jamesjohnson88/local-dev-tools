package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"local-dev-tools/dynamic-request-scheduler/internal/engine"
	"local-dev-tools/dynamic-request-scheduler/internal/spec"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file (YAML or JSON)")
	intervalSeconds := flag.Int("interval", 60, "Request interval in seconds (legacy mode)")
	dryRun := flag.Bool("dry-run", false, "Show resolved requests without sending")
	once := flag.Bool("once", false, "Run all requests once and exit")
	workers := flag.Int("workers", 1, "Number of worker goroutines")
	concurrency := flag.Int("concurrency", 10, "Maximum concurrent requests")
	timeout := flag.Duration("timeout", 30*time.Second, "HTTP request timeout")
	flag.Parse()

	if *configPath == "" {
		// Legacy mode - run with hardcoded request every interval
		fmt.Printf("No config file specified, running in legacy mode with interval of %ds\n", *intervalSeconds)
		runLegacyMode(*intervalSeconds)
		return
	}

	// Load configuration
	requests, err := spec.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	fmt.Printf("Loaded %d requests from %s\n", len(requests), *configPath)

	// Create scheduler configuration
	config := engine.SchedulerConfig{
		Workers:     *workers,
		Concurrency: *concurrency,
		Once:        *once,
		DryRun:      *dryRun,
		Timeout:     *timeout,
	}

	// Create and start scheduler
	scheduler := engine.NewScheduler(requests, config)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, stopping scheduler...")
		scheduler.Stop()
	}()

	// Start the scheduler
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Scheduler error: %v", err)
	}
}

func runLegacyMode(intervalSeconds int) {
	// Create a legacy request for backward compatibility
	legacyRequest := &spec.ScheduledRequest{
		Name: "Legacy Run Once",
		Schedule: spec.ScheduleSpec{
			Relative: stringPtr("10m"),
		},
		HTTP: spec.HttpRequestSpec{
			Method: "POST",
			URL:    "https://localhost:10001/core/scheduler/tasks/run-once",
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
			Body: map[string]interface{}{
				"scheduled_for":        time.Now().Unix() + 600,
				"task_request_method":  "GET",
				"task_request_url":     "https://localhost:10001/fad/health",
				"task_request_headers": nil,
				"task_request_payload": nil,
			},
		},
	}

	// Create scheduler for legacy mode
	config := engine.SchedulerConfig{
		Workers:     1,
		Concurrency: 1,
		Once:        false,
		DryRun:      false,
		Timeout:     30 * time.Second,
	}

	scheduler := engine.NewScheduler([]spec.ScheduledRequest{*legacyRequest}, config)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nReceived shutdown signal, stopping legacy scheduler...")
		scheduler.Stop()
	}()

	// Start the scheduler
	if err := scheduler.Start(); err != nil {
		log.Fatalf("Legacy scheduler error: %v", err)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
