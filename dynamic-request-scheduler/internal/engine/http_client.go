package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"local-dev-tools/dynamic-request-scheduler/internal/spec"
)

// HTTPClient handles HTTP request execution
type HTTPClient struct {
	client  *http.Client
	timeout time.Duration
}

// NewHTTPClient creates a new HTTP client
func NewHTTPClient(timeout time.Duration) *HTTPClient {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &HTTPClient{
		client: &http.Client{
			Timeout: timeout,
		},
		timeout: timeout,
	}
}

// SendRequest sends an HTTP request and returns the response details
func (c *HTTPClient) SendRequest(resolved *spec.ResolvedRequest) (*HTTPResponse, error) {
	start := time.Now()

	// Prepare request body
	var body io.Reader
	if resolved.Body != nil && resolved.Method != "GET" && resolved.Method != "HEAD" {
		jsonData, err := json.Marshal(resolved.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		body = bytes.NewReader(jsonData)
	}

	// Create HTTP request
	req, err := http.NewRequest(resolved.Method, resolved.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	for key, value := range resolved.Headers {
		req.Header.Set(key, value)
	}

	// Set default Content-Type for requests with body
	if body != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Send request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	duration := time.Since(start)

	return &HTTPResponse{
		StatusCode:    resp.StatusCode,
		Status:        resp.Status,
		Headers:       resp.Header,
		Body:          responseBody,
		Duration:      duration,
		ContentLength: len(responseBody),
	}, nil
}

// HTTPResponse represents an HTTP response
type HTTPResponse struct {
	StatusCode    int
	Status        string
	Headers       http.Header
	Body          []byte
	Duration      time.Duration
	ContentLength int
}

// IsSuccess returns true if the response indicates success
func (r *HTTPResponse) IsSuccess() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

// String returns a string representation of the response
func (r *HTTPResponse) String() string {
	return fmt.Sprintf("HTTP %d %s (%v, %d bytes)",
		r.StatusCode, r.Status, r.Duration, r.ContentLength)
}
