package utils

import (
	"testing"
)

func TestStripProtocolFromDomain(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		hasError bool
	}{
		{
			name:     "Basic domain",
			input:    "example.com",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "HTTPS URL",
			input:    "https://c.yoyfun.biz.ua/",
			expected: "c.yoyfun.biz.ua",
			hasError: false,
		},
		{
			name:     "HTTP URL",
			input:    "http://yoyfun.biz.ua",
			expected: "yoyfun.biz.ua",
			hasError: false,
		},
		{
			name:     "Domain with port",
			input:    "example.com:8080",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "URL with port and path",
			input:    "https://api.example.com:3000/health",
			expected: "api.example.com",
			hasError: false,
		},
		{
			name:     "Subdomain",
			input:    "api.v1.example.com",
			expected: "api.v1.example.com",
			hasError: false,
		},
		{
			name:     "Domain with trailing slash",
			input:    "https://example.com/",
			expected: "example.com",
			hasError: false,
		},
		{
			name:     "Empty input",
			input:    "",
			expected: "",
			hasError: true,
		},
		{
			name:     "Invalid domain",
			input:    "not..valid",
			expected: "",
			hasError: true,
		},
		{
			name:     "Just protocol",
			input:    "https://",
			expected: "",
			hasError: true,
		},
		{
			name:     "Domain with spaces",
			input:    "  example.com  ",
			expected: "example.com",
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := StripProtocolFromDomain(tt.input)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %q, got %q for input %q", tt.expected, result, tt.input)
				}
			}
		})
	}
}

func TestValidateDomains(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
		hasError bool
	}{
		{
			name:     "Valid domains",
			input:    []string{"example.com", "https://api.example.com", "test.org"},
			expected: []string{"example.com", "api.example.com", "test.org"},
			hasError: false,
		},
		{
			name:     "Mixed valid and invalid",
			input:    []string{"example.com", "not..valid", "api.example.com"},
			expected: []string{"example.com", "api.example.com"},
			hasError: true,
		},
		{
			name:     "Empty domains filtered",
			input:    []string{"example.com", "", "api.example.com"},
			expected: []string{"example.com", "api.example.com"},
			hasError: false,
		},
		{
			name:     "All invalid",
			input:    []string{"not..valid", ""},
			expected: []string{},
			hasError: true,
		},
		{
			name:     "Empty slice",
			input:    []string{},
			expected: []string{},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateDomains(tt.input)
			
			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %v, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %v: %v", tt.input, err)
				}
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d domains, got %d for input %v", len(tt.expected), len(result), tt.input)
			}

			for i, expected := range tt.expected {
				if i >= len(result) || result[i] != expected {
					t.Errorf("Expected domain %q at index %d, got %q for input %v", expected, i, result[i], tt.input)
				}
			}
		})
	}
}