package main

import (
	"bufio"
	"fmt"
	"io"
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
func loadRuleignore(rulesDir string, ignoreFilesFlag string) (map[string]bool, error) {
	ignoreMap := make(map[string]bool)

	// Load from command line flag first
	if ignoreFilesFlag != "" {
		files := strings.Split(ignoreFilesFlag, ",")
		for _, file := range files {
			file = strings.TrimSpace(file)
			if file != "" {
				ignoreMap[file] = true
			}
		}
	}

	// Load from .ruleignore file
	ruleignorePath := filepath.Join(rulesDir, ruleignoreFileName)
	file, err := os.Open(ruleignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return ignoreMap, nil // .ruleignore doesn't exist, that's fine
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
		ignoreMap[line] = true
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading %s: %w", ruleignorePath, err)
	}

	return ignoreMap, nil
}

// checkIgnoredFilesConflict checks if any ignored files exist in destination and returns error
func checkIgnoredFilesConflict(destDir string, ignoreMap map[string]bool) error {
	if len(ignoreMap) == 0 {
		return nil
	}

	var conflicts []string
	for fileName := range ignoreMap {
		destPath := filepath.Join(destDir, fileName)
		if _, err := os.Stat(destPath); err == nil {
			conflicts = append(conflicts, fileName)
		}
	}

	if len(conflicts) > 0 {
		return fmt.Errorf("ignored files exist in destination (remove them or update .ruleignore): %s", strings.Join(conflicts, ", "))
	}

	return nil
}

// filterIgnoredFiles removes ignored files from the list
func filterIgnoredFiles(files []string, ignoreMap map[string]bool) []string {
	if len(ignoreMap) == 0 {
		return files
	}

	var filtered []string
	for _, file := range files {
		fileName := filepath.Base(file)
		if !ignoreMap[fileName] {
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

// findMdcFiles finds all .mdc files in the specified directory.
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

// copyFileWithHeaderPreservation copies a file, preserving the header of the destination file if it exists.
// The header is defined by lines between --- at the beginning of the file.
func copyFileWithHeaderPreservation(srcPath, dstPath string) error {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", srcPath, err)
	}
	defer srcFile.Close()

	existingHeader, err := extractExistingHeader(dstPath)
	if err != nil {
		return err
	}

	newDstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create/overwrite destination file %s: %w", dstPath, err)
	}
	defer newDstFile.Close()

	if existingHeader != "" {
		_, err = newDstFile.WriteString(existingHeader)
		if err != nil {
			return fmt.Errorf("failed to write header to %s: %w", dstPath, err)
		}
	}

	return copySourceContent(srcFile, newDstFile, srcPath, dstPath, existingHeader != "")
}

// extractExistingHeader reads and extracts the header from the destination file if it exists
func extractExistingHeader(dstPath string) (string, error) {
	if _, statErr := os.Stat(dstPath); statErr != nil {
		return "", nil // File doesn't exist, no header to preserve
	}

	dstFile, openErr := os.Open(dstPath)
	if openErr != nil {
		return "", nil // Can't open file, continue without header
	}
	defer dstFile.Close()

	return readHeaderFromFile(dstFile, dstPath)
}

// readHeaderFromFile reads the header section from a file
func readHeaderFromFile(file *os.File, filePath string) (string, error) {
	scanner := bufio.NewScanner(file)
	inHeader := false
	lineCount := 0
	var headerLines []string
	firstSeparatorFound := false

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		switch {
		case line == headerSeparator:
			headerLines = append(headerLines, line)
			if !firstSeparatorFound {
				inHeader = true
				firstSeparatorFound = true
			} else {
				// Second separator found, header is complete
				goto headerComplete
			}
		case inHeader:
			headerLines = append(headerLines, line)
		case lineCount == 1 && line != headerSeparator:
			// First line is not a separator, so no header
			goto headerComplete
		}
		// Limit the number of lines for the header to avoid reading the entire file
		if lineCount > 20 && inHeader {
			fmt.Printf("Warning: Header in %s is too long or not properly terminated, it might not be fully preserved.\n", filePath)
			headerLines = nil // Discard potentially incomplete header
			break
		}
	}

headerComplete:
	if scanErr := scanner.Err(); scanErr != nil {
		return "", fmt.Errorf("error reading destination file %s: %w", filePath, scanErr)
	}

	if len(headerLines) > 1 && headerLines[0] == headerSeparator && headerLines[len(headerLines)-1] == headerSeparator {
		return strings.Join(headerLines, "\n") + "\n", nil
	}

	return "", nil
}

// copySourceContent copies content from source file to destination, optionally skipping source header
func copySourceContent(srcFile *os.File, dstFile *os.File, srcPath, dstPath string, preserveDestHeader bool) error {
	srcScanner := bufio.NewScanner(srcFile)

	if preserveDestHeader {
		if err := skipSourceHeader(srcFile, srcScanner, srcPath); err != nil {
			return err
		}
	}

	return writeContentLines(srcScanner, dstFile, srcPath, dstPath, preserveDestHeader)
}

// skipSourceHeader skips the header section in the source file if destination header is being preserved
func skipSourceHeader(srcFile *os.File, srcScanner *bufio.Scanner, srcPath string) error {
	initialSrcFilePosition, seekErr := srcFile.Seek(0, io.SeekCurrent)
	if seekErr != nil {
		return fmt.Errorf("error getting source file position %s: %w", srcPath, seekErr)
	}

	srcContentLines := 0
	skippingSrcHeader := false
	srcFirstSeparatorFound := false

	for srcScanner.Scan() {
		line := srcScanner.Text()
		srcContentLines++

		if line == headerSeparator {
			if !srcFirstSeparatorFound {
				srcFirstSeparatorFound = true
			} else {
				skippingSrcHeader = true // Found the second separator, start copying from the next line
				break
			}
		}
		if srcContentLines > 20 && srcFirstSeparatorFound {
			// Could not find the end of the source header, copy everything from source
			skippingSrcHeader = false
			break
		}
		if srcContentLines > 1 && !srcFirstSeparatorFound && line != headerSeparator {
			// First line was not a separator, so no header in source or malformed
			skippingSrcHeader = false
			break
		}
	}

	if srcScanner.Err() != nil {
		return fmt.Errorf("error skipping source file header %s: %w", srcPath, srcScanner.Err())
	}

	if !skippingSrcHeader {
		// If source header was not skipped (e.g. not found, too long), reset scanner and write all read lines
		_, err := srcFile.Seek(initialSrcFilePosition, io.SeekStart) // Reset to position before header scan
		if err != nil {
			return fmt.Errorf("error seeking source file %s: %w", srcPath, err)
		}
		// Re-initialize scanner - this will be done by the caller
	}

	return nil
}

// writeContentLines writes the remaining content lines to the destination file
func writeContentLines(srcScanner *bufio.Scanner, dstFile *os.File, srcPath, dstPath string, hasDestHeader bool) error {
	firstRealLineWritten := false
	for srcScanner.Scan() {
		line := srcScanner.Text()
		if !hasDestHeader && !firstRealLineWritten && strings.TrimSpace(line) == "" {
			// Skip leading empty lines if there's no destination header being preserved
			continue
		}
		_, err := dstFile.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("error writing data to %s: %w", dstPath, err)
		}
		firstRealLineWritten = true
	}
	if err := srcScanner.Err(); err != nil {
		return fmt.Errorf("error reading source file %s: %w", srcPath, err)
	}

	return nil
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
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = repoDir
	output, err := addCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running 'git add .': %s\n%v\n", string(output), err)
		return err // Return the original error to allow main to decide if it's fatal for the whole op
	}
	// No output for successful add

	commitCmd := exec.Command("git", "commit", "-m", commitMessage)
	commitCmd.Dir = repoDir
	output, err = commitCmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "nothing to commit") || strings.Contains(string(output), "no changes added to commit") {
			// fmt.Println("No changes to commit.") // Suppressed for minimal output unless specifically requested
			return nil // Not an error, just nothing to do for commit
		}
		fmt.Fprintf(os.Stderr, "Error running 'git commit': %s\n%v\n", string(output), err)
		return err
	}
	// No output for successful commit, or a very brief one if needed e.g. fmt.Printf("Committed: %s\n", commitMessage)
	// For now, keeping it minimal. The calling function (push) will know if it was called.

	if !gitWithoutPush {
		originExists, err := checkGitRemoteOrigin(repoDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not verify remote 'origin': %v. Skipping push.\n", err)
			// Not returning error here, as push is optional
		}

		if originExists {
			pushCmd := exec.Command("git", "push")
			pushCmd.Dir = repoDir
			output, err = pushCmd.CombinedOutput()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running 'git push': %s\n%v\n", string(output), err)
				return err // Push error should be reported
			}
			// No output for successful push
		}
	}
	return nil
}

// filesAreEqual compares the content of two files and returns true if they are identical.
func filesAreEqual(file1, file2 string) (bool, error) {
	return compareFiles(file1, file2)
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
