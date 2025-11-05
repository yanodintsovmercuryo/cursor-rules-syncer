package sync

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	mdcExtension    = ".mdc"
	headerSeparator = "---"
)

// cleanupExtraFiles removes files that exist in destination but not in source
func (s *SyncService) cleanupExtraFiles(srcFiles []string, srcBase, dstBase string, patterns interface{}) error {
	// Build map of source files by relative path
	srcFilesMap := make(map[string]bool)
	for _, srcFile := range srcFiles {
		relativePath, err := s.pathUtils.GetRelativePath(srcFile, srcBase)
		if err != nil {
			continue
		}
		srcFilesMap[relativePath] = true
	}

	// Найти все файлы назначения
	destFiles, err := s.fileOps.FindAllFiles(dstBase)
	if err != nil {
		return fmt.Errorf("error walking destination directory: %w", err)
	}

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

// copyFile копирует файл, опционально сохраняя заголовки
func (s *SyncService) copyFile(srcPath, dstPath string, preserveHeaders bool) error {
	// Полностью прочитать содержимое исходного файла
	srcContent, err := s.fileOps.ReadFileNormalized(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", srcPath, err)
	}

	var finalContent string

	if preserveHeaders {
		// Проверить существование файла назначения и извлечь заголовок
		existingHeader, err := s.extractExistingHeader(dstPath)
		if err != nil {
			return err
		}

		// Если есть существующий заголовок в назначении, извлечь содержимое из источника без заголовка
		if existingHeader != "" {
			srcContentWithoutHeader := s.removeHeaderFromContent(srcContent)
			finalContent = existingHeader + srcContentWithoutHeader
		} else {
			finalContent = srcContent
		}
	} else {
		// Просто перезаписать без сохранения заголовков
		finalContent = srcContent
	}

	// Записать финальное содержимое
	err = s.fileOps.WriteFile(dstPath, finalContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dstPath, err)
	}

	return nil
}

// removeHeaderFromContent removes the YAML header from markdown content
func (s *SyncService) removeHeaderFromContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	// Check if it starts with header separator
	if lines[0] != headerSeparator {
		return content // No header, return as is
	}

	// Look for closing separator
	for i := 1; i < len(lines); i++ {
		if lines[i] == headerSeparator {
			// Found closing separator, return content after it
			if i+1 < len(lines) {
				remainingLines := lines[i+1:]
				// Remove leading empty lines
				remainingLines = removeLeadingEmptyLines(remainingLines)
				return strings.Join(remainingLines, "\n")
			}
			return "" // Header takes up entire file
		}
		// Limit header search to reasonable number of lines
		if i > 20 {
			break
		}
	}

	// No closing separator found, return entire content
	return content
}

// extractHeaderFromContent extracts the YAML header from markdown content
func (s *SyncService) extractHeaderFromContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != headerSeparator {
		return "" // No header
	}

	var headerLines []string
	headerLines = append(headerLines, lines[0]) // Add first separator

	// Look for closing separator
	for i := 1; i < len(lines) && i <= 20; i++ { // Limit search to 20 lines
		headerLines = append(headerLines, lines[i])
		if lines[i] == headerSeparator {
			// Found closing separator, return header with newline
			return strings.Join(headerLines, "\n") + "\n"
		}
	}

	return "" // No proper header found
}

// removeLeadingEmptyLines removes leading empty lines from slice
func removeLeadingEmptyLines(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	return lines
}

// extractExistingHeader extracts the YAML header from an existing file
func (s *SyncService) extractExistingHeader(dstPath string) (string, error) {
	if exists, err := s.fileOps.FileExists(dstPath); err != nil || !exists {
		return "", nil // File doesn't exist, no header to preserve
	}

	content, err := s.fileOps.ReadFileNormalized(dstPath)
	if err != nil {
		return "", nil // Can't read file, continue without header
	}

	return s.extractHeaderFromContent(content), nil
}

// copyFileBasedOnExtension copies a file, applying header preservation only for .mdc files
func (s *SyncService) copyFileBasedOnExtension(srcPath, dstPath string, overwriteHeaders bool) error {
	// Only apply header logic for .mdc files
	if filepath.Ext(srcPath) == mdcExtension {
		return s.copyFile(srcPath, dstPath, !overwriteHeaders)
	}

	// For non-.mdc files, just copy directly without header processing
	return s.fileOps.CopyFile(srcPath, dstPath)
}

// filesAreEqualBasedOnExtension compares files, using header-aware comparison only for .mdc files
func (s *SyncService) filesAreEqualBasedOnExtension(file1, file2 string, overwriteHeaders bool) (bool, error) {
	// Only apply header logic for .mdc files
	if filepath.Ext(file1) == mdcExtension && filepath.Ext(file2) == mdcExtension {
		if overwriteHeaders {
			return s.filesAreEqualNormalized(file1, file2)
		} else {
			return s.filesAreEqualNormalizedWithoutHeaders(file1, file2)
		}
	}

	// For non-.mdc files, use simple comparison
	return s.filesAreEqualNormalized(file1, file2)
}

// filesAreEqualNormalized compares two files after normalizing their content
func (s *SyncService) filesAreEqualNormalized(file1, file2 string) (bool, error) {
	content1, err := s.fileOps.ReadFileNormalized(file1)
	if err != nil {
		return false, err
	}

	content2, err := s.fileOps.ReadFileNormalized(file2)
	if err != nil {
		return false, err
	}

	return content1 == content2, nil
}

// filesAreEqualNormalizedWithoutHeaders compares two files ignoring YAML headers
func (s *SyncService) filesAreEqualNormalizedWithoutHeaders(file1, file2 string) (bool, error) {
	content1, err := s.fileOps.ReadFileNormalized(file1)
	if err != nil {
		return false, err
	}

	content2, err := s.fileOps.ReadFileNormalized(file2)
	if err != nil {
		return false, err
	}

	// Remove headers from both files before comparison
	contentWithoutHeader1 := s.removeHeaderFromContent(content1)
	contentWithoutHeader2 := s.removeHeaderFromContent(content2)

	return contentWithoutHeader1 == contentWithoutHeader2, nil
}
