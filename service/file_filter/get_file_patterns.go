package file_filter

import (
	"os"

	"github.com/yanodintsovmercuryo/cursor-rules-syncer/pkg/string_utils"
)

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
		patterns = string_utils.SplitTrimFilter(patternsSource, ",")
	}

	return patterns, nil
}
