package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "cursor-rules-syncer",
	Short: "A CLI tool to sync cursor rules",
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pulls rules from the source directory to the current git project's .cursor/rules directory, deleting extra files in the project.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executing pull command...")

		rulesSourceDir, err := getRulesSourceDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Rules source directory: %s\n", rulesSourceDir)

		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		gitRoot, err := getGitRootDir(currentDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Git project root: %s\n", gitRoot)

		destRulesDir := filepath.Join(gitRoot, cursorDirName, rulesDirName)
		fmt.Printf("Destination directory for rules in project: %s\n", destRulesDir)

		if err := os.MkdirAll(destRulesDir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating destination directory %s: %v\n", destRulesDir, err)
			os.Exit(1)
		}

		// 1. Get all .mdc files from CURSOR_RULES_DIR (source)
		sourceMdcFiles, err := findMdcFiles(rulesSourceDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding source files: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d source files in %s\n", len(sourceMdcFiles), rulesSourceDir)

		sourceFilesMap := make(map[string]string) // basename -> full path
		for _, f := range sourceMdcFiles {
			sourceFilesMap[filepath.Base(f)] = f
		}

		// 2. Get all .mdc files from project's .cursor/rules (destination)
		destMdcFiles, err := findMdcFiles(destRulesDir)
		if err != nil {
			// If the directory doesn't exist or no files, it's not necessarily an error here,
			// as we might be pulling for the first time.
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error finding files in destination directory %s: %v\n", destRulesDir, err)
				// Decide if this is fatal. For now, we'll attempt to continue.
			}
		}
		fmt.Printf("Found %d files in destination %s before sync.\n", len(destMdcFiles), destRulesDir)

		// 3. Delete files in destination that are not in source
		for _, destFile := range destMdcFiles {
			fileName := filepath.Base(destFile)
			if _, existsInSource := sourceFilesMap[fileName]; !existsInSource {
				fmt.Printf("Deleting %s as it's not in the source directory.\n", destFile)
				if err := os.Remove(destFile); err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting file %s: %v\n", destFile, err)
				}
			}
		}

		// 4. Copy each .mdc file from source to destination, preserving headers
		if len(sourceMdcFiles) == 0 {
			fmt.Printf("No .mdc files found in source %s to pull.\n", rulesSourceDir)
		}
		for _, srcFile := range sourceMdcFiles {
			fileName := filepath.Base(srcFile)
			dstFile := filepath.Join(destRulesDir, fileName)
			fmt.Printf("Copying/Updating %s to %s...\n", srcFile, dstFile)
			if err := copyFileWithHeaderPreservation(srcFile, dstFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error copying file %s: %v\n", srcFile, err)
			}
		}
		fmt.Println("Pull command finished successfully.")
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes rules from the current git project's .cursor/rules directory to the source directory, deleting extra files in the source, and commits changes",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executing push command...")

		// 1. Get the path to the rules destination directory from env (CURSOR_RULES_DIR)
		rulesEnvDir, err := getRulesSourceDir() // This is the destination directory in this context
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Rules destination directory (from env): %s\n", rulesEnvDir)

		// 2. Find the git project root from the current working directory
		currentDir, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
			os.Exit(1)
		}
		projectGitRoot, err := getGitRootDir(currentDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Current Git project root: %s\n", projectGitRoot)

		// 3. Construct the path to the source rules directory in the project: <project_git_root>/.cursor/rules
		rulesSourceDirInProject := filepath.Join(projectGitRoot, cursorDirName, rulesDirName)
		fmt.Printf("Source rules directory in project: %s\n", rulesSourceDirInProject)

		// 4. Check if the <project_git_root>/.cursor/rules directory exists
		if _, err := os.Stat(rulesSourceDirInProject); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Project rules directory %s not found. Nothing to push.\n", rulesSourceDirInProject)
			os.Exit(1)
		}

		// 5. Get all .mdc files from <project_git_root>/.cursor/rules (source for push)
		projectMdcFiles, err := findMdcFiles(rulesSourceDirInProject)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding files in project rules directory: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Found %d files in project rules directory %s to push.\n", len(projectMdcFiles), rulesSourceDirInProject)

		projectFilesMap := make(map[string]string) // basename -> full path
		for _, f := range projectMdcFiles {
			projectFilesMap[filepath.Base(f)] = f
		}

		// 6. Create the destination directory (from env) if it doesn't exist
		if err := os.MkdirAll(rulesEnvDir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating destination directory %s: %v\n", rulesEnvDir, err)
			os.Exit(1)
		}

		// 7. Get all .mdc files from CURSOR_RULES_DIR (destination for push, to check for deletions)
		destEnvMdcFiles, err := findMdcFiles(rulesEnvDir)
		if err != nil {
			if !os.IsNotExist(err) {
				fmt.Fprintf(os.Stderr, "Error finding files in environment destination directory %s: %v\n", rulesEnvDir, err)
			}
		}
		fmt.Printf("Found %d files in environment destination %s before sync.\n", len(destEnvMdcFiles), rulesEnvDir)

		// 8. Delete files in CURSOR_RULES_DIR that are not in the project's rules directory
		for _, destFile := range destEnvMdcFiles {
			fileName := filepath.Base(destFile)
			if _, existsInProject := projectFilesMap[fileName]; !existsInProject {
				fmt.Printf("Deleting %s from %s as it's not in the project's rules directory.\n", fileName, rulesEnvDir)
				if err := os.Remove(destFile); err != nil {
					fmt.Fprintf(os.Stderr, "Error deleting file %s: %v\n", destFile, err)
				}
			}
		}

		// 9. Copy each .mdc file from project to CURSOR_RULES_DIR, preserving headers
		if len(projectMdcFiles) == 0 {
			fmt.Printf("No .mdc files found in project %s to push.\n", rulesSourceDirInProject)
			// If no files to push, but deletions might have occurred, still attempt commit
		}

		for _, srcFile := range projectMdcFiles {
			fileName := filepath.Base(srcFile)
			dstFile := filepath.Join(rulesEnvDir, fileName)
			fmt.Printf("Copying/Updating %s to %s...\n", srcFile, dstFile)
			if err := copyFileWithHeaderPreservation(srcFile, dstFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error copying file %s: %v\n", srcFile, err)
			}
		}

		// 10. Go to the Git repository root of the CURSOR_RULES_DIR and commit changes.
		rulesEnvRepoRoot, err := getGitRootDir(rulesEnvDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding Git repository root for %s: %v\n", rulesEnvDir, err)
			fmt.Fprintf(os.Stderr, "Commit will not be performed. Please copy files and commit manually if needed.\n")
		} else {
			commitMessage := "Sync cursor rules: updated from project " + filepath.Base(projectGitRoot)
			fmt.Printf("Attempting to commit in %s with message: '%s'\n", rulesEnvRepoRoot, commitMessage)
			if err := commitChanges(rulesEnvRepoRoot, commitMessage); err != nil {
				fmt.Fprintf(os.Stderr, "Error during commit in %s: %v\n", rulesEnvRepoRoot, err)
			} else {
				fmt.Println("Changes committed successfully.")
			}
		}

		fmt.Println("Push command finished successfully.")
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(pushCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		// cobra already prints the error, so just exit
		os.Exit(1)
	}
}
