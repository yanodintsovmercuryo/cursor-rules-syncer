package file_filter

import (
	"fmt"
	"os"
	"path/filepath"
)

// CleanupExtraFilesByPatterns удаляет файлы, которые существуют в назначении, но не в источнике, учитывая паттерны
func (s *FileFilterService) CleanupExtraFilesByPatterns(srcFiles []string, srcBase, dstBase string, patterns []string) error {
	// Построить карту исходных файлов по относительному пути (отфильтрованных по паттернам)
	srcFilesMap := make(map[string]bool)
	for _, srcFile := range srcFiles {
		relativePath, err := s.pathUtils.GetRelativePath(srcFile, srcBase)
		if err != nil {
			continue
		}
		srcFilesMap[relativePath] = true
	}

	// Найти все файлы назначения
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

	// Фильтровать файлы назначения по паттернам
	destFiles = s.filterFilesByPatterns(destFiles, dstBase, patterns)

	// Удалить файлы, которых нет в источнике
	for _, destFile := range destFiles {
		relativePath, err := s.pathUtils.GetRelativePath(destFile, dstBase)
		if err != nil {
			continue
		}

		if !srcFilesMap[relativePath] {
			if err := s.fileOps.RemoveFile(destFile); err != nil {
				s.outputService.PrintErrorf("Error deleting file %s: %v\n", relativePath, err)
			} else {
				s.outputService.PrintOperation("delete", relativePath)
			}
		}
	}

	return nil
}
