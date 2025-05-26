package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// ANSI Color Codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
)

// Operation Symbols
const (
	symbolAdd    = "+"
	symbolDelete = "-"
	symbolUpdate = "*"
)

var (
	rulesDir       string
	gitWithoutPush bool
	ignoreFiles    string
)

var rootCmd = &cobra.Command{
	Use:   "cursor-rules-syncer",
	Short: "A CLI tool to sync cursor rules",
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pulls rules from the source directory to the current git project's .cursor/rules directory, deleting extra files in the project.",
	Run: func(cmd *cobra.Command, args []string) {
		rulesSourceDir, err := getRulesSourceDir(rulesDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		gitRoot, err := getGitRootDir(currentDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding git root: %v\n", err)
			os.Exit(1)
		}

		destRulesDir := filepath.Join(gitRoot, cursorDirName, rulesDirName)

		// Load ignore patterns
		ignoreMap, err := loadRuleignore(rulesSourceDir, ignoreFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading ignore patterns: %v\n", err)
			os.Exit(1)
		}

		// Check for conflicts with ignored files
		if err := checkIgnoredFilesConflict(destRulesDir, ignoreMap); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if err := os.MkdirAll(destRulesDir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating destination directory %s: %v\n", destRulesDir, err)
			os.Exit(1)
		}

		sourceMdcFiles, err := findMdcFiles(rulesSourceDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding source files in %s: %v\n", rulesSourceDir, err)
			os.Exit(1)
		}

		// Filter out ignored files
		sourceMdcFiles = filterIgnoredFiles(sourceMdcFiles, ignoreMap)

		sourceFilesMap := make(map[string]string) // basename -> full path
		for _, f := range sourceMdcFiles {
			sourceFilesMap[filepath.Base(f)] = f
		}

		destMdcFiles, err := findMdcFiles(destRulesDir)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error finding files in destination directory %s: %v\n", destRulesDir, err)
			}
		}

		for _, destFileFullPath := range destMdcFiles {
			fileName := filepath.Base(destFileFullPath)
			if _, existsInSource := sourceFilesMap[fileName]; !existsInSource {
				if err := os.Remove(destFileFullPath); err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting file %s: %v\n", fileName, err)
				} else {
					fmt.Printf("%s%s %s%s\n", colorRed, symbolDelete, fileName, colorReset)
				}
			}
		}

		for _, srcFileFullPath := range sourceMdcFiles {
			fileName := filepath.Base(srcFileFullPath)
			dstFileFullPath := filepath.Join(destRulesDir, fileName)

			fileExistedBeforeCopy := true
			if _, err := os.Stat(dstFileFullPath); os.IsNotExist(err) {
				fileExistedBeforeCopy = false
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking destination file %s: %v\n", fileName, err)
				continue // Skip this file if stat fails for other reasons
			}

			// Check if files are different before copying
			shouldCopy := true
			if fileExistedBeforeCopy {
				equal, err := filesAreEqual(srcFileFullPath, dstFileFullPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error comparing files %s: %v\n", fileName, err)
					// Continue with copying in case of comparison error
				} else if equal {
					shouldCopy = false // Files are identical, no need to copy
				}
			}

			if shouldCopy {
				if err := copyFileWithHeaderPreservation(srcFileFullPath, dstFileFullPath); err != nil {
					fmt.Fprintf(os.Stderr, "Error synchronizing file %s: %v\n", fileName, err)
				} else {
					if fileExistedBeforeCopy {
						fmt.Printf("%s%s %s%s\n", colorYellow, symbolUpdate, fileName, colorReset)
					} else {
						fmt.Printf("%s%s %s%s\n", colorGreen, symbolAdd, fileName, colorReset)
					}
				}
			}
		}
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes rules from the current git project's .cursor/rules directory to the source directory, deleting extra files in the source, and commits changes",
	Run: func(cmd *cobra.Command, args []string) {
		rulesEnvDir, err := getRulesSourceDir(rulesDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		projectGitRoot, err := getGitRootDir(currentDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding git root for project: %v\n", err)
			os.Exit(1)
		}

		rulesSourceDirInProject := filepath.Join(projectGitRoot, cursorDirName, rulesDirName)

		if _, err := os.Stat(rulesSourceDirInProject); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Project rules directory %s not found. Nothing to push.\n", rulesSourceDirInProject)
			os.Exit(1)
		}

		// Load ignore patterns
		ignoreMap, err := loadRuleignore(rulesEnvDir, ignoreFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading ignore patterns: %v\n", err)
			os.Exit(1)
		}

		projectMdcFiles, err := findMdcFiles(rulesSourceDirInProject)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding files in project rules directory %s: %v\n", rulesSourceDirInProject, err)
			os.Exit(1)
		}

		// Filter out ignored files
		projectMdcFiles = filterIgnoredFiles(projectMdcFiles, ignoreMap)

		projectFilesMap := make(map[string]string)
		for _, f := range projectMdcFiles {
			projectFilesMap[filepath.Base(f)] = f
		}

		if err := os.MkdirAll(rulesEnvDir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating destination directory %s: %v\n", rulesEnvDir, err)
			os.Exit(1)
		}

		destEnvMdcFiles, err := findMdcFiles(rulesEnvDir)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error finding files in environment destination directory %s: %v\n", rulesEnvDir, err)
			}
		}

		// Filter out ignored files from destination
		destEnvMdcFiles = filterIgnoredFiles(destEnvMdcFiles, ignoreMap)

		filesDeletedInEnv := false
		for _, destFileFullPath := range destEnvMdcFiles {
			fileName := filepath.Base(destFileFullPath)
			if _, existsInProject := projectFilesMap[fileName]; !existsInProject {
				if err := os.Remove(destFileFullPath); err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting file %s from %s: %v\n", fileName, rulesEnvDir, err)
				} else {
					fmt.Printf("%s%s %s (from %s)%s\n", colorRed, symbolDelete, fileName, filepath.Base(rulesEnvDir), colorReset)
					filesDeletedInEnv = true
				}
			}
		}

		filesCopiedOrUpdated := false
		if len(projectMdcFiles) == 0 && !filesDeletedInEnv {
			return
		}

		for _, srcFileFullPath := range projectMdcFiles {
			fileName := filepath.Base(srcFileFullPath)
			dstFileFullPath := filepath.Join(rulesEnvDir, fileName)

			fileExists := true
			if _, err := os.Stat(dstFileFullPath); os.IsNotExist(err) {
				fileExists = false
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking destination file %s in %s: %v\n", fileName, rulesEnvDir, err)
				continue
			}

			// Check if files are different before copying
			shouldCopy := true
			if fileExists {
				equal, err := filesAreEqual(srcFileFullPath, dstFileFullPath)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error comparing files %s: %v\n", fileName, err)
					// Continue with copying in case of comparison error
				} else if equal {
					shouldCopy = false // Files are identical, no need to copy
				}
			}

			if shouldCopy {
				if err := copyFileWithHeaderPreservation(srcFileFullPath, dstFileFullPath); err != nil {
					fmt.Fprintf(os.Stderr, "Error synchronizing file %s to %s: %v\n", fileName, rulesEnvDir, err)
				} else {
					if fileExists {
						fmt.Printf("%s%s %s (to %s)%s\n", colorYellow, symbolUpdate, fileName, filepath.Base(rulesEnvDir), colorReset)
					} else {
						fmt.Printf("%s%s %s (to %s)%s\n", colorGreen, symbolAdd, fileName, filepath.Base(rulesEnvDir), colorReset)
					}
					filesCopiedOrUpdated = true
				}
			}
		}

		if !filesCopiedOrUpdated && !filesDeletedInEnv {
			return
		}

		rulesEnvRepoRoot, err := getGitRootDir(rulesEnvDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding Git repository root for %s: %v\n", rulesEnvDir, err)
			fmt.Fprintf(os.Stderr, "Commit will not be performed. Please commit changes in %s manually if needed.\n", rulesEnvDir)
		} else {
			commitMessage := "Sync cursor rules: updated from project " + filepath.Base(projectGitRoot)
			if err := commitChanges(rulesEnvRepoRoot, commitMessage, gitWithoutPush); err != nil {
				// commitChanges handles its own error printing
			} else {
				// Success message for commit is handled by commitChanges if needed, or kept silent
			}
		}
	},
}

func init() {
	// Add flags to both commands
	pullCmd.Flags().StringVar(&rulesDir, "rules-dir", "", "Path to rules directory (overrides CURSOR_RULES_DIR env var)")
	pullCmd.Flags().StringVar(&ignoreFiles, "ignore-files", "", "Comma-separated list of files to ignore")

	pushCmd.Flags().StringVar(&rulesDir, "rules-dir", "", "Path to rules directory (overrides CURSOR_RULES_DIR env var)")
	pushCmd.Flags().BoolVar(&gitWithoutPush, "git-without-push", false, "Commit changes but don't push to remote")
	pushCmd.Flags().StringVar(&ignoreFiles, "ignore-files", "", "Comma-separated list of files to ignore")

	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pushCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
