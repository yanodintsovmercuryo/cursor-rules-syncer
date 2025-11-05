package sync

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/yanodintsovmercuryo/cursync/models"
)

// PushRules pushes rules from project .cursor/rules directory to source directory
func (s *SyncService) PushRules(options *models.SyncOptions) (*models.SyncResult, error) {
	rulesEnvDir, rulesSourceDirInProject, projectGitRoot, err := s.preparePushPaths(options.RulesDir)
	if err != nil {
		return nil, err
	}

	if validateErr := s.validateRulesDirectory(rulesSourceDirInProject); validateErr != nil {
		return nil, validateErr
	}

	filePatterns, err := s.fileService.GetFilePatterns(options.FilePatterns, cursorRulesPatternsEnvVar)
	if err != nil {
		return nil, fmt.Errorf("failed to get file patterns: %w", err)
	}

	projectFiles, err := s.findFilesWithPatterns(rulesSourceDirInProject, filePatterns)
	if err != nil {
		return nil, err
	}

	if mkdirErr := s.fileOps.MkdirAll(rulesEnvDir, os.ModePerm); mkdirErr != nil {
		return nil, fmt.Errorf("failed to create destination directory %s: %w", rulesEnvDir, mkdirErr)
	}

	if err := s.cleanupExtraFilesWithPatterns(projectFiles, rulesSourceDirInProject, rulesEnvDir, filePatterns); err != nil {
		return nil, err
	}

	result := s.copyFilesForPush(projectFiles, rulesSourceDirInProject, rulesEnvDir, options.OverwriteHeaders)

	if result.HasChanges {
		if err := s.gitOps.CommitChanges(rulesEnvDir, "Sync cursor rules: updated from project "+s.pathUtils.GetBaseName(projectGitRoot), options.GitWithoutPush); err != nil {
			s.output.PrintErrorf("Commit failed for %s: %v\n", rulesEnvDir, err)
		}
	}

	return result, nil
}

// preparePushPaths prepares paths for push operation
func (s *SyncService) preparePushPaths(rulesDir string) (string, string, string, error) {
	rulesEnvDir, err := s.getRulesSourceDir(rulesDir)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get rules source dir: %w", err)
	}

	currentDir, err := s.fileOps.GetCurrentDir()
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get current directory: %w", err)
	}

	projectGitRoot, err := s.gitOps.GetGitRootDir(currentDir)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to find git root for project: %w", err)
	}

	const (
		cursorDirName = ".cursor"
		rulesDirName  = "rules"
	)
	rulesSourceDirInProject := filepath.Join(projectGitRoot, cursorDirName, rulesDirName)

	return rulesEnvDir, rulesSourceDirInProject, projectGitRoot, nil
}

// validateRulesDirectory checks if rules directory exists in project
func (s *SyncService) validateRulesDirectory(rulesSourceDirInProject string) error {
	exists, err := s.fileOps.FileExists(rulesSourceDirInProject)
	if err != nil {
		return fmt.Errorf("failed to check if rules directory exists: %w", err)
	}
	if !exists {
		return fmt.Errorf("project rules directory %s not found. Nothing to push", rulesSourceDirInProject)
	}
	return nil
}

// copyFilesForPush copies files from project to source directory
func (s *SyncService) copyFilesForPush(projectFiles []string, srcBase, dstBase string, overwriteHeaders bool) *models.SyncResult {
	result := &models.SyncResult{
		Operations: []models.FileOperation{},
		HasChanges: false,
	}

	if len(projectFiles) == 0 {
		return result
	}

	for _, srcFileFullPath := range projectFiles {
		dstFileFullPath, err := s.pathUtils.RecreateDirectoryStructure(srcFileFullPath, srcBase, dstBase)
		if err != nil {
			s.output.PrintErrorf("Error recreating directory structure for %s: %v\n", srcFileFullPath, err)
			continue
		}

		relativePath, err := s.pathUtils.GetRelativePath(srcFileFullPath, srcBase)
		if err != nil {
			relativePath = s.pathUtils.GetBaseName(srcFileFullPath)
		}

		fileExists, err := s.checkFileExistsForPush(dstFileFullPath, relativePath, dstBase)
		if err != nil {
			continue
		}

		shouldCopy := s.shouldCopyFile(srcFileFullPath, dstFileFullPath, fileExists, overwriteHeaders, relativePath)
		if !shouldCopy {
			continue
		}

		if s.copySingleFileForPush(srcFileFullPath, dstFileFullPath, relativePath, fileExists, overwriteHeaders, dstBase, result) {
			result.HasChanges = true
		}
	}

	return result
}

// checkFileExistsForPush checks if destination file exists for push operation
func (s *SyncService) checkFileExistsForPush(dstFileFullPath, relativePath, dstBase string) (bool, error) {
	exists, err := s.fileOps.FileExists(dstFileFullPath)
	if err != nil {
		s.output.PrintErrorf("Error checking destination file %s in %s: %v\n", relativePath, dstBase, err)
		return false, err
	}
	return exists, nil
}

// copySingleFileForPush copies a single file for push operation
func (s *SyncService) copySingleFileForPush(srcFileFullPath, dstFileFullPath, relativePath string, fileExists, overwriteHeaders bool, dstBase string, result *models.SyncResult) bool {
	copyErr := s.fileService.Copy(srcFileFullPath, dstFileFullPath, overwriteHeaders)
	if copyErr != nil {
		s.output.PrintErrorf("Error synchronizing file %s to %s: %v\n", relativePath, dstBase, copyErr)
		return false
	}

	operationType := "update"
	if !fileExists {
		operationType = "add"
	}
	s.output.PrintOperationWithTarget(operationType, relativePath, s.pathUtils.GetBaseName(dstBase))

	result.Operations = append(result.Operations, models.FileOperation{
		Type:         models.OperationType(operationType),
		SourcePath:   srcFileFullPath,
		TargetPath:   dstFileFullPath,
		RelativePath: relativePath,
	})

	return true
}
