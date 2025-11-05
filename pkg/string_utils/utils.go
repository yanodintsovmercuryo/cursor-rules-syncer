package string_utils

import (
	"strings"
)

// SplitTrimFilter splits string by separator, trims spaces and filters empty strings
func SplitTrimFilter(input, separator string) []string {
	if input == "" {
		return []string{}
	}
	
	result := []string{}
	for _, part := range strings.Split(input, separator) {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// RemoveDuplicates preserves order while removing duplicates
func RemoveDuplicates(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}

	seen := make(map[string]bool)
	result := []string{}
	for _, item := range items {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}
