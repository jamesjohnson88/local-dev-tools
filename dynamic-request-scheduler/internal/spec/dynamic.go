package spec

import (
	"encoding/json"
	"fmt"
	"strings"
)

// DynamicString represents a string that can be either literal or a template
type DynamicString struct {
	value      string
	template   string
	isTemplate bool
}

// DynamicInt64 represents an int64 that can be either literal or a template
type DynamicInt64 struct {
	value      int64
	template   string
	isTemplate bool
}

// DynamicAny represents any value that can be either literal or a template
type DynamicAny struct {
	value      interface{}
	template   string
	isTemplate bool
}

// UnmarshalJSON implements json.Unmarshaler for DynamicString
func (d *DynamicString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a string first (literal value)
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		d.value = s
		d.isTemplate = false
		return nil
	}

	// Try to unmarshal as a template object
	var templateObj struct {
		Template string `json:"template"`
	}
	if err := json.Unmarshal(data, &templateObj); err == nil && templateObj.Template != "" {
		d.template = templateObj.Template
		d.isTemplate = true
		return nil
	}

	return fmt.Errorf("DynamicString must be a string or {template: \"...\"}")
}

// UnmarshalJSON implements json.Unmarshaler for DynamicInt64
func (d *DynamicInt64) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a number first (literal value)
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		d.value = n
		d.isTemplate = false
		return nil
	}

	// Try to unmarshal as a template object
	var templateObj struct {
		Template string `json:"template"`
	}
	if err := json.Unmarshal(data, &templateObj); err == nil && templateObj.Template != "" {
		d.template = templateObj.Template
		d.isTemplate = true
		return nil
	}

	return fmt.Errorf("DynamicInt64 must be a number or {template: \"...\"}")
}

// UnmarshalJSON implements json.Unmarshaler for DynamicAny
func (d *DynamicAny) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as a template object first
	var templateObj struct {
		Template string `json:"template"`
	}
	if err := json.Unmarshal(data, &templateObj); err == nil && templateObj.Template != "" {
		d.template = templateObj.Template
		d.isTemplate = true
		return nil
	}

	// Otherwise, treat as literal value
	d.value = json.RawMessage(data)
	d.isTemplate = false
	return nil
}

// IsTemplate returns true if this dynamic value contains a template
func (d *DynamicString) IsTemplate() bool { return d.isTemplate }
func (d *DynamicInt64) IsTemplate() bool  { return d.isTemplate }
func (d *DynamicAny) IsTemplate() bool    { return d.isTemplate }

// GetTemplate returns the template string if this is a template, empty string otherwise
func (d *DynamicString) GetTemplate() string { return d.template }
func (d *DynamicInt64) GetTemplate() string  { return d.template }
func (d *DynamicAny) GetTemplate() string    { return d.template }

// GetValue returns the literal value if this is not a template
func (d *DynamicString) GetValue() string   { return d.value }
func (d *DynamicInt64) GetValue() int64     { return d.value }
func (d *DynamicAny) GetValue() interface{} { return d.value }

// String returns a string representation for debugging
func (d *DynamicString) String() string {
	if d.isTemplate {
		return fmt.Sprintf("template:\"%s\"", d.template)
	}
	return fmt.Sprintf("literal:\"%s\"", d.value)
}

func (d *DynamicInt64) String() string {
	if d.isTemplate {
		return fmt.Sprintf("template:\"%s\"", d.template)
	}
	return fmt.Sprintf("literal:%d", d.value)
}

func (d *DynamicAny) String() string {
	if d.isTemplate {
		return fmt.Sprintf("template:\"%s\"", d.template)
	}
	return fmt.Sprintf("literal:%v", d.value)
}

// IsTemplateString checks if a string contains template syntax
func IsTemplateString(s string) bool {
	return strings.Contains(s, "{{") && strings.Contains(s, "}}")
}

// ExtractTemplateStrings recursively finds all template strings in a value
func ExtractTemplateStrings(v interface{}) []string {
	var templates []string

	switch val := v.(type) {
	case string:
		if IsTemplateString(val) {
			templates = append(templates, val)
		}
	case map[string]interface{}:
		for _, item := range val {
			templates = append(templates, ExtractTemplateStrings(item)...)
		}
	case []interface{}:
		for _, item := range val {
			templates = append(templates, ExtractTemplateStrings(item)...)
		}
	case DynamicString:
		if val.IsTemplate() {
			templates = append(templates, val.GetTemplate())
		}
	case DynamicInt64:
		if val.IsTemplate() {
			templates = append(templates, val.GetTemplate())
		}
	case DynamicAny:
		if val.IsTemplate() {
			templates = append(templates, val.GetTemplate())
		}
	}

	return templates
}
