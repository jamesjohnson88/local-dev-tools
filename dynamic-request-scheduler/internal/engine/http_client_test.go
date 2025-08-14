package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"local-dev-tools/dynamic-request-scheduler/internal/spec"
)

func TestNewHTTPClient(t *testing.T) {
	client := NewHTTPClient(30 * time.Second)
	if client == nil {
		t.Fatal("NewHTTPClient returned nil")
	}
	if client.timeout != 30*time.Second {
		t.Errorf("Expected timeout 30s, got %v", client.timeout)
	}
	if client.client.Timeout != 30*time.Second {
		t.Errorf("Expected http.Client timeout 30s, got %v", client.client.Timeout)
	}
}

func TestHTTPClient_SendRequest_GET(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/test" {
			t.Errorf("Expected path /test, got %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(30 * time.Second)
	resolved := &spec.ResolvedRequest{
		Method:  "GET",
		URL:     server.URL + "/test",
		Headers: map[string]string{"X-Test": "value"},
		Body:    nil,
	}

	resp, err := client.SendRequest(resolved)
	if err != nil {
		t.Fatalf("SendRequest failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if resp.Duration <= 0 {
		t.Errorf("Expected positive duration, got %v", resp.Duration)
	}
	if resp.ContentLength != 16 {
		t.Errorf("Expected content length 16, got %d", resp.ContentLength)
	}
}

func TestHTTPClient_SendRequest_POST(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		
		// Read and verify body
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("Failed to decode body: %v", err)
		}
		if body["test"] != "value" {
			t.Errorf("Expected body.test=value, got %v", body["test"])
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": "123"}`))
	}))
	defer server.Close()

	client := NewHTTPClient(30 * time.Second)
	resolved := &spec.ResolvedRequest{
		Method: "POST",
		URL:    server.URL + "/test",
		Headers: map[string]string{
			"X-Test": "header-value",
		},
		Body: map[string]interface{}{
			"test": "value",
		},
	}

	resp, err := client.SendRequest(resolved)
	if err != nil {
		t.Fatalf("SendRequest failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_SendRequest_WithCustomContentType(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/xml" {
			t.Errorf("Expected Content-Type application/xml, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(30 * time.Second)
	resolved := &spec.ResolvedRequest{
		Method: "POST",
		URL:    server.URL + "/test",
		Headers: map[string]string{
			"Content-Type": "application/xml",
		},
		Body: "<test>value</test>",
	}

	resp, err := client.SendRequest(resolved)
	if err != nil {
		t.Fatalf("SendRequest failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestHTTPClient_SendRequest_Timeout(t *testing.T) {
	// Create a slow test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Longer than our timeout
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewHTTPClient(100 * time.Millisecond) // Short timeout
	resolved := &spec.ResolvedRequest{
		Method: "GET",
		URL:    server.URL + "/test",
		Headers: map[string]string{},
		Body:    nil,
	}

	_, err := client.SendRequest(resolved)
	if err == nil {
		t.Fatal("Expected timeout error, got nil")
	}
}

func TestHTTPClient_SendRequest_InvalidURL(t *testing.T) {
	client := NewHTTPClient(30 * time.Second)
	resolved := &spec.ResolvedRequest{
		Method:  "GET",
		URL:     "http://invalid-url-that-does-not-exist.localhost:99999",
		Headers: map[string]string{},
		Body:    nil,
	}

	_, err := client.SendRequest(resolved)
	if err == nil {
		t.Fatal("Expected error for invalid URL, got nil")
	}
}

func TestHTTPResponse_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		expected bool
	}{
		{"200 OK", 200, true},
		{"201 Created", 201, true},
		{"204 No Content", 204, true},
		{"299 Custom", 299, true},
		{"400 Bad Request", 400, false},
		{"404 Not Found", 404, false},
		{"500 Internal Server Error", 500, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &HTTPResponse{StatusCode: tt.status}
			if resp.IsSuccess() != tt.expected {
				t.Errorf("IsSuccess() for status %d = %v, want %v", tt.status, resp.IsSuccess(), tt.expected)
			}
		})
	}
}

func TestHTTPResponse_String(t *testing.T) {
	resp := &HTTPResponse{
		StatusCode:    200,
		Status:        "200 OK",
		Duration:      150 * time.Millisecond,
		ContentLength: 100,
	}

	str := resp.String()
	expected := "HTTP 200 200 OK (150ms, 100 bytes)"
	if str != expected {
		t.Errorf("String() = %s, want %s", str, expected)
	}
}
