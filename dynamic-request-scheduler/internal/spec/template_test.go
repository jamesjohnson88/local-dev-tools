package spec

import (
	"strings"
	"testing"
	"time"
)

// MockClock implements Clock for testing
type MockClock struct {
	now time.Time
}

func (m *MockClock) Now() time.Time { return m.now }

func TestNewTemplateEngine(t *testing.T) {
	ctx := &EvaluationContext{
		Variables: map[string]interface{}{"test": "value"},
		Sequence:  42,
		Seed:      123,
		Clock:     &MockClock{now: time.Unix(1000, 0)},
	}

	engine := NewTemplateEngine(ctx)
	if engine == nil {
		t.Fatal("NewTemplateEngine() returned nil")
	}

	if engine.ctx != ctx {
		t.Error("Template engine context not set correctly")
	}
}

func TestTemplateEngine_EvaluateTemplate(t *testing.T) {
	fixedTime := time.Unix(1000, 0)
	ctx := &EvaluationContext{
		Variables: map[string]interface{}{"name": "test"},
		Clock:     &MockClock{now: fixedTime},
	}
	engine := NewTemplateEngine(ctx)

	tests := []struct {
		name    string
		tmpl    string
		want    string
		wantErr bool
	}{
		{
			name: "simple text",
			tmpl: "hello world",
			want: "hello world",
		},
		{
			name: "variable substitution",
			tmpl: "hello {{ .Variables.name }}",
			want: "hello test",
		},
		{
			name: "time function",
			tmpl: "{{ now | unix }}",
			want: "1000",
		},
		{
			name: "time formatting",
			tmpl: "{{ now | rfc3339 }}",
			want: fixedTime.Format(time.RFC3339),
		},
		{
			name: "time arithmetic",
			tmpl: "{{ addMinutes 5 now | unix }}",
			want: "1300", // 1000 + 5*60
		},
		{
			name: "multiple functions",
			tmpl: "{{ addMinutes 30 (addHours 1 now) | unix }}",
			want: "6400", // 1000 + 1*3600 + 30*60
		},
		{
			name: "uuid generation",
			tmpl: "{{ uuid }}",
			want: "", // We'll check it's not empty
		},
		{
			name: "sequence counter",
			tmpl: "{{ seq }}",
			want: "1",
		},
		{
			name:    "invalid template",
			tmpl:    "{{ invalid function }}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.EvaluateTemplate(tt.tmpl)

			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if tt.name == "uuid generation" {
					if result == "" {
						t.Error("UUID generation should not return empty string")
					}
					// Check UUID format (8-4-4-4-12 hex digits)
					parts := strings.Split(result, "-")
					if len(parts) != 5 {
						t.Errorf("UUID format incorrect: %s", result)
					}
				} else if result != tt.want {
					t.Errorf("EvaluateTemplate() = %v, want %v", result, tt.want)
				}
			}
		})
	}
}

func TestTemplateEngine_EvaluateTemplateToInt64(t *testing.T) {
	fixedTime := time.Unix(1000, 0)
	ctx := &EvaluationContext{
		Clock: &MockClock{now: fixedTime},
	}
	engine := NewTemplateEngine(ctx)

	tests := []struct {
		name    string
		tmpl    string
		want    int64
		wantErr bool
	}{
		{
			name: "simple number",
			tmpl: "42",
			want: 42,
		},
		{
			name: "time unix",
			tmpl: "{{ now | unix }}",
			want: 1000,
		},
		{
			name: "time arithmetic",
			tmpl: "{{ addMinutes 5 now | unix }}",
			want: 1300,
		},
		{
			name:    "invalid number",
			tmpl:    "not a number",
			wantErr: true,
		},
		{
			name:    "template error",
			tmpl:    "{{ invalid }}",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.EvaluateTemplateToInt64(tt.tmpl)

			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateTemplateToInt64() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && result != tt.want {
				t.Errorf("EvaluateTemplateToInt64() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestTemplateEngine_TimeFunctions(t *testing.T) {
	fixedTime := time.Unix(1000, 0)
	ctx := &EvaluationContext{
		Clock: &MockClock{now: fixedTime},
	}
	engine := NewTemplateEngine(ctx)

	// Test now function
	now := engine.now()
	if !now.Equal(fixedTime) {
		t.Errorf("now() = %v, want %v", now, fixedTime)
	}

	// Test unix function
	unix := engine.unix(fixedTime)
	if unix != 1000 {
		t.Errorf("unix() = %v, want %v", unix, 1000)
	}

	// Test rfc3339 function
	rfc3339 := engine.rfc3339(fixedTime)
	expected := fixedTime.Format(time.RFC3339)
	if rfc3339 != expected {
		t.Errorf("rfc3339() = %v, want %v", rfc3339, expected)
	}

	// Test addMinutes function
	added := engine.addMinutes(5, fixedTime)
	expectedTime := fixedTime.Add(5 * time.Minute)
	if !added.Equal(expectedTime) {
		t.Errorf("addMinutes() = %v, want %v", added, expectedTime)
	}

	// Test addHours function
	added = engine.addHours(2, fixedTime)
	expectedTime = fixedTime.Add(2 * time.Hour)
	if !added.Equal(expectedTime) {
		t.Errorf("addHours() = %v, want %v", added, expectedTime)
	}
}

func TestTemplateEngine_RandomFunctions(t *testing.T) {
	ctx := &EvaluationContext{
		Seed: 42, // Fixed seed for deterministic testing
	}
	engine := NewTemplateEngine(ctx)

	// Test deterministic random with seed
	result1 := engine.randInt(1, 100)
	result2 := engine.randInt(1, 100)
	result3 := engine.randInt(1, 100)

	// With fixed seed, results should be deterministic (same sequence each time)
	// Create a new engine with same seed to verify determinism
	engine2 := NewTemplateEngine(&EvaluationContext{Seed: 42})
	result1_2 := engine2.randInt(1, 100)
	result2_2 := engine2.randInt(1, 100)
	result3_2 := engine2.randInt(1, 100)

	// Results should be identical with same seed
	if result1 != result1_2 || result2 != result2_2 || result3 != result3_2 {
		t.Errorf("Deterministic random failed: same seed should produce same sequence")
	}

	// Test bounds
	result := engine.randInt(10, 20)
	if result < 10 || result > 20 {
		t.Errorf("randInt(10, 20) = %v, should be between 10 and 20", result)
	}

	// Test edge case: min >= max
	result = engine.randInt(20, 10)
	if result != 20 {
		t.Errorf("randInt(20, 10) = %v, should return 20", result)
	}

	// Test randFloat
	float1 := engine.randFloat()
	float2 := engine.randFloat()
	if float1 < 0 || float1 > 1 {
		t.Errorf("randFloat() = %v, should be between 0 and 1", float1)
	}
	if float2 < 0 || float2 > 1 {
		t.Errorf("randFloat() = %v, should be between 0 and 1", float2)
	}
}

func TestTemplateEngine_EnvironmentAndVariables(t *testing.T) {
	ctx := &EvaluationContext{
		Variables: map[string]interface{}{
			"test_var": "test_value",
			"number":   42,
		},
	}
	engine := NewTemplateEngine(ctx)

	// Test variable access
	result := engine.getVar("test_var")
	if result != "test_value" {
		t.Errorf("getVar(\"test_var\") = %v, want %v", result, "test_value")
	}

	result = engine.getVar("number")
	if result != 42 {
		t.Errorf("getVar(\"number\") = %v, want %v", result, 42)
	}

	// Test non-existent variable
	result = engine.getVar("nonexistent")
	if result != "" {
		t.Errorf("getVar(\"nonexistent\") = %v, want %v", result, "")
	}

	// Test environment variable (we can't easily test this without setting env vars)
	// But we can test the function exists and doesn't panic
	_ = engine.env("PATH")
}

func TestTemplateEngine_Sequence(t *testing.T) {
	ctx := &EvaluationContext{}
	engine := NewTemplateEngine(ctx)

	// Test sequence counter
	result1 := engine.seq()
	if result1 != 1 {
		t.Errorf("First seq() = %v, want %v", result1, 1)
	}

	result2 := engine.seq()
	if result2 != 2 {
		t.Errorf("Second seq() = %v, want %v", result2, 2)
	}

	result3 := engine.seq()
	if result3 != 3 {
		t.Errorf("Third seq() = %v, want %v", result3, 3)
	}
}

func TestTemplateEngine_UtilityFunctions(t *testing.T) {
	fixedTime := time.Unix(1000, 0)
	ctx := &EvaluationContext{
		Clock: &MockClock{now: fixedTime},
		Seed:  42,
	}
	engine := NewTemplateEngine(ctx)

	// Test jitter function
	jittered := engine.jitter(fixedTime, "30s")
	if jittered.Before(fixedTime) || jittered.After(fixedTime.Add(30*time.Second)) {
		t.Errorf("jitter() = %v, should be between %v and %v", jittered, fixedTime, fixedTime.Add(30*time.Second))
	}

	// Test string functions
	upper := engine.funcMap["upper"].(func(string) string)
	if result := upper("hello"); result != "HELLO" {
		t.Errorf("upper(\"hello\") = %v, want %v", result, "HELLO")
	}

	lower := engine.funcMap["lower"].(func(string) string)
	if result := lower("WORLD"); result != "world" {
		t.Errorf("lower(\"WORLD\") = %v, want %v", result, "world")
	}

	trim := engine.funcMap["trim"].(func(string) string)
	if result := trim("  hello  "); result != "hello" {
		t.Errorf("trim(\"  hello  \") = %v, want %v", result, "hello")
	}
}

func TestTemplateEngine_SetVariable(t *testing.T) {
	ctx := &EvaluationContext{
		Variables: make(map[string]interface{}),
	}
	engine := NewTemplateEngine(ctx)

	// Test setting variables
	engine.SetVariable("key1", "value1")
	engine.SetVariable("key2", 42)

	if ctx.Variables["key1"] != "value1" {
		t.Errorf("Variable key1 not set correctly")
	}
	if ctx.Variables["key2"] != 42 {
		t.Errorf("Variable key2 not set correctly")
	}

	// Test variable substitution in template
	result, err := engine.EvaluateTemplate("{{ .Variables.key1 }} and {{ .Variables.key2 }}")
	if err != nil {
		t.Fatalf("Failed to evaluate template: %v", err)
	}

	expected := "value1 and 42"
	if result != expected {
		t.Errorf("Template evaluation = %v, want %v", result, expected)
	}
}

func TestTemplateEngine_SetSeed(t *testing.T) {
	ctx := &EvaluationContext{}
	engine := NewTemplateEngine(ctx)

	// Test seed setting
	engine.SetSeed(12345)
	if ctx.Seed != 12345 {
		t.Errorf("Seed not set correctly: %v", ctx.Seed)
	}

	// Test deterministic behavior with seed
	result1 := engine.randInt(1, 100)
	result2 := engine.randInt(1, 100)
	result3 := engine.randInt(1, 100)

	// Reset seed and test again
	engine.SetSeed(12345)
	result4 := engine.randInt(1, 100)
	result5 := engine.randInt(1, 100)
	result6 := engine.randInt(1, 100)

	// Results should be identical with same seed
	if result1 != result4 || result2 != result5 || result3 != result6 {
		t.Errorf("Deterministic behavior failed with same seed")
	}
}
