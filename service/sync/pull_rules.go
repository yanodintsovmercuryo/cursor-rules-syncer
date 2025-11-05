package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yanodintsovmercuryo/cursync/models"
	"github.com/yanodintsovmercuryo/cursync/pkg/string_utils"
)

// PullRules pulls rules from source directory to project .cursor/rules directory
func (s *SyncService) PullRules(options *models.SyncOptions) (*models.SyncResult, error) {
	rulesSourceDir, destRulesDir, err := s.preparePullPaths(options.RulesDir)
	if err != nil {
		return nil, err
	}

	filePatterns, err := s.fileService.GetFilePatterns(options.FilePatterns, cursorRulesPatternsEnvVar)
	if err != nil {
		return nil, fmt.Errorf("failed to get file patterns: %w", err)
	}

	if mkdirErr := s.fileOps.MkdirAll(destRulesDir, os.ModePerm); mkdirErr != nil {
		return nil, fmt.Errorf("failed to create destination directory %s: %w", destRulesDir, mkdirErr)
	}

	sourceFiles, err := s.findFilesWithPatterns(rulesSourceDir, filePatterns)
	if err != nil {
		return nil, err
	}

	if err := s.cleanupExtraFilesWithPatterns(sourceFiles, rulesSourceDir, destRulesDir, filePatterns); err != nil {
		return nil, err
	}

	return s.copyFiles(sourceFiles, rulesSourceDir, destRulesDir, options.OverwriteHeaders), nil
}

// preparePullPaths prepares source and destination paths for pull operation
func (s *SyncService) preparePullPaths(rulesDir string) (string, string, error) {
	rulesSourceDir, err := s.getRulesSourceDir(rulesDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get rules source dir: %w", err)
	}

	currentDir, err := s.fileOps.GetCurrentDir()
	if err != nil {
		return "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	gitRoot, err := s.gitOps.GetGitRootDir(currentDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to find git root: %w", err)
	}

	const (
		cursorDirName = ".cursor"
		rulesDirName  = "rules"
	)
	destRulesDir := filepath.Join(gitRoot, cursorDirName, rulesDirName)

	return rulesSourceDir, destRulesDir, nil
}

// findFilesWithPatterns finds files using patterns or returns all files if patterns are empty
func (s *SyncService) findFilesWithPatterns(dir string, patterns []string) ([]string, error) {
	if len(patterns) == 0 {
		files, err := s.fileOps.FindAllFiles(dir)
		if err != nil {
			return nil, fmt.Errorf("failed to find source files in %s: %w", dir, err)
		}
		return files, nil
	}

	files, err := s.fileService.FindFilesByPatterns(dir, patterns)
	if err != nil {
		return nil, fmt.Errorf("failed to find files by patterns in %s: %w", dir, err)
	}
	return files, nil
}

// cleanupExtraFilesWithPatterns cleans up extra files using pattern-aware or simple cleanup
func (s *SyncService) cleanupExtraFilesWithPatterns(sourceFiles []string, srcBase, dstBase string, patterns []string) error {
	effectivePatterns := string_utils.RemoveDuplicates(patterns)
	if len(effectivePatterns) == 0 {
		if err := s.cleanupExtraFiles(sourceFiles, srcBase, dstBase); err != nil {
			return fmt.Errorf("failed to cleanup extra files: %w", err)
		}
		return nil
	}

	if err := s.fileService.CleanupExtraFilesByPatterns(sourceFiles, srcBase, dstBase, effectivePatterns); err != nil {
		return fmt.Errorf("failed to cleanup extra files: %w", err)
	}
	return nil
}

// copyFiles copies files from source to destination with proper directory structure
func (s *SyncService) copyFiles(sourceFiles []string, srcBase, dstBase string, overwriteHeaders bool) *models.SyncResult {
	result := &models.SyncResult{
		Operations: []models.FileOperation{},
		HasChanges: false,
	}

	for _, srcFileFullPath := range sourceFiles {
		dstFileFullPath, err := s.pathUtils.RecreateDirectoryStructure(srcFileFullPath, srcBase, dstBase)
		if err != nil {
			s.output.PrintErrorf("Error recreating directory structure for %s: %v\n", srcFileFullPath, err)
			continue
		}

		relativePath, err := s.pathUtils.GetRelativePath(srcFileFullPath, srcBase)
		if err != nil {
			relativePath = s.pathUtils.GetBaseName(srcFileFullPath)
		}

		fileExistedBeforeCopy, err := s.checkFileExists(dstFileFullPath, relativePath)
		if err != nil {
			continue
		}

		shouldCopy := s.shouldCopyFile(srcFileFullPath, dstFileFullPath, fileExistedBeforeCopy, overwriteHeaders, relativePath)
		if !shouldCopy {
			continue
		}

		if s.copySingleFile(srcFileFullPath, dstFileFullPath, relativePath, fileExistedBeforeCopy, overwriteHeaders, result) {
			result.HasChanges = true
		}
	}

	return result
}

// checkFileExists checks if destination file exists
func (s *SyncService) checkFileExists(dstFileFullPath, relativePath string) (bool, error) {
	if _, err := s.fileOps.Stat(dstFileFullPath); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		s.output.PrintErrorf("Error checking destination file %s: %v\n", relativePath, err)
		return false, err
	}
	return true, nil
}

// shouldCopyFile determines if file should be copied based on comparison
func (s *SyncService) shouldCopyFile(srcFileFullPath, dstFileFullPath string, fileExistedBeforeCopy, overwriteHeaders bool, relativePath string) bool {
	if !fileExistedBeforeCopy {
		return true
	}

	equal, err := s.fileService.AreEqual(srcFileFullPath, dstFileFullPath, overwriteHeaders)
	if err != nil {
		s.output.PrintErrorf("Error comparing files %s: %v\n", relativePath, err)
		return true
	}

	return !equal
}

// copySingleFile copies a single file and updates result
func (s *SyncService) copySingleFile(srcFileFullPath, dstFileFullPath, relativePath string, fileExistedBeforeCopy, overwriteHeaders bool, result *models.SyncResult) bool {
	copyErr := s.fileService.Copy(srcFileFullPath, dstFileFullPath, overwriteHeaders)
	if copyErr != nil {
		s.output.PrintErrorf("Error synchronizing file %s: %v\n", relativePath, copyErr)
		return false
	}

	operationType := "update"
	if !fileExistedBeforeCopy {
		operationType = "add"
	}
	s.output.PrintOperation(operationType, relativePath)

	result.Operations = append(result.Operations, models.FileOperation{
		Type:         models.OperationType(operationType),
		SourcePath:   srcFileFullPath,
		TargetPath:   dstFileFullPath,
		RelativePath: relativePath,
	})

	return true
}
