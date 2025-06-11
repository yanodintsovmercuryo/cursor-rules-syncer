package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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

// ignorePattern represents a compiled ignore pattern
type ignorePattern struct {
	pattern    string
	regex      *regexp.Regexp
	isNegation bool
}

// getRulesSourceDir retrieves the path to the rules directory from flag or environment variable.
func getRulesSourceDir(flagValue string) (string, error) {
	if flagValue != "" {
		return flagValue, nil
	}

	rulesDir := os.Getenv(cursorRulesDirEnvVar)
	if rulesDir == "" {
		return "", fmt.Errorf("rules directory not specified: use --rules-dir flag or set %s environment variable", cursorRulesDirEnvVar)
	}
	return rulesDir, nil
}

// loadRuleignore loads ignore patterns from .ruleignore file and command line flag
func loadRuleignore(rulesDir string, ignoreFilesFlag string) ([]ignorePattern, error) {
	var patterns []ignorePattern

	// Load from command line flag first
	if ignoreFilesFlag != "" {
		files := strings.Split(ignoreFilesFlag, ",")
		for _, file := range files {
			file = strings.TrimSpace(file)
			if file != "" {
				pattern, err := compileIgnorePattern(file)
				if err != nil {
					return nil, fmt.Errorf("error compiling ignore pattern '%s': %w", file, err)
				}
				patterns = append(patterns, pattern)
			}
		}
	}

	// Load from .ruleignore file
	ruleignorePath := filepath.Join(rulesDir, ruleignoreFileName)
	file, err := os.Open(ruleignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return patterns, nil // .ruleignore doesn't exist, that's fine
		}
		return nil, fmt.Errorf("error opening %s: %w", ruleignorePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pattern, err := compileIgnorePattern(line)
		if err != nil {
			return nil, fmt.Errorf("error compiling ignore pattern '%s' in %s: %w", line, ruleignorePath, err)
		}
		patterns = append(patterns, pattern)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", ruleignorePath, err)
	}

	return patterns, nil
}

// compileIgnorePattern compiles a gitignore-style pattern into a regex
func compileIgnorePattern(pattern string) (ignorePattern, error) {
	isNegation := strings.HasPrefix(pattern, "!")
	if isNegation {
		pattern = pattern[1:]
	}

	// Convert gitignore pattern to regex
	regexPattern := gitignoreToRegex(pattern)

	regex, err := regexp.Compile(regexPattern)
	if err != nil {
		return ignorePattern{}, err
	}

	return ignorePattern{
		pattern:    pattern,
		regex:      regex,
		isNegation: isNegation,
	}, nil
}

// gitignoreToRegex converts a gitignore pattern to a regex pattern
func gitignoreToRegex(pattern string) string {
	// Escape special regex characters, but handle gitignore special chars differently
	result := strings.Builder{}

	i := 0
	for i < len(pattern) {
		switch pattern[i] {
		case '*':
			if i+1 < len(pattern) && pattern[i+1] == '*' {
				// ** matches any number of directories
				if i+2 < len(pattern) && pattern[i+2] == '/' {
					result.WriteString("(?:.*/)?")
					i += 3
				} else if i == 0 || pattern[i-1] == '/' {
					result.WriteString(".*")
					i += 2
				} else {
					result.WriteString("[^/]*")
					i++
				}
			} else {
				// * matches anything except /
				result.WriteString("[^/]*")
				i++
			}
		case '?':
			result.WriteString("[^/]")
			i++
		case '[':
			// Character class - pass through
			result.WriteByte('[')
			i++
		case ']':
			result.WriteByte(']')
			i++
		case '.', '^', '$', '+', '{', '}', '|', '(', ')':
			// Escape regex special characters
			result.WriteByte('\\')
			result.WriteByte(pattern[i])
			i++
		default:
			result.WriteByte(pattern[i])
			i++
		}
	}

	// Add anchors
	return "^" + result.String() + "$"
}

// shouldIgnoreFile checks if a file should be ignored based on patterns
func shouldIgnoreFile(relativePath string, patterns []ignorePattern) bool {
	ignored := false

	for _, pattern := range patterns {
		if pattern.regex.MatchString(relativePath) || pattern.regex.MatchString(filepath.Base(relativePath)) {
			if pattern.isNegation {
				ignored = false
			} else {
				ignored = true
			}
		}
	}

	return ignored
}

// checkIgnoredFilesConflict checks if any ignored files exist in destination and returns error
func checkIgnoredFilesConflict(destDir string, patterns []ignorePattern) error {
	if len(patterns) == 0 {
		return nil
	}

	var conflicts []string

	err := filepath.Walk(destDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relativePath, err := filepath.Rel(destDir, path)
			if err != nil {
				return err
			}

			if shouldIgnoreFile(relativePath, patterns) {
				conflicts = append(conflicts, relativePath)
			}
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error checking for conflicts: %w", err)
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("ignored files exist in destination (remove them or update .ruleignore): %s", strings.Join(conflicts, ", "))
	}

	return nil
}

// filterIgnoredFiles removes ignored files from the list
func filterIgnoredFiles(files []string, patterns []ignorePattern, baseDir string) []string {
	if len(patterns) == 0 {
		return files
	}

	var filtered []string
	for _, file := range files {
		relativePath, err := filepath.Rel(baseDir, file)
		if err != nil {
			// If we can't get relative path, include the file
			filtered = append(filtered, file)
			continue
		}

		if !shouldIgnoreFile(relativePath, patterns) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// getGitRootDir finds the root of the Git repository starting from the given directory.
func getGitRootDir(startDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = startDir
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("error finding git repository root: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// findMdcFiles finds all .mdc files in the specified directory recursively.
func findMdcFiles(dir string) ([]string, error) {
	var mdcFiles []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == mdcExtension {
			mdcFiles = append(mdcFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error finding .mdc files in %s: %w", dir, err)
	}
	return mdcFiles, nil
}

// findAllFiles finds all files in the specified directory recursively.
func findAllFiles(dir string) ([]string, error) {
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

// copyFile copies a file, optionally preserving headers
func copyFile(srcPath, dstPath string, preserveHeaders bool) error {
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
		existingHeader, err := extractExistingHeader(dstPath)
		if err != nil {
			return err
		}

		// If there's an existing header in destination, extract content from source without its header
		if existingHeader != "" {
			srcContentWithoutHeader := removeHeaderFromContent(srcContentStr)
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

// removeHeaderFromContent removes header from content string
func removeHeaderFromContent(content string) string {
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

// extractExistingHeader reads and extracts the header from the destination file if it exists
func extractExistingHeader(dstPath string) (string, error) {
	if _, statErr := os.Stat(dstPath); statErr != nil {
		return "", nil // File doesn't exist, no header to preserve
	}

	content, err := os.ReadFile(dstPath)
	if err != nil {
		return "", nil // Can't read file, continue without header
	}

	return extractHeaderFromContent(string(content)), nil
}

// extractHeaderFromContent extracts header from content string
func extractHeaderFromContent(content string) string {
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

// checkGitRemoteOrigin checks if 'origin' remote exists.
func checkGitRemoteOrigin(repoDir string) (bool, error) {
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
func commitChanges(repoDir string, commitMessage string, gitWithoutPush bool) error {
	// Use git add -A instead of git add . to add all changes including deletions
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = repoDir
	output, err := addCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running 'git add -A' in %s: %s\n%v\n", repoDir, string(output), err)
		return err
	}

	// Check status for debugging
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = repoDir
	statusOutput, statusErr := statusCmd.CombinedOutput()
	if statusErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not check git status in %s: %v\n", repoDir, statusErr)
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
		fmt.Fprintf(os.Stderr, "Error running 'git commit' in %s: %s\n%v\n", repoDir, string(output), err)
		return err
	}

	if !gitWithoutPush {
		originExists, err := checkGitRemoteOrigin(repoDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not verify remote 'origin' in %s: %v. Skipping push.\n", repoDir, err)
			// Not returning error here, as push is optional
		}

		if originExists {
			pushCmd := exec.Command("git", "push")
			pushCmd.Dir = repoDir
			output, err = pushCmd.CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running 'git push' in %s: %s\n%v\n", repoDir, string(output), err)
				return err // Push error should be reported
			}
		}
	}
	return nil
}

// compareFiles compares two files for equality
func compareFiles(file1, file2 string) (bool, error) {
	info1, err := os.Stat(file1)
	if err != nil {
		return false, err
	}
	info2, err := os.Stat(file2)
	if err != nil {
		return false, err
	}

	// If sizes are different, files are not equal
	if info1.Size() != info2.Size() {
		return false, nil
	}

	return compareFileContents(file1, file2)
}

// compareFileContents compares the actual content of two files
func compareFileContents(file1, file2 string) (bool, error) {
	f1, err := os.Open(file1)
	if err != nil {
		return false, err
	}
	defer f1.Close()

	f2, err := os.Open(file2)
	if err != nil {
		return false, err
	}
	defer f2.Close()

	return compareReaders(f1, f2)
}

// compareReaders compares content from two readers
func compareReaders(r1, r2 io.Reader) (bool, error) {
	const chunkSize = 4096
	buf1 := make([]byte, chunkSize)
	buf2 := make([]byte, chunkSize)

	for {
		n1, err1 := r1.Read(buf1)
		n2, err2 := r2.Read(buf2)

		if n1 != n2 {
			return false, nil
		}

		if n1 > 0 && string(buf1[:n1]) != string(buf2[:n2]) {
			return false, nil
		}

		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}

		if err1 != nil || err2 != nil {
			if err1 == io.EOF || err2 == io.EOF {
				return false, nil
			}
			if err1 != nil {
				return false, err1
			}
			return false, err2
		}
	}
}

// filesAreEqual compares the content of two files and returns true if they are identical.
func filesAreEqual(file1, file2 string) (bool, error) {
	return compareFiles(file1, file2)
}

// filesAreEqualNormalized compares files with line ending normalization
func filesAreEqualNormalized(file1, file2 string) (bool, error) {
	content1, err := readFileNormalized(file1)
	if err != nil {
		return false, err
	}

	content2, err := readFileNormalized(file2)
	if err != nil {
		return false, err
	}

	return content1 == content2, nil
}

// filesAreEqualNormalizedWithoutHeaders compares files with line ending normalization, ignoring headers
func filesAreEqualNormalizedWithoutHeaders(file1, file2 string) (bool, error) {
	content1, err := readFileNormalized(file1)
	if err != nil {
		return false, err
	}

	content2, err := readFileNormalized(file2)
	if err != nil {
		return false, err
	}

	// Remove headers from both files before comparison
	contentWithoutHeader1 := removeHeaderFromContent(content1)
	contentWithoutHeader2 := removeHeaderFromContent(content2)

	return contentWithoutHeader1 == contentWithoutHeader2, nil
}

// readFileNormalized reads file and normalizes line endings
func readFileNormalized(filePath string) (string, error) {
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

// copyFileWithHeaderPreservation copies a file, preserving the header of the destination file if it exists.
// This is for backward compatibility
func copyFileWithHeaderPreservation(srcPath, dstPath string) error {
	return copyFile(srcPath, dstPath, true)
}

// copyFileWithoutHeaderPreservation copies a file, overwriting headers
func copyFileWithoutHeaderPreservation(srcPath, dstPath string) error {
	return copyFile(srcPath, dstPath, false)
}

// getRelativePathInRules gets the relative path of a file within the rules structure
func getRelativePathInRules(filePath, baseRulesDir string) (string, error) {
	relativePath, err := filepath.Rel(baseRulesDir, filePath)
	if err != nil {
		return "", fmt.Errorf("cannot determine relative path for %s: %w", filePath, err)
	}
	return relativePath, nil
}

// recreateDirectoryStructure ensures the directory structure exists for a file
func recreateDirectoryStructure(srcPath, srcBase, dstBase string) (string, error) {
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

// cleanupExtraFiles removes files that exist in destination but not in source
func cleanupExtraFiles(srcFiles []string, srcBase, dstBase string, patterns []ignorePattern) error {
	// Build map of source files by relative path
	srcFilesMap := make(map[string]bool)
	for _, srcFile := range srcFiles {
		relativePath, err := filepath.Rel(srcBase, srcFile)
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

	// Filter out ignored files from destination
	destFiles = filterIgnoredFiles(destFiles, patterns, dstBase)

	// Remove files that don't exist in source
	for _, destFile := range destFiles {
		relativePath, err := filepath.Rel(dstBase, destFile)
		if err != nil {
			continue
		}

		if !srcFilesMap[relativePath] {
			if err := os.Remove(destFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error deleting file %s: %v\n", relativePath, err)
			} else {
				fmt.Printf("%s%s %s%s\n", colorRed, symbolDelete, relativePath, colorReset)
			}
		}
	}

	return nil
}

// copyFileBasedOnExtension copies a file, applying header preservation only for .mdc files
func copyFileBasedOnExtension(srcPath, dstPath string, overwriteHeaders bool) error {
	// Only apply header logic for .mdc files
	if filepath.Ext(srcPath) == mdcExtension {
		return copyFile(srcPath, dstPath, !overwriteHeaders)
	}

	// For non-.mdc files, just copy directly without header processing
	return copyFileDirectly(srcPath, dstPath)
}

// copyFileDirectly copies a file without any header processing
func copyFileDirectly(srcPath, dstPath string) error {
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
func filesAreEqualBasedOnExtension(file1, file2 string, overwriteHeaders bool) (bool, error) {
	// Only apply header logic for .mdc files
	if filepath.Ext(file1) == mdcExtension && filepath.Ext(file2) == mdcExtension {
		if overwriteHeaders {
			return filesAreEqualNormalized(file1, file2)
		} else {
			return filesAreEqualNormalizedWithoutHeaders(file1, file2)
		}
	}

	// For non-.mdc files, use simple comparison
	return filesAreEqualNormalized(file1, file2)
}
