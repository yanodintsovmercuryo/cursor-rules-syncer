package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yanodintsovmercuryo/cursor-rules-syncer/models"
)

// PushRules pushes rules from project .cursor/rules directory to source directory
func (s *SyncService) PushRules(options *models.SyncOptions) (*models.SyncResult, error) {
	rulesEnvDir, err := s.GetRulesSourceDir(options.RulesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules source dir: %w", err)
	}

	currentDir, err := s.fileOps.GetCurrentDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	projectGitRoot, err := s.gitOps.GetGitRootDir(currentDir)
	if err != nil {
		return nil, fmt.Errorf("failed to find git root for project: %w", err)
	}

	const (
		cursorDirName = ".cursor"
		rulesDirName  = "rules"
	)
	rulesSourceDirInProject := filepath.Join(projectGitRoot, cursorDirName, rulesDirName)

	if exists, _ := s.fileOps.FileExists(rulesSourceDirInProject); !exists {
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

	if mkdirErr := s.fileOps.MkdirAll(rulesEnvDir, os.ModePerm); mkdirErr != nil {
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

	// Копировать файлы с правильной структурой директорий
	for _, srcFileFullPath := range projectFiles {
		dstFileFullPath, err := s.pathUtils.RecreateDirectoryStructure(srcFileFullPath, rulesSourceDirInProject, rulesEnvDir)
		if err != nil {
			s.outputService.PrintErrorf("Error recreating directory structure for %s: %v\n", srcFileFullPath, err)
			continue
		}

		// Получить относительный путь для отображения
		relativePath, err := s.pathUtils.GetRelativePath(srcFileFullPath, rulesSourceDirInProject)
		if err != nil {
			relativePath = s.pathUtils.GetBaseName(srcFileFullPath)
		}

		fileExists := true
		if exists, _ := s.fileOps.FileExists(dstFileFullPath); !exists {
			fileExists = false
		} else if err != nil {
			s.outputService.PrintErrorf("Error checking destination file %s in %s: %v\n", relativePath, rulesEnvDir, err)
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
				s.outputService.PrintOperationWithTarget(operationType, relativePath, s.pathUtils.GetBaseName(rulesEnvDir))

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

	// Коммитить только если есть изменения
	if result.HasChanges {
		if err := s.gitOps.CommitChanges(rulesEnvDir, "Sync cursor rules: updated from project "+s.pathUtils.GetBaseName(projectGitRoot), options.GitWithoutPush); err != nil {
			s.outputService.PrintErrorf("Commit failed for %s: %v\n", rulesEnvDir, err)
		}
	}

	return result, nil
}
