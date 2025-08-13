package spec

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level configuration file
type Config struct {
	Requests []ScheduledRequest `json:"requests" yaml:"requests"`
}

// LoadConfig loads configuration from a file (supports both YAML and JSON)
func LoadConfig(path string) ([]ScheduledRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config

	// Determine format based on file extension
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &config)
	case ".json":
		// For now, we'll use YAML unmarshaler which can handle JSON
		// In the future, we could add explicit JSON support
		err = yaml.Unmarshal(data, &config)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s (use .yaml, .yml, or .json)", ext)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate all requests
	for i, req := range config.Requests {
		if err := req.Validate(); err != nil {
			return nil, fmt.Errorf("request %d (%s): %w", i, req.Name, err)
		}
	}

	return config.Requests, nil
}

// Validate validates the entire configuration
func (c *Config) Validate() error {
	if len(c.Requests) == 0 {
		return &ValidationError{
			Field:   "requests",
			Message: "at least one request must be specified",
		}
	}

	for i, req := range c.Requests {
		if err := req.Validate(); err != nil {
			return fmt.Errorf("request %d (%s): %w", i, req.Name, err)
		}
	}

	return nil
}

// Validate validates a single scheduled request
func (r *ScheduledRequest) Validate() error {
	if r.Name == "" {
		return &ValidationError{
			Field:   "name",
			Message: "request name is required",
		}
	}

	if err := r.Schedule.Validate(); err != nil {
		return err
	}

	if err := r.HTTP.Validate(); err != nil {
		return err
	}

	return nil
}

// Validate validates HTTP request specification
func (h *HttpRequestSpec) Validate() error {
	if h.Method == "" {
		return &ValidationError{
			Field:   "http.method",
			Message: "HTTP method is required",
		}
	}

	if h.URL == "" {
		return &ValidationError{
			Field:   "http.url",
			Message: "HTTP URL is required",
		}
	}

	// Validate HTTP method
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true,
		"HEAD": true, "OPTIONS": true, "TRACE": true,
	}

	if !validMethods[strings.ToUpper(h.Method)] {
		return &ValidationError{
			Field:   "http.method",
			Message: fmt.Sprintf("invalid HTTP method: %s", h.Method),
		}
	}

	return nil
}
