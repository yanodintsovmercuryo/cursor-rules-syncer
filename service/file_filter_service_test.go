package service

import (
	"testing"
)

func TestMatchesPatternGoFiles(t *testing.T) {
	outputService := NewOutputService()
	fileFilterService := NewFileFilterService(outputService)

	tests := []struct {
		pattern     string
		filepath    string
		expected    bool
		description string
	}{
		{
			pattern:     "go-*.mdc",
			filepath:    "go-concurrency.mdc",
			expected:    true,
			description: "Go concurrency file should match",
		},
		{
			pattern:     "go-*.mdc",
			filepath:    "go-development.mdc",
			expected:    true,
			description: "Go development file should match",
		},
		{
			pattern:     "go-*.mdc",
			filepath:    "go-test.mdc",
			expected:    true,
			description: "Go test file should match",
		},
		{
			pattern:     "go-*.mdc",
			filepath:    "auto-git.mdc",
			expected:    false,
			description: "Auto git file should NOT match",
		},
		{
			pattern:     "go-*.mdc",
			filepath:    "auto-go.mdc",
			expected:    false,
			description: "Auto go file should NOT match",
		},
		{
			pattern:     "go-*.mdc",
			filepath:    "rules/go-concurrency.mdc",
			expected:    true,
			description: "Go file with path should match",
		},
		{
			pattern:     "go-*.mdc",
			filepath:    "/full/path/go-test.mdc",
			expected:    true,
			description: "Go file with full path should match",
		},
		{
			pattern:     "*.mdc",
			filepath:    "anyfile.mdc",
			expected:    true,
			description: "Any .mdc file should match *.mdc pattern",
		},
		{
			pattern:     "rules/**/*.mdc",
			filepath:    "rules/subfolder/file.mdc",
			expected:    true,
			description: "Nested files should match ** pattern",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			result := fileFilterService.MatchesPattern(test.filepath, test.pattern)
			if result != test.expected {
				t.Errorf("Pattern %s should match %s: expected %v, got %v",
					test.pattern, test.filepath, test.expected, result)
			}
		})
	}
}

func TestFilterFilesByPatterns(t *testing.T) {
	outputService := NewOutputService()
	fileFilterService := NewFileFilterService(outputService)

	testFiles := []string{
		"rules/go-concurrency.mdc",
		"rules/go-development.mdc",
		"rules/go-test.mdc",
		"rules/auto-git.mdc",
		"rules/auto-go.mdc",
		"rules/meta-rules.mdc",
	}

	tests := []struct {
		patterns    []string
		baseDir     string
		expected    []string
		description string
	}{
		{
			patterns:    []string{"go-*.mdc"},
			baseDir:     "rules",
			expected:    []string{"rules/go-concurrency.mdc", "rules/go-development.mdc", "rules/go-test.mdc"},
			description: "Should filter Go patterns with baseDir",
		},
		{
			patterns:    []string{"go-*.mdc"},
			baseDir:     "",
			expected:    []string{"rules/go-concurrency.mdc", "rules/go-development.mdc", "rules/go-test.mdc"},
			description: "Should filter Go patterns without baseDir",
		},
		{
			patterns:    []string{"auto-*.mdc"},
			baseDir:     "rules",
			expected:    []string{"rules/auto-git.mdc", "rules/auto-go.mdc"},
			description: "Should filter auto patterns",
		},
		{
			patterns:    []string{"*.mdc"},
			baseDir:     "rules",
			expected:    []string{"rules/go-concurrency.mdc", "rules/go-development.mdc", "rules/go-test.mdc", "rules/auto-git.mdc", "rules/auto-go.mdc", "rules/meta-rules.mdc"},
			description: "Should match all .mdc files with wildcard",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			filtered := fileFilterService.FilterFilesByPatterns(testFiles, test.baseDir, test.patterns)

			if len(filtered) != len(test.expected) {
				t.Errorf("Expected %d files, got %d", len(test.expected), len(filtered))
				return
			}

			for _, expectedFile := range test.expected {
				found := false
				for _, actualFile := range filtered {
					if actualFile == expectedFile {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected file %s not found in filtered results", expectedFile)
				}
			}
		})
	}
}
