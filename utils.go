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
)

// getRulesSourceDir retrieves the path to the rules directory from the environment variable.
func getRulesSourceDir() (string, error) {
	rulesDir := os.Getenv(cursorRulesDirEnvVar)
	if rulesDir == "" {
		return "", fmt.Errorf("environment variable %s is not set", cursorRulesDirEnvVar)
	}
	return rulesDir, nil
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
			// Exit code 128 typically means "no such remote" or other git config errors.
			// Other non-zero exit codes might also indicate 'origin' is not properly configured or accessible.
			return false, nil // Assuming any error from "git remote get-url origin" means origin is not usable/present
		}
		return false, fmt.Errorf("error checking git remote origin: %w", err) // Unexpected error
	}
	return true, nil
}

// commitChanges performs git add ., git commit -m "message", and git push in the specified directory.
// Git push is only attempted if 'origin' remote exists.
func commitChanges(repoDir string, commitMessage string) error {
	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = repoDir
	output, err := addCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running 'git add .': %s, %w", string(output), err)
	}
	fmt.Printf("'git add .' executed successfully: %s\n", string(output))

	commitCmd := exec.Command("git", "commit", "-m", commitMessage)
	commitCmd.Dir = repoDir
	output, err = commitCmd.CombinedOutput()
	if err != nil {
		// Check if the error is "nothing to commit"
		if strings.Contains(string(output), "nothing to commit") || strings.Contains(string(output), "no changes added to commit") {
			fmt.Println("No changes to commit.")
			return nil // Not an error, just nothing to do for commit
		}
		return fmt.Errorf("error running 'git commit': %s, %w", string(output), err)
	}
	fmt.Printf("'git commit' executed successfully: %s\n", string(output))

	originExists, err := checkGitRemoteOrigin(repoDir)
	if err != nil {
		// Log error but don't stop, push is optional
		fmt.Fprintf(os.Stderr, "Could not verify remote 'origin': %v. Skipping push.\n", err)
	}

	if originExists {
		pushCmd := exec.Command("git", "push")
		pushCmd.Dir = repoDir
		output, err = pushCmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("error running 'git push': %s, %w", string(output), err)
		}
		fmt.Printf("'git push' executed successfully: %s\n", string(output))
	} else {
		fmt.Println("Remote 'origin' not found or not accessible. Skipping 'git push'.")
		fmt.Println("Please push the changes manually if required.")
	}
	return nil
}
