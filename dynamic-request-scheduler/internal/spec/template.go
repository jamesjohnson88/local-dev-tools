package spec

import (
	"crypto/rand"
	"fmt"
	"math"
	mrand "math/rand"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// TemplateEngine provides template evaluation functionality
type TemplateEngine struct {
	funcMap template.FuncMap
	ctx     *EvaluationContext
}

// EvaluationContext holds variables and state for template evaluation
type EvaluationContext struct {
	Variables map[string]interface{}
	Sequence  int64
	Seed      int64
	Clock     Clock
	randSource *mrand.Rand
}

// Clock interface for time operations (allows injection for testing)
type Clock interface {
	Now() time.Time
}

// RealClock implements Clock using system time
type RealClock struct{}

func (r *RealClock) Now() time.Time { return time.Now().UTC() }

// NewTemplateEngine creates a new template engine with the standard function map
func NewTemplateEngine(ctx *EvaluationContext) *TemplateEngine {
	if ctx == nil {
		ctx = &EvaluationContext{
			Variables: make(map[string]interface{}),
			Clock:     &RealClock{},
		}
	}

	engine := &TemplateEngine{
		ctx: ctx,
	}

	engine.funcMap = template.FuncMap{
		// Time functions
		"now":        engine.now,
		"unix":       engine.unix,
		"rfc3339":    engine.rfc3339,
		"addSeconds": engine.addSeconds,
		"addMinutes": engine.addMinutes,
		"addHours":   engine.addHours,
		"parseTime":  engine.parseTime,

		// ID and random functions
		"uuid":      engine.uuid,
		"randInt":   engine.randInt,
		"randFloat": engine.randFloat,

		// Environment and variables
		"env": engine.env,
		"var": engine.getVar,

		// Sequence and iteration
		"seq": engine.seq,

		// Utility functions
		"jitter": engine.jitter,
		"upper":  strings.ToUpper,
		"lower":  strings.ToLower,
		"trim":   strings.TrimSpace,
	}

	return engine
}

// EvaluateTemplate evaluates a template string and returns the result
func (e *TemplateEngine) EvaluateTemplate(tmpl string) (string, error) {
	t, err := template.New("dynamic").Funcs(e.funcMap).Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var result strings.Builder
	err = t.Execute(&result, e.ctx)
	if err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return result.String(), nil
}

// EvaluateTemplateToInt64 evaluates a template string and returns an int64 result
func (e *TemplateEngine) EvaluateTemplateToInt64(tmpl string) (int64, error) {
	result, err := e.EvaluateTemplate(tmpl)
	if err != nil {
		return 0, err
	}

	// Try to parse as int64
	val, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("template result '%s' is not a valid int64: %w", result, err)
	}

	return val, nil
}

// Time functions
func (e *TemplateEngine) now() time.Time {
	return e.ctx.Clock.Now()
}

func (e *TemplateEngine) unix(t time.Time) int64 {
	return t.Unix()
}

func (e *TemplateEngine) rfc3339(t time.Time) string {
	return t.Format(time.RFC3339)
}

func (e *TemplateEngine) addSeconds(seconds int, t time.Time) time.Time {
	return t.Add(time.Duration(seconds) * time.Second)
}

func (e *TemplateEngine) addMinutes(minutes int, t time.Time) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}

func (e *TemplateEngine) addHours(hours int, t time.Time) time.Time {
	return t.Add(time.Duration(hours) * time.Hour)
}

func (e *TemplateEngine) parseTime(layout, value string) (time.Time, error) {
	return time.Parse(layout, value)
}

// ID and random functions
func (e *TemplateEngine) uuid() string {
	// Generate a simple UUID v4
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based ID if crypto/rand fails
		return fmt.Sprintf("%d-%d", time.Now().UnixNano(), e.ctx.Sequence)
	}

	// Set version (4) and variant bits
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func (e *TemplateEngine) randInt(min, max int) int {
	if min >= max {
		return min
	}

	// Use seed if available for deterministic results
	if e.ctx.Seed != 0 {
		// Use a separate random source for deterministic results
		if e.ctx.randSource == nil {
			e.ctx.randSource = mrand.New(mrand.NewSource(e.ctx.Seed))
		}
		return min + e.ctx.randSource.Intn(max-min+1)
	}

	// Non-deterministic random
	return min + int(time.Now().UnixNano())%(max-min+1)
}

func (e *TemplateEngine) randFloat() float64 {
	if e.ctx.Seed != 0 {
		// Use a separate random source for deterministic results
		if e.ctx.randSource == nil {
			e.ctx.randSource = mrand.New(mrand.NewSource(e.ctx.Seed))
		}
		return e.ctx.randSource.Float64()
	}
	return float64(time.Now().UnixNano()) / float64(math.MaxInt64)
}

// Environment and variables
func (e *TemplateEngine) env(key string) string {
	return os.Getenv(key)
}

func (e *TemplateEngine) getVar(key string) interface{} {
	if val, exists := e.ctx.Variables[key]; exists {
		return val
	}
	return ""
}

// Sequence and iteration
func (e *TemplateEngine) seq() int64 {
	e.ctx.Sequence++
	return e.ctx.Sequence
}

// Utility functions
func (e *TemplateEngine) jitter(base time.Time, duration string) time.Time {
	d, err := time.ParseDuration(duration)
	if err != nil {
		return base
	}

	// Add random jitter within the duration range
	jitterAmount := time.Duration(e.randInt(0, int(d.Nanoseconds())))
	return base.Add(jitterAmount)
}

// SetVariable sets a variable in the evaluation context
func (e *TemplateEngine) SetVariable(key string, value interface{}) {
	if e.ctx.Variables == nil {
		e.ctx.Variables = make(map[string]interface{})
	}
	e.ctx.Variables[key] = value
}

// SetSeed sets the seed for deterministic random functions
func (e *TemplateEngine) SetSeed(seed int64) {
	e.ctx.Seed = seed
	e.ctx.randSource = nil // Reset random source so new one is created with new seed
}

// GetContext returns the evaluation context
func (e *TemplateEngine) GetContext() *EvaluationContext {
	return e.ctx
}
