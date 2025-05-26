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

	// Try to read the header from the destination file if it exists
	existingHeader := ""
	if _, err := os.Stat(dstPath); err == nil {
		dstFile, openErr := os.Open(dstPath)
		if openErr != nil {
			// If opening failed, it might not be an error (file doesn't exist or is corrupted)
			// Continue without a header
		} else {
			defer dstFile.Close()
			scanner := bufio.NewScanner(dstFile)
			inHeader := false
			lineCount := 0
			var headerLines []string
			firstSeparatorFound := false

			for scanner.Scan() {
				line := scanner.Text()
				lineCount++

				if line == headerSeparator {
					headerLines = append(headerLines, line)
					if !firstSeparatorFound {
						inHeader = true
						firstSeparatorFound = true
					} else {
						// Second separator found, header is complete
						break
					}
				} else if inHeader {
					headerLines = append(headerLines, line)
				} else if lineCount == 1 && line != headerSeparator {
					// First line is not a separator, so no header
					break
				}
				// Limit the number of lines for the header to avoid reading the entire file
				if lineCount > 20 && inHeader { // Increased limit slightly
					fmt.Printf("Warning: Header in %s is too long or not properly terminated, it might not be fully preserved.\n", dstPath)
					headerLines = nil // Discard potentially incomplete header
					break
				}
			}

			if err := scanner.Err(); err != nil {
				return fmt.Errorf("error reading destination file %s: %w", dstPath, err)
			}

			if len(headerLines) > 1 && headerLines[0] == headerSeparator && headerLines[len(headerLines)-1] == headerSeparator {
				existingHeader = strings.Join(headerLines, "\n") + "\n"
			}
		}
	}

	// Create or overwrite the destination file
	newDstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create/overwrite destination file %s: %w", dstPath, err)
	}
	defer newDstFile.Close()

	// Write the existing header if it's not empty
	if existingHeader != "" {
		_, err = newDstFile.WriteString(existingHeader)
		if err != nil {
			return fmt.Errorf("failed to write header to %s: %w", dstPath, err)
		}
	}

	// Copy the rest of the content from the source file
	srcScanner := bufio.NewScanner(srcFile)
	srcContentLines := 0
	skippingSrcHeader := false
	srcFirstSeparatorFound := false

	if existingHeader != "" { // If we are preserving the destination header, skip the source header
		// Create a temporary slice to hold lines we might need if source header is not found
		var potentialBodyLines []string
		initialSrcFilePosition, _ := srcFile.Seek(0, io.SeekCurrent) // Get current position

		for srcScanner.Scan() {
			line := srcScanner.Text()
			srcContentLines++
			potentialBodyLines = append(potentialBodyLines, line)

			if line == headerSeparator {
				if !srcFirstSeparatorFound {
					srcFirstSeparatorFound = true
				} else {
					skippingSrcHeader = true // Found the second separator, start copying from the next line
					break
				}
			}
			if srcContentLines > 20 && srcFirstSeparatorFound { // Increased limit
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
			_, err = srcFile.Seek(initialSrcFilePosition, io.SeekStart) // Reset to position before header scan
			if err != nil {
				return fmt.Errorf("error seeking source file %s: %w", srcPath, err)
			}
			srcScanner = bufio.NewScanner(srcFile) // Re-initialize scanner
			// The content will be written by the next loop.
		}
	}

	// Copy content (after the header, if it was skipped)
	firstRealLineWritten := false
	for srcScanner.Scan() {
		line := srcScanner.Text()
		if existingHeader == "" && !firstRealLineWritten && strings.TrimSpace(line) == "" {
			// Skip leading empty lines if there's no destination header being preserved
			continue
		}
		_, err = newDstFile.WriteString(line + "\n")
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
		fmt.Fprintf(os.Stderr, "Error running 'git add .': %s\n%w\n", string(output), err)
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
		fmt.Fprintf(os.Stderr, "Error running 'git commit': %s\n%w\n", string(output), err)
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
				fmt.Fprintf(os.Stderr, "Error running 'git push': %s\n%w\n", string(output), err)
				return err // Push error should be reported
			}
			// No output for successful push
		} else {
			// fmt.Println("Remote 'origin' not found or not accessible. Skipping 'git push'.") // Suppressed
			// fmt.Println("Please push the changes manually if required.") // Suppressed
		}
	}
	return nil
}

// filesAreEqual compares the content of two files and returns true if they are identical.
func filesAreEqual(file1, file2 string) (bool, error) {
	// Check if both files exist
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

	// Open both files
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

	// Compare content in chunks
	const chunkSize = 4096
	buf1 := make([]byte, chunkSize)
	buf2 := make([]byte, chunkSize)

	for {
		n1, err1 := f1.Read(buf1)
		n2, err2 := f2.Read(buf2)

		// If read different amounts, files are different
		if n1 != n2 {
			return false, nil
		}

		// Compare the read chunks
		if n1 > 0 && string(buf1[:n1]) != string(buf2[:n2]) {
			return false, nil
		}

		// If both reached EOF at the same time, files are equal
		if err1 == io.EOF && err2 == io.EOF {
			return true, nil
		}

		// If only one reached EOF or any other error occurred
		if err1 != nil || err2 != nil {
			if err1 == io.EOF || err2 == io.EOF {
				return false, nil // One file ended earlier than the other
			}
			// Return the first error encountered
			if err1 != nil {
				return false, err1
			}
			return false, err2
		}
	}
}
