package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yanodintsovmercuryo/cursor-rules-syncer/models"
)

const cursorRulesPatternsEnvVar = "CURSOR_RULES_PATTERNS"

// SyncService handles all sync operations
type SyncService struct {
	outputService     *OutputService
	fileFilterService *FileFilterService
}

// NewSyncService creates a new SyncService
func NewSyncService(outputService *OutputService) *SyncService {
	return &SyncService{
		outputService:     outputService,
		fileFilterService: NewFileFilterService(outputService),
	}
}

// PullRules pulls rules from source directory to project .cursor/rules directory
func (s *SyncService) PullRules(options *models.SyncOptions) (*models.SyncResult, error) {
	rulesSourceDir, err := s.GetRulesSourceDir(options.RulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules source dir: %w", err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	gitRoot, err := s.getGitRootDir(currentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find git root: %w", err)
	}

	destRulesDir := filepath.Join(gitRoot, cursorDirName, rulesDirName)

	// Get file patterns for filtering
	filePatterns, err := s.fileFilterService.GetFilePatterns(options.FilePatterns, cursorRulesPatternsEnvVar)
	if err != nil {
		return nil, fmt.Errorf("failed to get file patterns: %w", err)
	}

	if mkdirErr := os.MkdirAll(destRulesDir, os.ModePerm); mkdirErr != nil {
		return nil, fmt.Errorf("failed to create destination directory %s: %w", destRulesDir, mkdirErr)
	}

	// Find source files with pattern filtering
	var sourceFiles []string
	if len(filePatterns) == 0 {
		// No patterns specified - get all files
		sourceFiles, err = s.findAllFiles(rulesSourceDir)
		if err != nil {
			return nil, fmt.Errorf("failed to find source files in %s: %w", rulesSourceDir, err)
		}
	} else {
		// Use pattern filtering
		sourceFiles, err = s.fileFilterService.FindFilesByPatterns(rulesSourceDir, filePatterns)
		if err != nil {
			return nil, fmt.Errorf("failed to find files by patterns in %s: %w", rulesSourceDir, err)
		}
	}

	// Clean up extra files in destination that don't exist in source
	// Use pattern-aware cleanup
	effectivePatterns := s.fileFilterService.GetEffectivePatterns(filePatterns)
	if len(effectivePatterns) == 0 {
		// No patterns - cleanup all extra files
		if err := s.cleanupExtraFiles(sourceFiles, rulesSourceDir, destRulesDir, nil); err != nil {
			return nil, fmt.Errorf("failed to cleanup extra files: %w", err)
		}
	} else {
		// Use pattern-aware cleanup
		if err := s.fileFilterService.CleanupExtraFilesByPatterns(sourceFiles, rulesSourceDir, destRulesDir, effectivePatterns); err != nil {
			return nil, fmt.Errorf("failed to cleanup extra files: %w", err)
		}
	}

	// Copy files with proper directory structure
	result := &models.SyncResult{
		Operations: []models.FileOperation{},
		HasChanges: false,
	}

	for _, srcFileFullPath := range sourceFiles {
		dstFileFullPath, err := s.RecreateDirectoryStructure(srcFileFullPath, rulesSourceDir, destRulesDir)
		if err != nil {
			s.outputService.PrintErrorf("Error recreating directory structure for %s: %v\n", srcFileFullPath, err)
			continue
		}

		// Get relative path for display
		relativePath, err := s.GetRelativePath(srcFileFullPath, rulesSourceDir)
		if err != nil {
			relativePath = filepath.Base(srcFileFullPath)
		}

		fileExistedBeforeCopy := true
		if _, err := os.Stat(dstFileFullPath); os.IsNotExist(err) {
			fileExistedBeforeCopy = false
		} else if err != nil {
			s.outputService.PrintErrorf("Error checking destination file %s: %v\n", relativePath, err)
			continue
		}

		// Check if files are different before copying
		shouldCopy := true
		if fileExistedBeforeCopy {
			equal, err := s.filesAreEqualBasedOnExtension(srcFileFullPath, dstFileFullPath, options.OverwriteHeaders)
			if err != nil {
				s.outputService.PrintErrorf("Error comparing files %s: %v\n", relativePath, err)
				// Continue with copying in case of comparison error
			} else if equal {
				shouldCopy = false // Files are identical, no need to copy
			}
		}

		if shouldCopy {
			copyErr := s.copyFileBasedOnExtension(srcFileFullPath, dstFileFullPath, options.OverwriteHeaders)
			if copyErr != nil {
				s.outputService.PrintErrorf("Error synchronizing file %s: %v\n", relativePath, copyErr)
			} else {
				operationType := "update"
				if !fileExistedBeforeCopy {
					operationType = "add"
				}
				s.outputService.PrintOperation(operationType, relativePath)

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

// PushRules pushes rules from project .cursor/rules directory to source directory
func (s *SyncService) PushRules(options *models.SyncOptions) (*models.SyncResult, error) {
	rulesEnvDir, err := s.GetRulesSourceDir(options.RulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules source dir: %w", err)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	projectGitRoot, err := s.getGitRootDir(currentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find git root for project: %w", err)
	}

	rulesSourceDirInProject := filepath.Join(projectGitRoot, cursorDirName, rulesDirName)

	if _, statErr := os.Stat(rulesSourceDirInProject); os.IsNotExist(statErr) {
		return nil, fmt.Errorf("project rules directory %s not found. Nothing to push", rulesSourceDirInProject)
	}

	// Get file patterns for filtering
	filePatterns, err := s.fileFilterService.GetFilePatterns(options.FilePatterns, cursorRulesPatternsEnvVar)
	if err != nil {
		return nil, fmt.Errorf("failed to get file patterns: %w", err)
	}

	// Find project files with pattern filtering
	var projectFiles []string
	if len(filePatterns) == 0 {
		// No patterns specified - get all files
		projectFiles, err = s.findAllFiles(rulesSourceDirInProject)
		if err != nil {
			return nil, fmt.Errorf("failed to find files in project rules directory %s: %w", rulesSourceDirInProject, err)
		}
	} else {
		// Use pattern filtering
		projectFiles, err = s.fileFilterService.FindFilesByPatterns(rulesSourceDirInProject, filePatterns)
		if err != nil {
			return nil, fmt.Errorf("failed to find files by patterns in project rules directory %s: %w", rulesSourceDirInProject, err)
		}
	}

	if mkdirErr := os.MkdirAll(rulesEnvDir, os.ModePerm); mkdirErr != nil {
		return nil, fmt.Errorf("failed to create destination directory %s: %w", rulesEnvDir, mkdirErr)
	}

	// Clean up extra files in destination that don't exist in source
	// Use pattern-aware cleanup
	effectivePatterns := s.fileFilterService.GetEffectivePatterns(filePatterns)
	if len(effectivePatterns) == 0 {
		// No patterns - cleanup all extra files
		if err := s.cleanupExtraFiles(projectFiles, rulesSourceDirInProject, rulesEnvDir, nil); err != nil {
			return nil, fmt.Errorf("failed to cleanup extra files: %w", err)
		}
	} else {
		// Use pattern-aware cleanup
		if err := s.fileFilterService.CleanupExtraFilesByPatterns(projectFiles, rulesSourceDirInProject, rulesEnvDir, effectivePatterns); err != nil {
			return nil, fmt.Errorf("failed to cleanup extra files: %w", err)
		}
	}

	result := &models.SyncResult{
		Operations: []models.FileOperation{},
		HasChanges: false,
	}

	if len(projectFiles) == 0 {
		// No files to process, but we might have deleted some files above
		return result, nil
	}

	// Copy files with proper directory structure
	for _, srcFileFullPath := range projectFiles {
		dstFileFullPath, err := s.RecreateDirectoryStructure(srcFileFullPath, rulesSourceDirInProject, rulesEnvDir)
		if err != nil {
			s.outputService.PrintErrorf("Error recreating directory structure for %s: %v\n", srcFileFullPath, err)
			continue
		}

		// Get relative path for display
		relativePath, err := s.GetRelativePath(srcFileFullPath, rulesSourceDirInProject)
		if err != nil {
			relativePath = filepath.Base(srcFileFullPath)
		}

		fileExists := true
		if _, statErr := os.Stat(dstFileFullPath); os.IsNotExist(statErr) {
			fileExists = false
		} else if statErr != nil {
			s.outputService.PrintErrorf("Error checking destination file %s in %s: %v\n", relativePath, rulesEnvDir, statErr)
			continue
		}

		// Check if files are different before copying
		shouldCopy := true
		if fileExists {
			equal, err := s.filesAreEqualBasedOnExtension(srcFileFullPath, dstFileFullPath, options.OverwriteHeaders)
			if err != nil {
				s.outputService.PrintErrorf("Error comparing files %s: %v\n", relativePath, err)
				// Continue with copying in case of comparison error
			} else if equal {
				shouldCopy = false // Files are identical, no need to copy
			}
		}

		if shouldCopy {
			copyErr := s.copyFileBasedOnExtension(srcFileFullPath, dstFileFullPath, options.OverwriteHeaders)
			if copyErr != nil {
				s.outputService.PrintErrorf("Error synchronizing file %s to %s: %v\n", relativePath, rulesEnvDir, copyErr)
			} else {
				operationType := "update"
				if !fileExists {
					operationType = "add"
				}
				s.outputService.PrintOperationWithTarget(operationType, relativePath, filepath.Base(rulesEnvDir))

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

	// Only commit if we have changes
	if result.HasChanges {
		if err := s.commitChanges(rulesEnvDir, "Sync cursor rules: updated from project "+filepath.Base(projectGitRoot), options.GitWithoutPush); err != nil {
			s.outputService.PrintErrorf("Commit failed for %s: %v\n", rulesEnvDir, err)
		}
	}

	return result, nil
}
