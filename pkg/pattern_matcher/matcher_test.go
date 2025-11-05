package pattern_matcher_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yanodintsovmercuryo/cursync/pkg/pattern_matcher"
)

func TestMatchesPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		pattern  string
		expected bool
	}{
		{
			name:     "exact match",
			filePath: "file.txt",
			pattern:  "file.txt",
			expected: true,
		},
		{
			name:     "wildcard match",
			filePath: "file.txt",
			pattern:  "*.txt",
			expected: true,
		},
		{
			name:     "no match",
			filePath: "file.txt",
			pattern:  "*.md",
			expected: false,
		},
		{
			name:     "match by filename",
			filePath: "dir/sub/file.txt",
			pattern:  "file.txt",
			expected: true,
		},
		{
			name:     "match by directory",
			filePath: "dir/sub/file.txt",
			pattern:  "dir",
			expected: true,
		},
		{
			name:     "match by subdirectory",
			filePath: "dir/sub/file.txt",
			pattern:  "sub",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := pattern_matcher.MatchesPattern(tt.filePath, tt.pattern)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestMatchesAnyPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		filePath string
		patterns []string
		expected bool
	}{
		{
			name:     "matches first pattern",
			filePath: "file.txt",
			patterns: []string{"*.txt", "*.md"},
			expected: true,
		},
		{
			name:     "matches second pattern",
			filePath: "file.md",
			patterns: []string{"*.txt", "*.md"},
			expected: true,
		},
		{
			name:     "no match",
			filePath: "file.go",
			patterns: []string{"*.txt", "*.md"},
			expected: false,
		},
		{
			name:     "empty patterns",
			filePath: "file.txt",
			patterns: []string{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := pattern_matcher.MatchesAnyPattern(tt.filePath, tt.patterns)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
