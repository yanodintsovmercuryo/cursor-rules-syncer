package file_filter

import (
	"fmt"
	"path/filepath"

	"github.com/yanodintsovmercuryo/cursor-rules-syncer/pkg/string_utils"
)

// GetEffectivePatterns возвращает эффективные паттерны, пустой срез означает отсутствие фильтрации
func (s *FileFilterService) GetEffectivePatterns(patterns []string) []string {
	if len(patterns) == 0 {
		return []string{} // Без фильтрации - синхронизировать все файлы
	}

	// Удалить дубликаты сохраняя порядок
	return string_utils.RemoveDuplicates(patterns)
}

// ValidatePatterns проверяет правильность паттернов
func (s *FileFilterService) ValidatePatterns(patterns []string) error {
	for _, pattern := range patterns {
		// Проверить работу паттерна с filepath.Match
		if _, err := filepath.Match(pattern, "test"); err != nil {
			return fmt.Errorf("invalid pattern '%s': %w", pattern, err)
		}
	}
	return nil
}
