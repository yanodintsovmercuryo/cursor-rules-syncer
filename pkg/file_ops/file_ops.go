//go:generate mockgen -source=file_ops.go -destination=mocks/mocks.go -package=mocks
package file_ops

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileOps handles file operations
type FileOps struct{}

// NewFileOps creates a new FileOps instance
func NewFileOps() *FileOps {
	return &FileOps{}
}

// FindAllFiles finds all files in the specified directory recursively
func (f *FileOps) FindAllFiles(dir string) ([]string, error) {
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

// ReadFileNormalized reads a file and normalizes line endings
func (f *FileOps) ReadFileNormalized(filePath string) (string, error) {
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

// WriteFile creates directory if needed and writes content to file
func (f *FileOps) WriteFile(filePath, content string, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(filePath), perm); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", filePath, err)
	}

	err := os.WriteFile(filePath, []byte(content), perm)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %w", filePath, err)
	}
	return nil
}

// FileExists checks if file exists
func (f *FileOps) FileExists(filePath string) (bool, error) {
	_, err := os.Stat(filePath)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CopyFile copies file from source to destination
func (f *FileOps) CopyFile(srcPath, dstPath string) error {
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

// RemoveFile removes a file
func (f *FileOps) RemoveFile(filePath string) error {
	return os.Remove(filePath)
}

// MkdirAll creates directory with all necessary parent directories
func (f *FileOps) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// GetCurrentDir returns current working directory
func (f *FileOps) GetCurrentDir() (string, error) {
	return os.Getwd()
}

// Stat returns file information
func (f *FileOps) Stat(filePath string) (os.FileInfo, error) {
	return os.Stat(filePath)
}
