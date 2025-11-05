package config

import (
	"testing"
)

func TestParseBoolFromString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"true lowercase", "true", true},
		{"true uppercase", "TRUE", true},
		{"true mixed case", "True", true},
		{"one", "1", true},
		{"false lowercase", "false", false},
		{"false uppercase", "FALSE", false},
		{"false mixed case", "False", false},
		{"zero", "0", false},
		{"empty string", "", false},
		{"whitespace", " ", false},
		{"whitespace multiple", "   ", false},
		{"tab", "\t", false},
		{"newline", "\n", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := parseBoolFromString(tt.input)
			if result != tt.expected {
				t.Errorf("parseBoolFromString(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
