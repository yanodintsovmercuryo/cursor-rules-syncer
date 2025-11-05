package file_ops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileOps определяет интерфейс для файловых операций
type FileOps interface {
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

// fileOps структура для файловых операций
type fileOps struct{}

// NewFileOps создает новый экземпляр FileOps
func NewFileOps() FileOps {
	return &fileOps{}
}

// FindAllFiles находит все файлы в указанной директории рекурсивно
func (f *fileOps) FindAllFiles(dir string) ([]string, error) {
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
		return nil, fmt.Errorf("error finding files in %s: %w", dir, err)
	}
	return allFiles, nil
}

// ReadFileNormalized читает файл и нормализует переносы строк
func (f *fileOps) ReadFileNormalized(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Normalize line endings - convert to LF
	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	// Remove trailing whitespace except newlines, then normalize newlines at the end
	normalized = strings.TrimRight(normalized, " \t")
	normalized = strings.TrimRight(normalized, "\n")

	// Add one newline at the end if file is not empty
	if len(normalized) > 0 {
		normalized += "\n"
	}

	return normalized, nil
}

// WriteFile создает директорию если нужно и записывает содержимое в файл
func (f *fileOps) WriteFile(filePath, content string, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filePath), perm); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
	}

	err := os.WriteFile(filePath, []byte(content), perm)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

// FileExists проверяет существование файла
func (f *fileOps) FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CopyFile копирует файл из источника в назначение
func (f *fileOps) CopyFile(srcPath, dstPath string) error {
	content, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", srcPath, err)
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dstPath, err)
	}

	err = os.WriteFile(dstPath, content, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dstPath, err)
	}

	return nil
}

// RemoveFile удаляет файл
func (f *fileOps) RemoveFile(filePath string) error {
	return os.Remove(filePath)
}

// MkdirAll создает директорию со всеми необходимыми родительскими директориями
func (f *fileOps) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// GetCurrentDir возвращает текущую рабочую директорию
func (f *fileOps) GetCurrentDir() (string, error) {
	return os.Getwd()
}

// Stat возвращает информацию о файле
func (f *fileOps) Stat(filePath string) (os.FileInfo, error) {
	return os.Stat(filePath)
}
