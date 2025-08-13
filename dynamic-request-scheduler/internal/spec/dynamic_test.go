package spec

import (
	"encoding/json"
	"testing"
)

func TestDynamicString_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DynamicString
		wantErr bool
	}{
		{
			name:  "literal string",
			input: `"hello world"`,
			want: DynamicString{
				value:      "hello world",
				template:   "",
				isTemplate: false,
			},
			wantErr: false,
		},
		{
			name:  "template object",
			input: `{"template": "{{ now | unix }}"}`,
			want: DynamicString{
				value:      "",
				template:   "{{ now | unix }}",
				isTemplate: true,
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "empty template",
			input:   `{"template": ""}`,
			wantErr: true,
		},
		{
			name:    "missing template field",
			input:   `{"other": "value"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DynamicString
			err := json.Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("DynamicString.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.value != tt.want.value {
					t.Errorf("DynamicString.UnmarshalJSON() value = %v, want %v", got.value, tt.want.value)
				}
				if got.template != tt.want.template {
					t.Errorf("DynamicString.UnmarshalJSON() template = %v, want %v", got.template, tt.want.template)
				}
				if got.isTemplate != tt.want.isTemplate {
					t.Errorf("DynamicString.UnmarshalJSON() isTemplate = %v, want %v", got.isTemplate, tt.want.isTemplate)
				}
			}
		})
	}
}

func TestDynamicInt64_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DynamicInt64
		wantErr bool
	}{
		{
			name:  "literal number",
			input: `42`,
			want: DynamicInt64{
				value:      42,
				template:   "",
				isTemplate: false,
			},
			wantErr: false,
		},
		{
			name:  "template object",
			input: `{"template": "{{ now | unix }}"}`,
			want: DynamicInt64{
				value:      0,
				template:   "{{ now | unix }}",
				isTemplate: true,
			},
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "empty template",
			input:   `{"template": ""}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DynamicInt64
			err := json.Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("DynamicInt64.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if got.value != tt.want.value {
					t.Errorf("DynamicInt64.UnmarshalJSON() value = %v, want %v", got.value, tt.want.value)
				}
				if got.template != tt.want.template {
					t.Errorf("DynamicInt64.UnmarshalJSON() template = %v, want %v", got.template, tt.want.template)
				}
				if got.isTemplate != tt.want.isTemplate {
					t.Errorf("DynamicInt64.UnmarshalJSON() isTemplate = %v, want %v", got.isTemplate, tt.want.isTemplate)
				}
			}
		})
	}
}

func TestDynamicAny_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "literal object",
			input:   `{"key": "value"}`,
			wantErr: false,
		},
		{
			name:    "template object",
			input:   `{"template": "{{ now | unix }}"}`,
			wantErr: false,
		},
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DynamicAny
			err := json.Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("DynamicAny.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsTemplateString(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want bool
	}{
		{"empty string", "", false},
		{"no template", "hello world", false},
		{"single brace", "hello {world", false},
		{"double brace start", "{{hello world", false},
		{"double brace end", "hello world}}", false},
		{"valid template", "{{ now | unix }}", true},
		{"template with text", "hello {{ now | unix }} world", true},
		{"multiple templates", "{{ now }} and {{ uuid }}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsTemplateString(tt.s); got != tt.want {
				t.Errorf("IsTemplateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractTemplateStrings(t *testing.T) {
	tests := []struct {
		name string
		v    interface{}
		want []string
	}{
		{
			name: "string with template",
			v:    "{{ now | unix }}",
			want: []string{"{{ now | unix }}"},
		},
		{
			name: "string without template",
			v:    "hello world",
			want: []string{},
		},
		{
			name: "map with templates",
			v: map[string]interface{}{
				"key1": "{{ uuid }}",
				"key2": "static value",
				"key3": "{{ now | rfc3339 }}",
			},
			want: []string{"{{ uuid }}", "{{ now | rfc3339 }}"},
		},
		{
			name: "slice with templates",
			v: []interface{}{
				"{{ seq }}",
				"static",
				"{{ env \"API_KEY\" }}",
			},
			want: []string{"{{ seq }}", "{{ env \"API_KEY\" }}"},
		},
		{
			name: "nested structure",
			v: map[string]interface{}{
				"outer": map[string]interface{}{
					"inner": "{{ now }}",
					"deep": []interface{}{
						"{{ uuid }}",
						"static",
					},
				},
			},
			want: []string{"{{ now }}", "{{ uuid }}"},
		},
		{
			name: "DynamicString with template",
			v: DynamicString{
				template:   "{{ now | unix }}",
				isTemplate: true,
			},
			want: []string{"{{ now | unix }}"},
		},
		{
			name: "DynamicString without template",
			v: DynamicString{
				value:      "static",
				isTemplate: false,
			},
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTemplateStrings(tt.v)
			if len(got) != len(tt.want) {
				t.Errorf("ExtractTemplateStrings() returned %d templates, want %d", len(got), len(tt.want))
				return
			}

			// Check that all expected templates are found
			for _, expected := range tt.want {
				found := false
				for _, actual := range got {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected template '%s' not found in result", expected)
				}
			}
		})
	}
}

func TestDynamicString_Methods(t *testing.T) {
	ds := DynamicString{
		value:      "hello",
		template:   "{{ now }}",
		isTemplate: true,
	}

	if !ds.IsTemplate() {
		t.Error("IsTemplate() should return true")
	}

	if ds.GetTemplate() != "{{ now }}" {
		t.Errorf("GetTemplate() = %v, want %v", ds.GetTemplate(), "{{ now }}")
	}

	if ds.GetValue() != "hello" {
		t.Errorf("GetValue() = %v, want %v", ds.GetValue(), "hello")
	}

	expectedStr := "template:\"{{ now }}\""
	if ds.String() != expectedStr {
		t.Errorf("String() = %v, want %v", ds.String(), expectedStr)
	}
}

func TestDynamicInt64_Methods(t *testing.T) {
	di := DynamicInt64{
		value:      42,
		template:   "{{ now | unix }}",
		isTemplate: true,
	}

	if !di.IsTemplate() {
		t.Error("IsTemplate() should return true")
	}

	if di.GetTemplate() != "{{ now | unix }}" {
		t.Errorf("GetTemplate() = %v, want %v", di.GetTemplate(), "{{ now | unix }}")
	}

	if di.GetValue() != 42 {
		t.Errorf("GetValue() = %v, want %v", di.GetValue(), 42)
	}

	expectedStr := "template:\"{{ now | unix }}\""
	if di.String() != expectedStr {
		t.Errorf("String() = %v, want %v", di.String(), expectedStr)
	}
}
