package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	"local-dev-tools/dynamic-request-scheduler/internal/spec"
)

func sendEvent(requests []spec.ScheduledRequest) {
	for _, req := range requests {
		// TODO: Implement dynamic field resolution and scheduling logic
		// For now, just send the request as-is
		sendSingleRequest(&req)
	}
}

func sendSingleRequest(req *spec.ScheduledRequest) {
	var body io.Reader
	if req.HTTP.Method != "GET" && req.HTTP.Body != nil {
		jsonData, err := json.Marshal(req.HTTP.Body)
		if err != nil {
			fmt.Printf("Error marshaling JSON for request '%s': %v\n", req.Name, err)
			return
		}
		body = bytes.NewReader(jsonData)
	}

	httpReq, err := http.NewRequest(req.HTTP.Method, req.HTTP.URL, body)
	if err != nil {
		fmt.Printf("Error creating request for '%s': %v\n", req.Name, err)
		return
	}

	// Set headers
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Set custom headers
	for key, value := range req.HTTP.Headers {
		httpReq.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("Error sending request '%s': %v\n", req.Name, err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("[%s] %s %s: %s\n", req.Name, req.HTTP.Method, req.HTTP.URL, resp.Status)
}

func main() {
	configPath := flag.String("config", "", "Path to configuration file (YAML or JSON)")
	intervalSeconds := flag.Int("interval", 60, "Request interval in seconds (legacy mode)")
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
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	fmt.Printf("Loaded %d requests from %s\n", len(requests), *configPath)

	// TODO: Implement proper scheduling engine
	// For now, just send all requests immediately
	sendEvent(requests)
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

	interval := time.Duration(intervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		sendSingleRequest(legacyRequest)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
