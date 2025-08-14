package engine

import (
	"testing"
	"time"

	"local-dev-tools/dynamic-request-scheduler/internal/spec"
)

func TestNewScheduler(t *testing.T) {
	requests := []spec.ScheduledRequest{
		{
			Name: "test-request",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("5m"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    "https://example.com",
			},
		},
	}

	config := SchedulerConfig{
		Workers:     2,
		Concurrency: 5,
		Once:        false,
		DryRun:      false,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)
	if scheduler == nil {
		t.Fatal("NewScheduler returned nil")
	}

	if len(scheduler.requests) != 1 {
		t.Errorf("Expected 1 request, got %d", len(scheduler.requests))
	}
	if scheduler.workers != 2 {
		t.Errorf("Expected 2 workers, got %d", scheduler.workers)
	}
	if scheduler.concurrency != 5 {
		t.Errorf("Expected concurrency 5, got %d", scheduler.concurrency)
	}
	if scheduler.once != false {
		t.Errorf("Expected once=false, got %v", scheduler.once)
	}
	if scheduler.dryRun != false {
		t.Errorf("Expected dryRun=false, got %v", scheduler.dryRun)
	}
}

func TestScheduler_DryRun(t *testing.T) {
	requests := []spec.ScheduledRequest{
		{
			Name: "test-request",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("5m"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    "https://example.com",
				Headers: map[string]string{
					"X-Test": "{{ uuid }}",
				},
			},
		},
	}

	config := SchedulerConfig{
		Workers:     1,
		Concurrency: 1,
		Once:        false,
		DryRun:      true,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)
	
	// Start the scheduler
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Wait a bit for dry-run to complete
	time.Sleep(100 * time.Millisecond)
	
	// Stop the scheduler
	scheduler.Stop()
}

func TestScheduler_Once(t *testing.T) {
	requests := []spec.ScheduledRequest{
		{
			Name: "test-request",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    "https://httpbin.org/get",
			},
		},
	}

	config := SchedulerConfig{
		Workers:     1,
		Concurrency: 1,
		Once:        true,
		DryRun:      false,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)
	
	// Start the scheduler
	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}
}

func TestScheduler_Stop(t *testing.T) {
	requests := []spec.ScheduledRequest{
		{
			Name: "test-request",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    "https://example.com",
			},
		},
	}

	config := SchedulerConfig{
		Workers:     1,
		Concurrency: 1,
		Once:        false,
		DryRun:      false,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)
	
	// Start the scheduler in a goroutine
	go func() {
		if err := scheduler.Start(); err != nil {
			t.Errorf("Start failed: %v", err)
		}
	}()

	// Wait a bit then stop
	time.Sleep(100 * time.Millisecond)
	scheduler.Stop()
}

func TestScheduler_ConcurrencyControl(t *testing.T) {
	// Create multiple requests that will run immediately
	requests := []spec.ScheduledRequest{
		{
			Name: "request-1",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    "https://httpbin.org/delay/1",
			},
		},
		{
			Name: "request-2",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    "https://httpbin.org/delay/1",
			},
		},
		{
			Name: "request-3",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    "https://httpbin.org/delay/1",
			},
		},
	}

	config := SchedulerConfig{
		Workers:     1,
		Concurrency: 2, // Only 2 concurrent requests allowed
		Once:        true,
		DryRun:      false,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)
	
	start := time.Now()
	err := scheduler.Start()
	duration := time.Since(start)
	
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// With concurrency=2 and 3 requests that each take 1 second,
	// the total time should be at least 2 seconds (2 batches)
	if duration < 2*time.Second {
		t.Errorf("Expected duration >= 2s due to concurrency limit, got %v", duration)
	}
}

func TestScheduler_ShouldRunRequest(t *testing.T) {
	scheduler := &Scheduler{}
	ctx := &spec.EvaluationContext{}
	templateEngine := spec.NewTemplateEngine(ctx)
	evaluator := spec.NewEvaluator(templateEngine)
	
	// Test relative schedule (should run immediately if in the past)
	relativeRequest := spec.ScheduledRequest{
		Schedule: spec.ScheduleSpec{
			Relative: stringPtr("1s"),
		},
	}
	
	// This should run immediately since it's a relative schedule
	if !scheduler.shouldRunRequest(&relativeRequest, evaluator) {
		t.Error("Relative request should run immediately")
	}

	// Test epoch schedule in the past
	pastRequest := spec.ScheduledRequest{
		Schedule: spec.ScheduleSpec{
			Epoch: int64Ptr(time.Now().Add(-1 * time.Hour).Unix()),
		},
	}
	
	if !scheduler.shouldRunRequest(&pastRequest, evaluator) {
		t.Error("Past epoch request should run")
	}

	// Test epoch schedule in the future
	futureRequest := spec.ScheduledRequest{
		Schedule: spec.ScheduleSpec{
			Epoch: int64Ptr(time.Now().Add(1 * time.Hour).Unix()),
		},
	}
	
	if scheduler.shouldRunRequest(&futureRequest, evaluator) {
		t.Error("Future epoch request should not run yet")
	}
}

func TestScheduler_ExecuteRequest(t *testing.T) {
	requests := []spec.ScheduledRequest{
		{
			Name: "test-request",
			Schedule: spec.ScheduleSpec{
				Relative: stringPtr("1s"),
			},
			HTTP: spec.HttpRequestSpec{
				Method: "GET",
				URL:    "https://httpbin.org/get",
				Headers: map[string]string{
					"X-Test": "{{ uuid }}",
				},
			},
		},
	}

	config := SchedulerConfig{
		Workers:     1,
		Concurrency: 1,
		Once:        true,
		DryRun:      false,
		Timeout:     30 * time.Second,
	}

	scheduler := NewScheduler(requests, config)
	
	// Test executeRequest directly
	ctx := &spec.EvaluationContext{
		Clock:     &spec.RealClock{},
		Variables: make(map[string]interface{}),
	}
	templateEngine := spec.NewTemplateEngine(ctx)
	evaluator := spec.NewEvaluator(templateEngine)
	scheduler.executeRequest(&requests[0], evaluator)
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
