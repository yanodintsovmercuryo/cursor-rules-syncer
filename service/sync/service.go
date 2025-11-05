//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
package sync

import (
	"fmt"
	"os"
)

const cursorRulesPatternsEnvVar = "CURSOR_RULES_PATTERNS"

type outputService interface {
	PrintErrorf(format string, args ...interface{})
	PrintOperation(operationType, relativePath string)
	PrintOperationWithTarget(operationType, relativePath, target string)
}

type pathUtils interface {
	RecreateDirectoryStructure(srcPath, srcBase, dstBase string) (string, error)
	GetRelativePath(filePath, baseDir string) (string, error)
	GetBaseName(filePath string) string
}

type gitOps interface {
	GetGitRootDir(startDir string) (string, error)
	CommitChanges(repoDir, commitMessage string, withoutPush bool) error
}

type fileService interface {
	GetFilePatterns(flagValue, envVarName string) ([]string, error)
	FindFilesByPatterns(dir string, patterns []string) ([]string, error)
	CleanupExtraFilesByPatterns(srcFiles []string, srcBase, dstBase string, patterns []string) error
	AreEqual(file1, file2 string, overwriteHeaders bool) (bool, error)
	Copy(srcPath, dstPath string, overwriteHeaders bool) error
}

// fileOps defines interface for file operations used in sync
type fileOps interface {
	FindAllFiles(dir string) ([]string, error)
	GetCurrentDir() (string, error)
	MkdirAll(path string, perm os.FileMode) error
	Stat(filePath string) (os.FileInfo, error)
	FileExists(filePath string) (bool, error)
	RemoveFile(filePath string) error
}

// SyncService handles all sync operations
type SyncService struct {
	output      outputService
	fileOps     fileOps
	pathUtils   pathUtils
	gitOps      gitOps
	fileService fileService
}

// NewSyncService creates a new SyncService instance
func NewSyncService(output outputService, fileOps fileOps, pathUtils pathUtils, gitOps gitOps, fileService fileService) *SyncService {
	return &SyncService{
		output:      output,
		fileOps:     fileOps,
		pathUtils:   pathUtils,
		gitOps:      gitOps,
		fileService: fileService,
	}
}

// NewSyncServiceWithMocks creates a new SyncService with provided mocks for testing
func NewSyncServiceWithMocks(output outputService, fileOps fileOps, pathUtils pathUtils, gitOps gitOps, fileService fileService) *SyncService {
	return &SyncService{
		output:      output,
		fileOps:     fileOps,
		pathUtils:   pathUtils,
		gitOps:      gitOps,
		fileService: fileService,
	}
}

// getRulesSourceDir gets rules directory path from flag or environment variable
func (s *SyncService) getRulesSourceDir(flagValue string) (string, error) {
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

// cleanupExtraFiles removes files that exist in destination but not in source
func (s *SyncService) cleanupExtraFiles(srcFiles []string, srcBase, dstBase string) error {
	srcFilesMap := make(map[string]bool)
	for _, srcFile := range srcFiles {
		relativePath, err := s.pathUtils.GetRelativePath(srcFile, srcBase)
		if err != nil {
			continue
		}
		srcFilesMap[relativePath] = true
	}

	destFiles, err := s.fileOps.FindAllFiles(dstBase)
	if err != nil {
		return fmt.Errorf("error walking destination directory: %w", err)
	}

	for _, destFile := range destFiles {
		relativePath, err := s.pathUtils.GetRelativePath(destFile, dstBase)
		if err != nil {
			continue
		}

		if !srcFilesMap[relativePath] {
			if err := s.fileOps.RemoveFile(destFile); err != nil {
				s.output.PrintErrorf("Error deleting file %s: %v\n", relativePath, err)
			} else {
				s.output.PrintOperation("delete", relativePath)
			}
		}
	}

	return nil
}
