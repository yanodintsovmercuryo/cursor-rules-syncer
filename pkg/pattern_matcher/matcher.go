package pattern_matcher

import (
	"path/filepath"
	"strings"
)

// MatchesPattern checks if file path matches the pattern
func MatchesPattern(filePath, pattern string) bool {
	normalizedPath := filepath.ToSlash(filePath)
	normalizedPattern := filepath.ToSlash(pattern)

	// Check full path
	if match, err := filepath.Match(normalizedPattern, normalizedPath); err == nil && match {
		return true
	}

	// Check filename
	filename := filepath.Base(normalizedPath)
	if match, err := filepath.Match(normalizedPattern, filename); err == nil && match {
		return true
	}

	// Check each path part separately
	pathParts := strings.Split(normalizedPath, "/")
	for _, part := range pathParts {
		if part != "" {
			if match, err := filepath.Match(normalizedPattern, part); err == nil && match {
				return true
			}
		}
	}

	return false
}

// MatchesAnyPattern checks if file path matches any of the provided patterns
func MatchesAnyPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		if MatchesPattern(filePath, pattern) {
			return true
		}
	}

	return false
}
