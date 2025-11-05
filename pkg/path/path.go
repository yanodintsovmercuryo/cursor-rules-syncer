package path

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// PathUtils handles path operations
type PathUtils struct{}

// NewPathUtils creates a new PathUtils instance
func NewPathUtils() *PathUtils {
	return &PathUtils{}
}

// RecreateDirectoryStructure recreates directory structure for destination file
func (p *PathUtils) RecreateDirectoryStructure(srcPath, srcBase, dstBase string) (string, error) {
	relativePath, err := p.GetRelativePath(srcPath, srcBase)
	if err != nil {
		return "", err
	}

	dstPath := filepath.Join(dstBase, relativePath)
	dstDir := filepath.Dir(dstPath)

	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("cannot create directory %s: %w", dstDir, err)
	}

	return dstPath, nil
}

// GetRelativePath returns relative path of file from base directory
func (p *PathUtils) GetRelativePath(filePath, baseDir string) (string, error) {
	rel, err := filepath.Rel(baseDir, filePath)
	if err != nil {
		return "", fmt.Errorf("cannot determine relative path: %w", err)
	}

	// Check that path is truly relative (doesn't contain ".." at the beginning)
	if len(rel) >= 3 && rel[:3] == ".."+string(filepath.Separator) {
		return "", fmt.Errorf("file path %s is not within base directory %s", filePath, baseDir)
	}

	return rel, nil
}

// NormalizePath normalizes path for cross-platform use
func (p *PathUtils) NormalizePath(filePath string) string {
	normalized := filepath.ToSlash(filePath)
	// Additional normalization to ensure backslash conversion
	normalized = strings.ReplaceAll(normalized, "\\", "/")
	return normalized
}

// GetDirectory returns file directory
func (p *PathUtils) GetDirectory(filePath string) string {
	return filepath.Dir(filePath)
}

// GetBaseName returns base file name (without directory)
func (p *PathUtils) GetBaseName(filePath string) string {
	return filepath.Base(filePath)
}
