package pattern

import (
	"path/filepath"

	"github.com/yanodintsovmercuryo/cursync/pkg/pattern_matcher"
)

// FilterFilesByPatterns filters files based on provided patterns
func (p *PatternFilterService) FilterFilesByPatterns(files []string, baseDir string, patterns []string) []string {
	if len(patterns) == 0 {
		return files
	}

	var filtered []string
	for _, file := range files {
		var relativePath string

		if baseDir != "" {
			rel, err := p.pathUtils.GetRelativePath(file, baseDir)
			if err != nil {
				relativePath = filepath.Base(file)
			} else {
				relativePath = rel
			}
		} else {
			relativePath = filepath.Base(file)
		}

		if pattern_matcher.MatchesAnyPattern(relativePath, patterns) {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

