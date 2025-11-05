package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileFilterService handles file pattern matching and filtering
type FileFilterService struct {
	outputService *OutputService
}

// NewFileFilterService creates a new FileFilterService
func NewFileFilterService(outputService *OutputService) *FileFilterService {
	return &FileFilterService{
		outputService: outputService,
	}
}

// GetFilePatterns returns file patterns from various sources
func (s *FileFilterService) GetFilePatterns(flagValue, envVarName string) ([]string, error) {
	patterns := []string{}

	// Priority: flag > environment variable
	patternsSource := flagValue
	if patternsSource == "" {
		patternsSource = os.Getenv(envVarName)
	}

	if patternsSource != "" {
		// Parse comma-separated patterns
		for _, pattern := range strings.Split(patternsSource, ",") {
			pattern = strings.TrimSpace(pattern)
			if pattern != "" {
				patterns = append(patterns, pattern)
			}
		}
	}

	return patterns, nil
}

// FilterFilesByPatterns filters files based on provided patterns
// If patterns is empty, returns all files
func (s *FileFilterService) FilterFilesByPatterns(files []string, baseDir string, patterns []string) []string {
	if len(patterns) == 0 {
		return files
	}

	var filtered []string
	for _, file := range files {
		var relativePath string

		if baseDir != "" {
			// Try to get relative path
			rel, err := filepath.Rel(baseDir, file)
			if err != nil {
				// If we can't get relative path, fall back to base name
				relativePath = filepath.Base(file)
			} else {
				relativePath = rel
			}
		} else {
			// No base dir provided, use base name
			relativePath = filepath.Base(file)
		}

		if s.MatchesAnyPattern(relativePath, patterns) {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// matchesAnyPattern checks if a file path matches any of the provided patterns
func (s *FileFilterService) MatchesAnyPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		if s.MatchesPattern(filePath, pattern) {
			return true
		}
	}
	return false
}

// matchesPattern checks if a file path matches a single pattern
// Uses filepath.Match which already supports glob patterns (*, ?, etc.)
func (s *FileFilterService) MatchesPattern(filePath, pattern string) bool {
	// Normalize path separators
	normalizedPath := filepath.ToSlash(filePath)
	normalizedPattern := filepath.ToSlash(pattern)

	// Use filepath.Match directly with glob pattern (no regex conversion needed)
	// Check exact match
	if match, err := filepath.Match(normalizedPattern, normalizedPath); err == nil && match {
		return true
	}

	// Also check match against just the filename
	filename := filepath.Base(normalizedPath)
	if match, err := filepath.Match(normalizedPattern, filename); err == nil && match {
		return true
	}

	// Check parent directory patterns
	dir := filepath.Dir(normalizedPath)
	for dir != "." && dir != "/" {
		if match, err := filepath.Match(normalizedPattern, dir); err == nil && match {
			return true
		}
		dir = filepath.Dir(dir)
	}

	return false
}

// FindFilesByPatterns finds all files matching patterns in directory
func (s *FileFilterService) FindFilesByPatterns(dir string, patterns []string) ([]string, error) {
	allFiles, err := s.findAllFiles(dir)
	if err != nil {
		return nil, err
	}

	return s.FilterFilesByPatterns(allFiles, dir, patterns), nil
}

// findAllFiles finds all files in the specified directory recursively
func (s *FileFilterService) findAllFiles(dir string) ([]string, error) {
	var allFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			allFiles = append(allFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return allFiles, nil
}

// CleanupExtraFilesByPatterns removes files that exist in destination but not in source, considering patterns
func (s *FileFilterService) CleanupExtraFilesByPatterns(srcFiles []string, srcBase, dstBase string, patterns []string) error {
	// Build map of source files by relative path (filtered by patterns)
	srcFilesMap := make(map[string]bool)
	for _, srcFile := range srcFiles {
		relativePath, err := filepath.Rel(srcBase, srcFile)
		if err != nil {
			continue
		}
		srcFilesMap[relativePath] = true
	}

	// Find all destination files
	var destFiles []string
	err := filepath.Walk(dstBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			destFiles = append(destFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking destination directory: %w", err)
	}

	// Filter destination files by patterns
	destFiles = s.FilterFilesByPatterns(destFiles, dstBase, patterns)

	// Remove files that don't exist in source
	for _, destFile := range destFiles {
		relativePath, err := filepath.Rel(dstBase, destFile)
		if err != nil {
			continue
		}

		if !srcFilesMap[relativePath] {
			if err := os.Remove(destFile); err != nil {
				s.outputService.PrintErrorf("Error deleting file %s: %v\n", relativePath, err)
			} else {
				s.outputService.PrintOperation("delete", relativePath)
			}
		}
	}

	return nil
}

// GetEffectivePatterns returns effective patterns, empty slice means no filtering
func (s *FileFilterService) GetEffectivePatterns(patterns []string) []string {
	if len(patterns) == 0 {
		return []string{} // No filtering - sync all files
	}

	// Remove duplicates while preserving order
	var effectivePatterns []string
	seen := make(map[string]bool)
	for _, pattern := range patterns {
		if !seen[pattern] {
			effectivePatterns = append(effectivePatterns, pattern)
			seen[pattern] = true
		}
	}

	return effectivePatterns
}

// ValidatePatterns validates if patterns are well-formed
func (s *FileFilterService) ValidatePatterns(patterns []string) error {
	for _, pattern := range patterns {
		// Test if pattern works with filepath.Match
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("invalid pattern '%s': %w", pattern, err)
		}
	}
	return nil
}

// PatternStats provides statistics about pattern matching
type PatternStats struct {
	TotalFiles      int
	MatchedFiles    int
	MatchedPatterns map[string]int
}

// AnalyzePatternMatching analyzes how patterns match against files
func (s *FileFilterService) AnalyzePatternMatching(files []string, baseDir string, patterns []string) *PatternStats {
	stats := &PatternStats{
		TotalFiles:      len(files),
		MatchedFiles:    0,
		MatchedPatterns: make(map[string]int),
	}

	for _, file := range files {
		relativePath, err := filepath.Rel(baseDir, file)
		if err != nil {
			continue
		}

		for _, pattern := range patterns {
			if s.MatchesPattern(relativePath, pattern) {
				stats.MatchedFiles++
				stats.MatchedPatterns[pattern]++
				break
			}
		}
	}

	return stats
}
