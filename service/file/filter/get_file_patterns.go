package filter

import (
	"os"

	"github.com/yanodintsovmercuryo/cursync/pkg/string_utils"
)

// GetFilePatterns returns file patterns from various sources
func (f *Filter) GetFilePatterns(flagValue, envVarName string) ([]string, error) {
	patterns := []string{}

	patternsSource := flagValue
	if patternsSource == "" {
		patternsSource = os.Getenv(envVarName)
	}

	if patternsSource != "" {
		patterns = string_utils.SplitTrimFilter(patternsSource, ",")
	}

	return patterns, nil
}
