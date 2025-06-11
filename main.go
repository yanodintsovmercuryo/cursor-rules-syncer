package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// version
const version = "v0.0.6"

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
	rulesDir         string
	gitWithoutPush   bool
	ignoreFiles      string
	overwriteHeaders bool
)

var rootCmd = &cobra.Command{
	Use:   "cursor-rules-syncer",
	Short: "A CLI tool to sync cursor rules",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version)
	},
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
		ignorePatterns, err := loadRuleignore(rulesSourceDir, ignoreFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading ignore patterns: %v\n", err)
			os.Exit(1)
		}

		// Check for conflicts with ignored files
		if conflictErr := checkIgnoredFilesConflict(destRulesDir, ignorePatterns); conflictErr != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", conflictErr)
			os.Exit(1)
		}

		if mkdirErr := os.MkdirAll(destRulesDir, os.ModePerm); mkdirErr != nil {
			fmt.Fprintf(os.Stderr, "Error creating destination directory %s: %v\n", destRulesDir, mkdirErr)
			os.Exit(1)
		}

		sourceFiles, err := findAllFiles(rulesSourceDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding source files in %s: %v\n", rulesSourceDir, err)
			os.Exit(1)
		}

		// Filter out ignored files
		sourceFiles = filterIgnoredFiles(sourceFiles, ignorePatterns, rulesSourceDir)

		// Clean up extra files in destination that don't exist in source
		if err := cleanupExtraFiles(sourceFiles, rulesSourceDir, destRulesDir, ignorePatterns); err != nil {
			fmt.Fprintf(os.Stderr, "Error cleaning up extra files: %v\n", err)
			os.Exit(1)
		}

		// Copy files with proper directory structure
		for _, srcFileFullPath := range sourceFiles {
			dstFileFullPath, err := recreateDirectoryStructure(srcFileFullPath, rulesSourceDir, destRulesDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error recreating directory structure for %s: %v\n", srcFileFullPath, err)
				continue
			}

			// Get relative path for display
			relativePath, err := filepath.Rel(rulesSourceDir, srcFileFullPath)
			if err != nil {
				relativePath = filepath.Base(srcFileFullPath)
			}

			fileExistedBeforeCopy := true
			if _, err := os.Stat(dstFileFullPath); os.IsNotExist(err) {
				fileExistedBeforeCopy = false
			} else if err != nil {
				fmt.Fprintf(os.Stderr, "Error checking destination file %s: %v\n", relativePath, err)
				continue
			}

			// Check if files are different before copying
			shouldCopy := true
			if fileExistedBeforeCopy {
				equal, err := filesAreEqualBasedOnExtension(srcFileFullPath, dstFileFullPath, overwriteHeaders)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error comparing files %s: %v\n", relativePath, err)
					// Continue with copying in case of comparison error
				} else if equal {
					shouldCopy = false // Files are identical, no need to copy
				}
			}

			if shouldCopy {
				copyErr := copyFileBasedOnExtension(srcFileFullPath, dstFileFullPath, overwriteHeaders)
				if copyErr != nil {
					fmt.Fprintf(os.Stderr, "Error synchronizing file %s: %v\n", relativePath, copyErr)
				} else {
					if fileExistedBeforeCopy {
						fmt.Printf("%s%s %s%s\n", colorYellow, symbolUpdate, relativePath, colorReset)
					} else {
						fmt.Printf("%s%s %s%s\n", colorGreen, symbolAdd, relativePath, colorReset)
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

		if _, statErr := os.Stat(rulesSourceDirInProject); os.IsNotExist(statErr) {
			fmt.Fprintf(os.Stderr, "Error: Project rules directory %s not found. Nothing to push.\n", rulesSourceDirInProject)
			os.Exit(1)
		}

		// Load ignore patterns
		ignorePatterns, err := loadRuleignore(rulesEnvDir, ignoreFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading ignore patterns: %v\n", err)
			os.Exit(1)
		}

		projectFiles, err := findAllFiles(rulesSourceDirInProject)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding files in project rules directory %s: %v\n", rulesSourceDirInProject, err)
			os.Exit(1)
		}

		// Filter out ignored files
		projectFiles = filterIgnoredFiles(projectFiles, ignorePatterns, rulesSourceDirInProject)

		if mkdirErr := os.MkdirAll(rulesEnvDir, os.ModePerm); mkdirErr != nil {
			fmt.Fprintf(os.Stderr, "Error creating destination directory %s: %v\n", rulesEnvDir, mkdirErr)
			os.Exit(1)
		}

		// Clean up extra files in destination that don't exist in source
		if err := cleanupExtraFiles(projectFiles, rulesSourceDirInProject, rulesEnvDir, ignorePatterns); err != nil {
			fmt.Fprintf(os.Stderr, "Error cleaning up extra files: %v\n", err)
			os.Exit(1)
		}

		filesCopiedOrUpdated := false
		if len(projectFiles) == 0 {
			// No files to process, but we might have deleted some files above
			filesCopiedOrUpdated = false
		} else {
			// Copy files with proper directory structure
			for _, srcFileFullPath := range projectFiles {
				dstFileFullPath, err := recreateDirectoryStructure(srcFileFullPath, rulesSourceDirInProject, rulesEnvDir)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error recreating directory structure for %s: %v\n", srcFileFullPath, err)
					continue
				}

				// Get relative path for display
				relativePath, err := filepath.Rel(rulesSourceDirInProject, srcFileFullPath)
				if err != nil {
					relativePath = filepath.Base(srcFileFullPath)
				}

				fileExists := true
				if _, statErr := os.Stat(dstFileFullPath); os.IsNotExist(statErr) {
					fileExists = false
				} else if statErr != nil {
					fmt.Fprintf(os.Stderr, "Error checking destination file %s in %s: %v\n", relativePath, rulesEnvDir, statErr)
					continue
				}

				// Check if files are different before copying
				shouldCopy := true
				if fileExists {
					equal, err := filesAreEqualBasedOnExtension(srcFileFullPath, dstFileFullPath, overwriteHeaders)
					if err != nil {
						fmt.Fprintf(os.Stderr, "Error comparing files %s: %v\n", relativePath, err)
						// Continue with copying in case of comparison error
					} else if equal {
						shouldCopy = false // Files are identical, no need to copy
					}
				}

				if shouldCopy {
					copyErr := copyFileBasedOnExtension(srcFileFullPath, dstFileFullPath, overwriteHeaders)
					if copyErr != nil {
						fmt.Fprintf(os.Stderr, "Error synchronizing file %s to %s: %v\n", relativePath, rulesEnvDir, copyErr)
					} else {
						if fileExists {
							fmt.Printf("%s%s %s (to %s)%s\n", colorYellow, symbolUpdate, relativePath, filepath.Base(rulesEnvDir), colorReset)
						} else {
							fmt.Printf("%s%s %s (to %s)%s\n", colorGreen, symbolAdd, relativePath, filepath.Base(rulesEnvDir), colorReset)
						}
						filesCopiedOrUpdated = true
					}
				}
			}
		}

		// Only commit if we have changes
		if filesCopiedOrUpdated {
			rulesEnvRepoRoot, err := getGitRootDir(rulesEnvDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error finding Git repository root for %s: %v\n", rulesEnvDir, err)
				fmt.Fprintf(os.Stderr, "Commit will not be performed. Please commit changes in %s manually if needed.\n", rulesEnvDir)
			} else {
				commitMessage := "Sync cursor rules: updated from project " + filepath.Base(projectGitRoot)
				_ = commitChanges(rulesEnvRepoRoot, commitMessage, gitWithoutPush) //nolint:errcheck
				// commitChanges handles its own error printing
			}
		}
	},
}

func init() {
	// Add flags to both commands
	pullCmd.Flags().StringVar(&rulesDir, "rules-dir", "", "Path to rules directory (overrides CURSOR_RULES_DIR env var)")
	pullCmd.Flags().StringVar(&ignoreFiles, "ignore-files", "", "Comma-separated list of files to ignore")
	pullCmd.Flags().BoolVar(&overwriteHeaders, "overwrite-headers", false, "Overwrite headers instead of preserving them")

	pushCmd.Flags().StringVar(&rulesDir, "rules-dir", "", "Path to rules directory (overrides CURSOR_RULES_DIR env var)")
	pushCmd.Flags().BoolVar(&gitWithoutPush, "git-without-push", false, "Commit changes but don't push to remote")
	pushCmd.Flags().StringVar(&ignoreFiles, "ignore-files", "", "Comma-separated list of files to ignore")
	pushCmd.Flags().BoolVar(&overwriteHeaders, "overwrite-headers", false, "Overwrite headers instead of preserving them")

	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(versionCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
