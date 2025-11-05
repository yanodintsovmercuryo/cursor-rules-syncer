//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
package file_filter

import (
	"os"
	"path/filepath"
)

type outputService interface {
	PrintErrorf(format string, args ...interface{})
	PrintOperation(operationType, relativePath string)
}

// Приватные интерфейсы - отражения публичных сервисов для инкапсуляции
type fileOps interface {
	FindAllFiles(dir string) ([]string, error)
	ReadFileNormalized(filePath string) (string, error)
	WriteFile(filePath, content string, perm os.FileMode) error
	FileExists(filePath string) (bool, error)
	CopyFile(srcPath, dstPath string) error
	RemoveFile(filePath string) error
	MkdirAll(path string, perm os.FileMode) error
	GetCurrentDir() (string, error)
	Stat(filePath string) (os.FileInfo, error)
}

type pathUtils interface {
	RecreateDirectoryStructure(srcPath, srcBase, dstBase string) (string, error)
	GetRelativePath(filePath, baseDir string) (string, error)
	NormalizePath(filePath string) string
	GetDirectory(filePath string) string
	GetBaseName(filePath string) string
}

// FileFilterService обрабатывает сопоставление и фильтрацию файлов
type FileFilterService struct {
	outputService outputService
	fileOps       fileOps
	pathUtils     pathUtils
}

// NewFileFilterService создает новый FileFilterService
func NewFileFilterService(outputService outputService, fileOps fileOps, pathUtils pathUtils) *FileFilterService {
	return &FileFilterService{
		outputService: outputService,
		fileOps:       fileOps,
		pathUtils:     pathUtils,
	}
}

// filterFilesByPatterns фильтрует файлы на основе предоставленных паттернов
func (s *FileFilterService) filterFilesByPatterns(files []string, baseDir string, patterns []string) []string {
	if len(patterns) == 0 {
		return files
	}

	var filtered []string
	for _, file := range files {
		var relativePath string

		if baseDir != "" {
			// Попытаться получить относительный путь
			rel, err := s.pathUtils.GetRelativePath(file, baseDir)
			if err != nil {
				// Если не удается получить относительный путь, использовать базовое имя
				relativePath = filepath.Base(file)
			} else {
				relativePath = rel
			}
		} else {
			// Базовая директория не указана, использовать базовое имя
			relativePath = filepath.Base(file)
		}

		if s.matchesAnyPattern(relativePath, patterns) {
			filtered = append(filtered, file)
		}
	}

	return filtered
}

// matchesAnyPattern проверяет, соответствует ли путь к файлу любому из предоставленных паттернов
func (s *FileFilterService) matchesAnyPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		if s.matchesPattern(filePath, pattern) {
			return true
		}
	}
	return false
}

func (s *FileFilterService) matchesPattern(filePath, pattern string) bool {
	normalizedPath := filepath.ToSlash(filePath)
	normalizedPattern := filepath.ToSlash(pattern)

	if match, err := filepath.Match(normalizedPattern, normalizedPath); err == nil && match {
		return true
	}

	filename := filepath.Base(normalizedPath)
	if match, err := filepath.Match(normalizedPattern, filename); err == nil && match {
		return true
	}

	dir := filepath.Dir(normalizedPath)
	for dir != "." && dir != "/" {
		if match, err := filepath.Match(normalizedPattern, dir); err == nil && match {
			return true
		}
		dir = filepath.Dir(dir)
	}

	return false
}
