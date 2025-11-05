//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
package sync

import (
	"fmt"
	"os"
)

const cursorRulesPatternsEnvVar = "CURSOR_RULES_PATTERNS"

// Приватные интерфейсы - отражения публичных сервисов для инкапсуляции
type outputService interface {
	PrintErrorf(format string, args ...interface{})
	PrintOperation(operationType, relativePath string)
	PrintOperationWithTarget(operationType, relativePath, target string)
}

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

type gitOps interface {
	GetGitRootDir(startDir string) (string, error)
	CommitChanges(repoDir, commitMessage string, withoutPush bool) error
}

type fileFilter interface {
	GetFilePatterns(flagValue, envVarName string) ([]string, error)
	FindFilesByPatterns(dir string, patterns []string) ([]string, error)
	CleanupExtraFilesByPatterns(srcFiles []string, srcBase, dstBase string, patterns []string) error
	GetEffectivePatterns(patterns []string) []string
}

// SyncService обрабатывает все sync операции
type SyncService struct {
	outputService     outputService
	fileFilterService fileFilter
	fileOps           fileOps
	pathUtils         pathUtils
	gitOps            gitOps
}

// NewSyncService создает новый SyncService
func NewSyncService(outputService outputService, fileFilterService fileFilter, fileOps fileOps, pathUtils pathUtils, gitOps gitOps) *SyncService {
	return &SyncService{
		outputService:     outputService,
		fileFilterService: fileFilterService,
		fileOps:           fileOps,
		pathUtils:         pathUtils,
		gitOps:            gitOps,
	}
}

// GetRulesSourceDir получает путь к директории правил из флага или переменной окружения
func (s *SyncService) GetRulesSourceDir(flagValue string) (string, error) {
	const cursorRulesDirEnvVar = "CURSOR_RULES_DIR"

	if flagValue != "" {
		return flagValue, nil
	}

	rulesDir := os.Getenv(cursorRulesDirEnvVar)
	if rulesDir == "" {
		return "", fmt.Errorf("rules directory not specified: use --rules-dir flag or set %s environment variable", cursorRulesDirEnvVar)
	}
	return rulesDir, nil
}
