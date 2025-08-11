package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ScheduledRequest struct {
	displayName string
	method      string
	url         string
	contentType string
	requestBody interface{}
}

func NewScheduledRequest() *ScheduledRequest {
	return &ScheduledRequest{
		displayName: "Send 'run once' request to core-scheduler",
		method:      "POST",
		url:         "https://localhost:10001/core/scheduler/tasks/run-once",
		contentType: "application/json",
		requestBody: NewRequestBody(),
	}
}

type RequestBody struct {
	ScheduledFor       int64             `json:"scheduled_for"`
	TaskRequestMethod  string            `json:"task_request_method"`
	TaskRequestUrl     string            `json:"task_request_url"`
	TaskRequestHeaders map[string]string `json:"task_request_headers"`
	TaskRequestPayload interface{}       `json:"task_request_payload"`
}

func NewRequestBody() *RequestBody {
	return &RequestBody{
		ScheduledFor:       time.Now().Unix() + 600,
		TaskRequestMethod:  "GET",
		TaskRequestUrl:     "https://localhost:10001/fad/health",
		TaskRequestHeaders: nil,
		TaskRequestPayload: nil,
	}
}

func sendEvent() {
	s := NewScheduledRequest()

	var body io.Reader
	if s.method != "GET" && s.requestBody != nil {
		jsonData, err := json.Marshal(s.requestBody)
		if err != nil {
			fmt.Printf("Error marshaling JSON: %v\n", err)
			return
		}
		body = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(s.method, s.url, body)
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	if body != nil {
		req.Header.Set("Content-Type", s.contentType)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("%v %v: %v\n", s.method, s.url, resp.Status)
}

func main() {
	intervalSeconds := flag.Int("interval", 60, "Request interval in seconds")
	flag.Parse()

	fmt.Printf("Running with interval of %ds \n", *intervalSeconds)

	interval := time.Duration(*intervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		sendEvent()
	}
}
