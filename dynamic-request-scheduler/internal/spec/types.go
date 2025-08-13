package spec

import (
	"time"
)

// ScheduledRequest represents a request that will be scheduled and executed
type ScheduledRequest struct {
	Name     string          `json:"name" yaml:"name"`
	Schedule ScheduleSpec    `json:"schedule" yaml:"schedule"`
	HTTP     HttpRequestSpec `json:"http" yaml:"http"`
}

// HttpRequestSpec defines the HTTP request to be made
type HttpRequestSpec struct {
	Method  string            `json:"method" yaml:"method"`
	URL     string            `json:"url" yaml:"url"`
	Headers map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Body    interface{}       `json:"body,omitempty" yaml:"body,omitempty"`
}

// ScheduleSpec defines when the request should be executed
// Only one of the fields should be set
type ScheduleSpec struct {
	// Epoch represents a specific Unix timestamp
	Epoch *int64 `json:"epoch,omitempty" yaml:"epoch,omitempty"`

	// Relative represents a duration from now (e.g., "5m", "1h")
	Relative *string `json:"relative,omitempty" yaml:"relative,omitempty"`

	// Template represents a Go template that evaluates to a Unix timestamp
	Template *string `json:"template,omitempty" yaml:"template,omitempty"`

	// Cron represents a cron expression (e.g., "*/5 * * * *")
	Cron *string `json:"cron,omitempty" yaml:"cron,omitempty"`

	// Jitter adds random variation to the scheduled time (e.g., "±30s")
	Jitter *string `json:"jitter,omitempty" yaml:"jitter,omitempty"`
}

// Validate ensures only one schedule strategy is specified
func (s *ScheduleSpec) Validate() error {
	count := 0
	if s.Epoch != nil {
		count++
	}
	if s.Relative != nil {
		count++
	}
	if s.Template != nil {
		count++
	}
	if s.Cron != nil {
		count++
	}

	if count != 1 {
		return &ValidationError{
			Field:   "schedule",
			Message: "exactly one schedule strategy must be specified (epoch, relative, template, or cron)",
		}
	}

	return nil
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}

// ResolvedRequest represents a request with all dynamic values resolved
type ResolvedRequest struct {
	Name         string
	Method       string
	URL          string
	Headers      map[string]string
	Body         interface{}
	ScheduledFor time.Time
}
