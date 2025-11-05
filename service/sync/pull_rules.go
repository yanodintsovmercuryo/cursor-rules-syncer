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
	rulesSourceDir, err := s.getRulesSourceDir(options.RulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules source dir: %w", err)
	}

	currentDir, err := s.fileOps.GetCurrentDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	gitRoot, err := s.gitOps.GetGitRootDir(currentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find git root: %w", err)
	}

	const (
		cursorDirName = ".cursor"
		rulesDirName  = "rules"
	)
	destRulesDir := filepath.Join(gitRoot, cursorDirName, rulesDirName)

	// Get file patterns for filtering
	filePatterns, err := s.fileService.GetFilePatterns(options.FilePatterns, cursorRulesPatternsEnvVar)
	if err != nil {
		return nil, fmt.Errorf("failed to get file patterns: %w", err)
	}

	if mkdirErr := s.fileOps.MkdirAll(destRulesDir, os.ModePerm); mkdirErr != nil {
		return nil, fmt.Errorf("failed to create destination directory %s: %w", destRulesDir, mkdirErr)
	}

	// Find source files with pattern filtering
	var sourceFiles []string
	if len(filePatterns) == 0 {
		// No patterns specified - get all files
		sourceFiles, err = s.fileOps.FindAllFiles(rulesSourceDir)
		if err != nil {
			return nil, fmt.Errorf("failed to find source files in %s: %w", rulesSourceDir, err)
		}
	} else {
		// Use pattern filtering
		sourceFiles, err = s.fileService.FindFilesByPatterns(rulesSourceDir, filePatterns)
		if err != nil {
			return nil, fmt.Errorf("failed to find files by patterns in %s: %w", rulesSourceDir, err)
		}
	}

	// Clean up extra files in destination that don't exist in source
	// Use pattern-aware cleanup
	effectivePatterns := string_utils.RemoveDuplicates(filePatterns)
	if len(effectivePatterns) == 0 {
		// No patterns - cleanup all extra files
		if err := s.cleanupExtraFiles(sourceFiles, rulesSourceDir, destRulesDir); err != nil {
			return nil, fmt.Errorf("failed to cleanup extra files: %w", err)
		}
	} else {
		// Use pattern-aware cleanup
		if err := s.fileService.CleanupExtraFilesByPatterns(sourceFiles, rulesSourceDir, destRulesDir, effectivePatterns); err != nil {
			return nil, fmt.Errorf("failed to cleanup extra files: %w", err)
		}
	}

	// Copy files with proper directory structure
	result := &models.SyncResult{
		Operations: []models.FileOperation{},
		HasChanges: false,
	}

	for _, srcFileFullPath := range sourceFiles {
		dstFileFullPath, err := s.pathUtils.RecreateDirectoryStructure(srcFileFullPath, rulesSourceDir, destRulesDir)
		if err != nil {
			s.output.PrintErrorf("Error recreating directory structure for %s: %v\n", srcFileFullPath, err)
			continue
		}

		// Get relative path for display
		relativePath, err := s.pathUtils.GetRelativePath(srcFileFullPath, rulesSourceDir)
		if err != nil {
			relativePath = s.pathUtils.GetBaseName(srcFileFullPath)
		}

		fileExistedBeforeCopy := true
		if _, err := s.fileOps.Stat(dstFileFullPath); os.IsNotExist(err) {
			fileExistedBeforeCopy = false
		} else if err != nil {
			s.output.PrintErrorf("Error checking destination file %s: %v\n", relativePath, err)
			continue
		}

		// Check if files are different before copying
		shouldCopy := true
		if fileExistedBeforeCopy {
			equal, err := s.fileService.AreEqual(srcFileFullPath, dstFileFullPath, options.OverwriteHeaders)
			if err != nil {
				s.output.PrintErrorf("Error comparing files %s: %v\n", relativePath, err)
				// Continue with copying in case of comparison error
			} else if equal {
				shouldCopy = false // Files are identical, no need to copy
			}
		}

		if shouldCopy {
			copyErr := s.fileService.Copy(srcFileFullPath, dstFileFullPath, options.OverwriteHeaders)
			if copyErr != nil {
				s.output.PrintErrorf("Error synchronizing file %s: %v\n", relativePath, copyErr)
			} else {
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
				result.HasChanges = true
			}
		}
	}

	return result, nil
}
