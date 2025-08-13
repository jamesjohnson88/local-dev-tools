package spec

import (
	"strings"
	"testing"
	"time"
)

func TestNewEvaluator(t *testing.T) {
	ctx := &EvaluationContext{
		Variables: map[string]interface{}{"test": "value"},
	}
	engine := NewTemplateEngine(ctx)
	evaluator := NewEvaluator(engine)

	if evaluator == nil {
		t.Fatal("NewEvaluator() returned nil")
	}

	if evaluator.engine != engine {
		t.Error("Evaluator engine not set correctly")
	}
}

func TestEvaluator_EvaluateRequest(t *testing.T) {
	fixedTime := time.Unix(1000, 0)
	ctx := &EvaluationContext{
		Variables: map[string]interface{}{"api_key": "secret123"},
		Clock:     &MockClock{now: fixedTime},
	}
	engine := NewTemplateEngine(ctx)
	evaluator := NewEvaluator(engine)

	tests := []struct {
		name    string
		request *ScheduledRequest
		want    *ResolvedRequest
		wantErr bool
	}{
		{
			name: "simple request without templates",
			request: &ScheduledRequest{
				Name: "Test Request",
				Schedule: ScheduleSpec{
					Relative: stringPtr("5m"),
				},
				HTTP: HttpRequestSpec{
					Method: "GET",
					URL:    "https://api.example.com/health",
					Headers: map[string]string{
						"User-Agent": "TestClient",
					},
					Body: nil,
				},
			},
			want: &ResolvedRequest{
				Name:   "Test Request",
				Method: "GET",
				URL:    "https://api.example.com/health",
				Headers: map[string]string{
					"User-Agent": "TestClient",
				},
				Body:         nil,
				ScheduledFor: fixedTime.Add(5 * time.Minute),
			},
			wantErr: false,
		},
		{
			name: "request with URL template",
			request: &ScheduledRequest{
				Name: "Dynamic URL",
				Schedule: ScheduleSpec{
					Relative: stringPtr("1m"),
				},
				HTTP: HttpRequestSpec{
					Method: "POST",
					URL:    "https://api.example.com/users/{{ uuid }}",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"id": "{{ uuid }}",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "request with header templates",
			request: &ScheduledRequest{
				Name: "Dynamic Headers",
				Schedule: ScheduleSpec{
					Relative: stringPtr("2m"),
				},
				HTTP: HttpRequestSpec{
					Method: "GET",
					URL:    "https://api.example.com/data",
					Headers: map[string]string{
						"X-Trace-ID":    "{{ uuid }}",
						"X-Timestamp":   "{{ now | unix }}",
						"Authorization": "Bearer {{ .Variables.api_key }}",
					},
					Body: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "request with body templates",
			request: &ScheduledRequest{
				Name: "Dynamic Body",
				Schedule: ScheduleSpec{
					Relative: stringPtr("3m"),
				},
				HTTP: HttpRequestSpec{
					Method: "POST",
					URL:    "https://api.example.com/events",
					Headers: map[string]string{
						"Content-Type": "application/json",
					},
					Body: map[string]interface{}{
						"event_id":  "{{ uuid }}",
						"timestamp": "{{ now | rfc3339 }}",
						"sequence":  "{{ seq }}",
						"metadata": map[string]interface{}{
							"source":  "test",
							"version": "{{ seq }}",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "request with epoch schedule",
			request: &ScheduledRequest{
				Name: "Epoch Schedule",
				Schedule: ScheduleSpec{
					Epoch: int64Ptr(2000),
				},
				HTTP: HttpRequestSpec{
					Method:  "GET",
					URL:     "https://api.example.com/health",
					Headers: map[string]string{},
					Body:    nil,
				},
			},
			want: &ResolvedRequest{
				Name:         "Epoch Schedule",
				Method:       "GET",
				URL:          "https://api.example.com/health",
				Headers:      map[string]string{},
				Body:         nil,
				ScheduledFor: time.Unix(2000, 0),
			},
			wantErr: false,
		},
		{
			name: "request with template schedule",
			request: &ScheduledRequest{
				Name: "Template Schedule",
				Schedule: ScheduleSpec{
					Template: stringPtr("{{ addMinutes 10 now | unix }}"),
				},
				HTTP: HttpRequestSpec{
					Method:  "GET",
					URL:     "https://api.example.com/health",
					Headers: map[string]string{},
					Body:    nil,
				},
			},
			wantErr: false,
		},
		{
			name:    "nil request",
			request: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.EvaluateRequest(tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if result == nil {
					t.Fatal("EvaluateRequest() returned nil result")
				}

				// Check basic fields
				if result.Name != tt.request.Name {
					t.Errorf("Name = %v, want %v", result.Name, tt.request.Name)
				}
				if result.Method != tt.request.HTTP.Method {
					t.Errorf("Method = %v, want %v", result.Method, tt.request.HTTP.Method)
				}

				// Check URL resolution
				if tt.want != nil && tt.want.URL != "" {
					if result.URL != tt.want.URL {
						t.Errorf("URL = %v, want %v", result.URL, tt.want.URL)
					}
				}

				// Check headers resolution
				if tt.want != nil && tt.want.Headers != nil {
					for key, expectedValue := range tt.want.Headers {
						if actualValue, exists := result.Headers[key]; !exists {
							t.Errorf("Header %s not found", key)
						} else if actualValue != expectedValue {
							t.Errorf("Header %s = %v, want %v", key, actualValue, expectedValue)
						}
					}
				}

				// Check scheduled time
				if tt.want != nil && !tt.want.ScheduledFor.IsZero() {
					if !result.ScheduledFor.Equal(tt.want.ScheduledFor) {
						t.Errorf("ScheduledFor = %v, want %v", result.ScheduledFor, tt.want.ScheduledFor)
					}
				}

				// For requests with templates, verify they were resolved
				if tt.name == "request with URL template" {
					if result.URL == "https://api.example.com/users/{{ uuid }}" {
						t.Error("URL template was not resolved")
					}
					if !strings.Contains(result.URL, "https://api.example.com/users/") {
						t.Errorf("URL resolution failed: %s", result.URL)
					}
				}

				if tt.name == "request with header templates" {
					if result.Headers["X-Trace-ID"] == "{{ uuid }}" {
						t.Error("Header template was not resolved")
					}
					if result.Headers["X-Timestamp"] == "{{ now | unix }}" {
						t.Error("Header template was not resolved")
					}
					if result.Headers["Authorization"] != "Bearer secret123" {
						t.Errorf("Variable substitution failed: %s", result.Headers["Authorization"])
					}
				}

				if tt.name == "request with body templates" {
					body, ok := result.Body.(map[string]interface{})
					if !ok {
						t.Fatal("Body is not a map")
					}
					if body["event_id"] == "{{ uuid }}" {
						t.Error("Body template was not resolved")
					}
					if body["timestamp"] == "{{ now | rfc3339 }}" {
						t.Error("Body template was not resolved")
					}
					if body["sequence"] == "{{ seq }}" {
						t.Error("Body template was not resolved")
					}
				}
			}
		})
	}
}

func TestEvaluator_ResolveValue(t *testing.T) {
	ctx := &EvaluationContext{
		Variables: map[string]interface{}{"test_var": "test_value"},
		Clock:     &MockClock{now: time.Unix(1000, 0)},
	}
	engine := NewTemplateEngine(ctx)
	evaluator := NewEvaluator(engine)

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name:  "nil value",
			input: nil,
		},
		{
			name:  "string without template",
			input: "hello world",
		},
		{
			name:  "string with template",
			input: "{{ now | unix }}",
		},
		{
			name: "map with templates",
			input: map[string]interface{}{
				"key1": "{{ uuid }}",
				"key2": "static value",
				"nested": map[string]interface{}{
					"deep": "{{ now | rfc3339 }}",
				},
			},
		},
		{
			name: "slice with templates",
			input: []interface{}{
				"{{ seq }}",
				"static",
				"{{ env \"PATH\" }}",
			},
		},
		{
			name: "DynamicString with template",
			input: DynamicString{
				template:   "{{ now | unix }}",
				isTemplate: true,
			},
		},
		{
			name: "DynamicString without template",
			input: DynamicString{
				value:      "static",
				isTemplate: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.resolveValue(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("resolveValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != nil {
				// Verify that templates were resolved
				if tt.name == "string with template" {
					if result == "{{ now | unix }}" {
						t.Error("Template was not resolved")
					}
					if result != "1000" {
						t.Errorf("Template resolution failed: %v", result)
					}
				}

				if tt.name == "map with templates" {
					resultMap, ok := result.(map[string]interface{})
					if !ok {
						t.Fatal("Result is not a map")
					}
					if resultMap["key1"] == "{{ uuid }}" {
						t.Error("Map key1 template was not resolved")
					}
					if resultMap["key2"] != "static value" {
						t.Error("Static map value was changed")
					}
					nested, ok := resultMap["nested"].(map[string]interface{})
					if !ok {
						t.Fatal("Nested map is not a map")
					}
					if nested["deep"] == "{{ now | rfc3339 }}" {
						t.Error("Nested template was not resolved")
					}
				}

				if tt.name == "DynamicString with template" {
					if result == "{{ now | unix }}" {
						t.Error("DynamicString template was not resolved")
					}
					if result != "1000" {
						t.Errorf("DynamicString template resolution failed: %v", result)
					}
				}

				if tt.name == "DynamicString without template" {
					if result != "static" {
						t.Errorf("DynamicString value was changed: %v", result)
					}
				}
			}
		})
	}
}

func TestEvaluator_ComputeScheduledTime(t *testing.T) {
	fixedTime := time.Unix(1000, 0)
	ctx := &EvaluationContext{
		Clock: &MockClock{now: fixedTime},
	}
	engine := NewTemplateEngine(ctx)
	evaluator := NewEvaluator(engine)

	tests := []struct {
		name     string
		schedule ScheduleSpec
		want     time.Time
		wantErr  bool
	}{
		{
			name: "epoch schedule",
			schedule: ScheduleSpec{
				Epoch: int64Ptr(2000),
			},
			want: time.Unix(2000, 0),
		},
		{
			name: "relative schedule",
			schedule: ScheduleSpec{
				Relative: stringPtr("5m"),
			},
			want: fixedTime.Add(5 * time.Minute),
		},
		{
			name: "relative schedule with jitter",
			schedule: ScheduleSpec{
				Relative: stringPtr("10m"),
				Jitter:   stringPtr("±30s"),
			},
			wantErr: false, // We'll check it's within range
		},
		{
			name: "template schedule",
			schedule: ScheduleSpec{
				Template: stringPtr("{{ addMinutes 15 now | unix }}"),
			},
			want: fixedTime.Add(15 * time.Minute),
		},
		{
			name: "cron schedule",
			schedule: ScheduleSpec{
				Cron: stringPtr("*/5 * * * *"),
			},
			wantErr: true, // Not implemented yet
		},
		{
			name:     "no schedule",
			schedule: ScheduleSpec{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := evaluator.computeScheduledTime(tt.schedule)

			if (err != nil) != tt.wantErr {
				t.Errorf("computeScheduledTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.name == "relative schedule with jitter" {
					// Check that jitter is within expected range
					baseTime := fixedTime.Add(10 * time.Minute)
					jitterRange := 30 * time.Second
					if result.Before(baseTime) || result.After(baseTime.Add(jitterRange)) {
						t.Errorf("Jittered time %v is outside expected range [%v, %v]",
							result, baseTime, baseTime.Add(jitterRange))
					}
				} else if !result.Equal(tt.want) {
					t.Errorf("computeScheduledTime() = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestEvaluator_SetVariable(t *testing.T) {
	ctx := &EvaluationContext{
		Variables: make(map[string]interface{}),
	}
	engine := NewTemplateEngine(ctx)
	evaluator := NewEvaluator(engine)

	// Test setting variables
	evaluator.SetVariable("key1", "value1")
	evaluator.SetVariable("key2", 42)

	if ctx.Variables["key1"] != "value1" {
		t.Errorf("Variable key1 not set correctly")
	}
	if ctx.Variables["key2"] != 42 {
		t.Errorf("Variable key2 not set correctly")
	}
}

func TestEvaluator_SetSeed(t *testing.T) {
	ctx := &EvaluationContext{}
	engine := NewTemplateEngine(ctx)
	evaluator := NewEvaluator(engine)

	// Test seed setting
	evaluator.SetSeed(12345)
	if ctx.Seed != 12345 {
		t.Errorf("Seed not set correctly: %v", ctx.Seed)
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func int64Ptr(i int64) *int64 {
	return &i
}
