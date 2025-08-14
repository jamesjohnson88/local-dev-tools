package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"local-dev-tools/dynamic-request-scheduler/internal/spec"
)

// MockServer represents a test HTTP server with request tracking
type MockServer struct {
	server     *httptest.Server
	requests   []MockRequest
	mu         sync.RWMutex
	statusCode int
	response   interface{}
}

// MockRequest represents a request received by the mock server
type MockRequest struct {
	Method  string
	Path    string
	Headers map[string]string
	Body    interface{}
	Time    time.Time
}

// NewMockServer creates a new mock HTTP server
func NewMockServer(statusCode int, response interface{}) *MockServer {
	ms := &MockServer{
		statusCode: statusCode,
		response:   response,
	}

	ms.server = httptest.NewServer(http.HandlerFunc(ms.handler))
	return ms
}

// handler handles incoming requests to the mock server
func (ms *MockServer) handler(w http.ResponseWriter, r *http.Request) {
	// Read body
	var body interface{}
	if r.Body != nil {
		json.NewDecoder(r.Body).Decode(&body)
	}

	// Convert headers to map[string]string
	headers := make(map[string]string)
	for key, values := range r.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	// Record request
	ms.mu.Lock()
	ms.requests = append(ms.requests, MockRequest{
		Method:  r.Method,
		Path:    r.URL.Path,
		Headers: headers,
		Body:    body,
		Time:    time.Now(),
	})
	ms.mu.Unlock()

	// Set response headers
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(ms.statusCode)

	// Write response
	if ms.response != nil {
		json.NewEncoder(w).Encode(ms.response)
	}
}

// GetRequests returns all requests received by the server
func (ms *MockServer) GetRequests() []MockRequest {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	requests := make([]MockRequest, len(ms.requests))
	copy(requests, ms.requests)
	return requests
}

// Close closes the mock server
func (ms *MockServer) Close() {
	ms.server.Close()
}

// URL returns the server's URL
func (ms *MockServer) URL() string {
	return ms.server.URL
}

func TestIntegration_BasicRequestFlow(t *testing.T) {
	// Create mock server
	mockServer := NewMockServer(http.StatusOK, map[string]string{"status": "ok"})
	defer mockServer.Close()

	// Create test requests
	requests := []spec.ScheduledRequest{
		{
			Name: "test-get",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    mockServer.URL() + "/test",
				Headers: map[string]string{
					"X-Test": "{{ uuid }}",
				},
			},
		},
		{
			Name: "test-post",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "POST",
				URL:    mockServer.URL() + "/test",
				Headers: map[string]string{
					"Content-Type": "application/json",
				},
				Body: map[string]interface{}{
					"message": "{{ uuid }}",
					"time":    "{{ now | rfc3339 }}",
				},
			},
		},
	}

	// Create scheduler
	config := SchedulerConfig{
		Workers:     1,
		Concurrency: 2,
		Once:        true,
		DryRun:      false,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)

	// Start scheduler
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Scheduler start failed: %v", err)
	}

	// Wait a bit for requests to complete
	time.Sleep(2 * time.Second)

	// Check that requests were received
	receivedRequests := mockServer.GetRequests()
	if len(receivedRequests) != 2 {
		t.Errorf("Expected 2 requests, got %d", len(receivedRequests))
	}

	// Find GET and POST requests (order may vary due to concurrency)
	var getRequest, postRequest *MockRequest
	for i := range receivedRequests {
		if receivedRequests[i].Method == "GET" {
			getRequest = &receivedRequests[i]
		} else if receivedRequests[i].Method == "POST" {
			postRequest = &receivedRequests[i]
		}
	}

	// Verify GET request
	if getRequest == nil {
		t.Error("GET request not found")
	} else {
		if getRequest.Path != "/test" {
			t.Errorf("Expected path /test, got %s", getRequest.Path)
		}
	}

	// Verify POST request
	if postRequest == nil {
		t.Error("POST request not found")
	} else {
		if postRequest.Path != "/test" {
			t.Errorf("Expected path /test, got %s", postRequest.Path)
		}
		if postRequest.Headers["Content-Type"] != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", postRequest.Headers["Content-Type"])
		}
	}
}

func TestIntegration_ConcurrencyControl(t *testing.T) {
	// Create mock server that delays responses
	mockServer := NewMockServer(http.StatusOK, map[string]string{"status": "ok"})
	defer mockServer.Close()

	// Override the handler to add delay
	mockServer.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add 1 second delay
		time.Sleep(1 * time.Second)
		mockServer.handler(w, r)
	})

	// Create multiple requests that will run concurrently
	requests := []spec.ScheduledRequest{
		{
			Name: "request-1",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    mockServer.URL() + "/delay/1",
			},
		},
		{
			Name: "request-2",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    mockServer.URL() + "/delay/1",
			},
		},
		{
			Name: "request-3",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    mockServer.URL() + "/delay/1",
			},
		},
	}

	// Create scheduler with limited concurrency
	config := SchedulerConfig{
		Workers:     1,
		Concurrency: 2, // Only 2 concurrent requests allowed
		Once:        true,
		DryRun:      false,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)

	// Measure execution time
	start := time.Now()
	err := scheduler.Start()
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Scheduler start failed: %v", err)
	}

	// With concurrency=2 and 3 requests that each take 1 second,
	// the total time should be at least 2 seconds (2 batches)
	if duration < 2*time.Second {
		t.Errorf("Expected duration >= 2s due to concurrency limit, got %v", duration)
	}

	// Check that all requests were received
	receivedRequests := mockServer.GetRequests()
	if len(receivedRequests) != 3 {
		t.Errorf("Expected 3 requests, got %d", len(receivedRequests))
	}
}

func TestIntegration_DryRunMode(t *testing.T) {
	// Create mock server
	mockServer := NewMockServer(http.StatusOK, map[string]string{"status": "ok"})
	defer mockServer.Close()

	// Create test request
	requests := []spec.ScheduledRequest{
		{
			Name: "test-dry-run",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("5m"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "POST",
				URL:    mockServer.URL() + "/test",
				Headers: map[string]string{
					"X-Test": "{{ uuid }}",
				},
				Body: map[string]interface{}{
					"message": "{{ uuid }}",
				},
			},
		},
	}

	// Create scheduler in dry-run mode
	config := SchedulerConfig{
		Workers:     1,
		Concurrency: 1,
		Once:        false,
		DryRun:      true,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)

	// Start scheduler
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Scheduler start failed: %v", err)
	}

	// Wait a bit for dry-run to complete
	time.Sleep(100 * time.Millisecond)

	// In dry-run mode, no actual requests should be sent
	receivedRequests := mockServer.GetRequests()
	if len(receivedRequests) != 0 {
		t.Errorf("Expected 0 requests in dry-run mode, got %d", len(receivedRequests))
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	// Create mock server that returns errors
	mockServer := NewMockServer(http.StatusInternalServerError, map[string]string{"error": "server error"})
	defer mockServer.Close()

	// Create test request
	requests := []spec.ScheduledRequest{
		{
			Name: "test-error",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    mockServer.URL() + "/error",
			},
		},
	}

	// Create scheduler
	config := SchedulerConfig{
		Workers:     1,
		Concurrency: 1,
		Once:        true,
		DryRun:      false,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)

	// Start scheduler
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Scheduler start failed: %v", err)
	}

	// Wait a bit for request to complete
	time.Sleep(2 * time.Second)

	// Check that request was received (even though it failed)
	receivedRequests := mockServer.GetRequests()
	if len(receivedRequests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(receivedRequests))
	}
}


