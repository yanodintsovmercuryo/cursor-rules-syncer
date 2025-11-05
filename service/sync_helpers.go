package service

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	cursorRulesDirEnvVar = "CURSOR_RULES_DIR"
	mdcExtension         = ".mdc"
	cursorDirName        = ".cursor"
	rulesDirName         = "rules"
	headerSeparator      = "---"
	ruleignoreFileName   = ".ruleignore"
)

// GetRulesSourceDir retrieves the path to the rules directory from flag or environment variable.
func (s *SyncService) GetRulesSourceDir(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	rulesDir := os.Getenv(cursorRulesDirEnvVar)
	if rulesDir == "" {
		return "", fmt.Errorf("rules directory not specified: use --rules-dir flag or set %s environment variable", cursorRulesDirEnvVar)
	}
	return rulesDir, nil
}

// getGitRootDir finds the root of the Git repository starting from the given directory.
func (s *SyncService) getGitRootDir(startDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = startDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error finding git repository root: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// findAllFiles finds all files in the specified directory recursively.
func (s *SyncService) findAllFiles(dir string) ([]string, error) {
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

// RecreateDirectoryStructure recreates the directory structure for the destination file
func (s *SyncService) RecreateDirectoryStructure(srcPath, srcBase, dstBase string) (string, error) {
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

// GetRelativePath returns the relative path of a file from base directory
func (s *SyncService) GetRelativePath(filePath, baseDir string) (string, error) {
	return filepath.Rel(baseDir, filePath)
}

// cleanupExtraFiles removes files that exist in destination but not in source
func (s *SyncService) cleanupExtraFiles(srcFiles []string, srcBase, dstBase string, patterns interface{}) error {
	// Build map of source files by relative path
	srcFilesMap := make(map[string]bool)
	for _, srcFile := range srcFiles {
		relativePath, err := s.GetRelativePath(srcFile, srcBase)
		if err != nil {
			continue
		}
		srcFilesMap[relativePath] = true
	}

	// Find all destination files
	var destFiles []string
	err := filepath.Walk(dstBase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			destFiles = append(destFiles, path)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking destination directory: %w", err)
	}

	// Remove files that don't exist in source
	for _, destFile := range destFiles {
		relativePath, err := s.GetRelativePath(destFile, dstBase)
		if err != nil {
			continue
		}

		if !srcFilesMap[relativePath] {
			if err := os.Remove(destFile); err != nil {
				s.outputService.PrintErrorf("Error deleting file %s: %v\n", relativePath, err)
			} else {
				s.outputService.PrintOperation("delete", relativePath)
			}
		}
	}

	return nil
}

// copyFile copies a file, optionally preserving headers
func (s *SyncService) copyFile(srcPath, dstPath string, preserveHeaders bool) error {
	// Read source file content completely
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", srcPath, err)
	}

	// Normalize line endings in source content
	srcContentStr := strings.ReplaceAll(string(srcContent), "\r\n", "\n")
	srcContentStr = strings.ReplaceAll(srcContentStr, "\r", "\n")

	var finalContent string

	if preserveHeaders {
		// Check destination file existence and extract header
		existingHeader, err := s.ExtractExistingHeader(dstPath)
		if err != nil {
			return err
		}

		// If there's an existing header in destination, extract content from source without its header
		if existingHeader != "" {
			srcContentWithoutHeader := s.RemoveHeaderFromContent(srcContentStr)
			finalContent = existingHeader + srcContentWithoutHeader
		} else {
			finalContent = srcContentStr
		}
	} else {
		// Simply overwrite without preserving headers
		finalContent = srcContentStr
	}

	// Ensure file ends with newline
	if !strings.HasSuffix(finalContent, "\n") {
		finalContent += "\n"
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dstPath, err)
	}

	// Write final content
	err = os.WriteFile(dstPath, []byte(finalContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dstPath, err)
	}

	return nil
}

// RemoveHeaderFromContent removes the YAML header from markdown content
func (s *SyncService) RemoveHeaderFromContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	// Check if it starts with header separator
	if lines[0] != headerSeparator {
		return content // No header, return as is
	}

	// Look for closing separator
	for i := 1; i < len(lines); i++ {
		if lines[i] == headerSeparator {
			// Found closing separator, return content after it
			if i+1 < len(lines) {
				remainingLines := lines[i+1:]
				// Remove leading empty lines
				for len(remainingLines) > 0 && strings.TrimSpace(remainingLines[0]) == "" {
					remainingLines = remainingLines[1:]
				}
				return strings.Join(remainingLines, "\n")
			}
			return "" // Header takes up entire file
		}
		// Limit header search to reasonable number of lines
		if i > 20 {
			break
		}
	}

	// No closing separator found, return entire content
	return content
}

// ExtractHeaderFromContent extracts the YAML header from markdown content
func (s *SyncService) ExtractHeaderFromContent(content string) string {
	// Normalize line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != headerSeparator {
		return "" // No header
	}

	var headerLines []string
	headerLines = append(headerLines, lines[0]) // Add first separator

	// Look for closing separator
	for i := 1; i < len(lines) && i <= 20; i++ { // Limit search to 20 lines
		headerLines = append(headerLines, lines[i])
		if lines[i] == headerSeparator {
			// Found closing separator, return header with newline
			return strings.Join(headerLines, "\n") + "\n"
		}
	}

	return "" // No proper header found
}

// extractExistingHeader extracts the YAML header from an existing file
func (s *SyncService) ExtractExistingHeader(dstPath string) (string, error) {
	if _, statErr := os.Stat(dstPath); statErr != nil {
		return "", nil // File doesn't exist, no header to preserve
	}

	content, err := os.ReadFile(dstPath)
	if err != nil {
		return "", nil // Can't read file, continue without header
	}

	return s.ExtractHeaderFromContent(string(content)), nil
}

// copyFileBasedOnExtension copies a file, applying header preservation only for .mdc files
func (s *SyncService) copyFileBasedOnExtension(srcPath, dstPath string, overwriteHeaders bool) error {
	// Only apply header logic for .mdc files
	if filepath.Ext(srcPath) == mdcExtension {
		return s.copyFile(srcPath, dstPath, !overwriteHeaders)
	}

	// For non-.mdc files, just copy directly without header processing
	return s.copyFileDirectly(srcPath, dstPath)
}

// copyFileDirectly copies a file from source to destination
func (s *SyncService) copyFileDirectly(srcPath, dstPath string) error {
	// Read source file
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", srcPath, err)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(dstPath), os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", dstPath, err)
	}

	// Write file
	err = os.WriteFile(dstPath, srcContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dstPath, err)
	}

	return nil
}

// filesAreEqualBasedOnExtension compares files, using header-aware comparison only for .mdc files
func (s *SyncService) filesAreEqualBasedOnExtension(file1, file2 string, overwriteHeaders bool) (bool, error) {
	// Only apply header logic for .mdc files
	if filepath.Ext(file1) == mdcExtension && filepath.Ext(file2) == mdcExtension {
		if overwriteHeaders {
			return s.filesAreEqualNormalized(file1, file2)
		} else {
			return s.filesAreEqualNormalizedWithoutHeaders(file1, file2)
		}
	}

	// For non-.mdc files, use simple comparison
	return s.filesAreEqualNormalized(file1, file2)
}

// filesAreEqualNormalized compares two files after normalizing their content
func (s *SyncService) filesAreEqualNormalized(file1, file2 string) (bool, error) {
	content1, err := s.readFileNormalized(file1)
	if err != nil {
		return false, err
	}

	content2, err := s.readFileNormalized(file2)
	if err != nil {
		return false, err
	}

	return content1 == content2, nil
}

// filesAreEqualNormalizedWithoutHeaders compares two files ignoring YAML headers
func (s *SyncService) filesAreEqualNormalizedWithoutHeaders(file1, file2 string) (bool, error) {
	content1, err := s.readFileNormalized(file1)
	if err != nil {
		return false, err
	}

	content2, err := s.readFileNormalized(file2)
	if err != nil {
		return false, err
	}

	// Remove headers from both files before comparison
	contentWithoutHeader1 := s.RemoveHeaderFromContent(content1)
	contentWithoutHeader2 := s.RemoveHeaderFromContent(content2)

	return contentWithoutHeader1 == contentWithoutHeader2, nil
}

// readFileNormalized reads a file and normalizes line endings
func (s *SyncService) readFileNormalized(filePath string) (string, error) {
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

// checkGitRemoteOrigin checks if a git repository has a remote origin
func (s *SyncService) checkGitRemoteOrigin(repoDir string) (bool, error) {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = repoDir
	if err := cmd.Run(); err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		return false, fmt.Errorf("error checking git remote origin: %w", err)
	}
	return true, nil
}

// commitChanges performs git add ., git commit -m "message", and git push in the specified directory.
// Git push is only attempted if 'origin' remote exists and gitWithoutPush is false. Output is minimal, only errors or specific statuses.
func (s *SyncService) commitChanges(repoDir string, commitMessage string, gitWithoutPush bool) error {
	// Use git add -A instead of git add . to add all changes including deletions
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = repoDir
	output, err := addCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running 'git add -A' in %s: %s\n%v\n", repoDir, string(output), err)
	}

	// Check status for debugging
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = repoDir
	statusOutput, statusErr := statusCmd.CombinedOutput()
	if statusErr != nil {
		s.outputService.PrintWarningf("Warning: could not check git status in %s: %v\n", repoDir, statusErr)
	} else {
		statusLines := strings.TrimSpace(string(statusOutput))
		if statusLines == "" {
			return nil // No changes to commit
		}
	}

	commitCmd := exec.Command("git", "commit", "-m", commitMessage)
	commitCmd.Dir = repoDir
	output, err = commitCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "nothing to commit") || strings.Contains(string(output), "no changes added to commit") {
			return nil // Not an error, just nothing to do for commit
		}
		return fmt.Errorf("error running 'git commit' in %s: %s\n%v\n", repoDir, string(output), err)
	}

	if !gitWithoutPush {
		originExists, err := s.checkGitRemoteOrigin(repoDir)
		if err != nil {
			s.outputService.PrintWarningf("Could not verify remote 'origin' in %s: %v. Skipping push.\n", repoDir, err)
			// Not returning error here, as push is optional
		}

		if originExists {
			pushCmd := exec.Command("git", "push")
			pushCmd.Dir = repoDir
			output, err = pushCmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("error running 'git push' in %s: %s\n%v\n", repoDir, string(output), err)
			}
		}
	}
	return nil
}
