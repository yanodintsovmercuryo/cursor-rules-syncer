package path

import (
	"fmt"
	"os"
	"path/filepath"
)

// pathUtils структура для работы с путями
type PathUtils struct{}

// NewPathUtils создает новый экземпляр PathUtils
func NewPathUtils() *PathUtils {
	return &PathUtils{}
}

// RecreateDirectoryStructure воссоздает структуру директорий для файла назначения
func (p *PathUtils) RecreateDirectoryStructure(srcPath, srcBase, dstBase string) (string, error) {
	relativePath, err := filepath.Rel(srcBase, srcPath)
	if err != nil {
		return "", fmt.Errorf("cannot determine relative path: %w", err)
	}

	dstPath := filepath.Join(dstBase, relativePath)
	dstDir := filepath.Dir(dstPath)

	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("cannot create directory %s: %w", dstDir, err)
	}

	return dstPath, nil
}

// GetRelativePath возвращает относительный путь файла от базовой директории
func (p *PathUtils) GetRelativePath(filePath, baseDir string) (string, error) {
	return filepath.Rel(baseDir, filePath)
}

// NormalizePath нормализует путь для использования в кроссплатформенных путях
func (p *PathUtils) NormalizePath(filePath string) string {
	return filepath.ToSlash(filePath)
}

// GetDirectory возвращает директорию файла
func (p *PathUtils) GetDirectory(filePath string) string {
	return filepath.Dir(filePath)
}

// GetBaseName возвращает базовое имя файла (без директории)
func (p *PathUtils) GetBaseName(filePath string) string {
	return filepath.Base(filePath)
}
