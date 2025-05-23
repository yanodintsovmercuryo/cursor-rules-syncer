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
	Short: "Pulls rules from the source directory to the current git project's .cursor/rules directory",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Executing pull command...")

		// 1. Get the path to the rules source directory from env (CURSOR_RULES_DIR)
		rulesSourceDir, err := getRulesSourceDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Rules source directory: %s\n", rulesSourceDir)

		// 2. Find the git project root from the current working directory
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

		// 3. Construct the destination path: <git_root>/.cursor/rules
		destRulesDir := filepath.Join(gitRoot, cursorDirName, rulesDirName)
		fmt.Printf("Destination directory for rules: %s\n", destRulesDir)

		// 4. Create the destination directory if it doesn't exist
		if err := os.MkdirAll(destRulesDir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating destination directory %s: %v\n", destRulesDir, err)
			os.Exit(1)
		}

		// 5. Get all .mdc files from CURSOR_RULES_DIR
		sourceMdcFiles, err := findMdcFiles(rulesSourceDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(sourceMdcFiles) == 0 {
			fmt.Printf("No .mdc files found in %s.\n", rulesSourceDir)
			return
		}
		fmt.Printf("Found the following .mdc files to copy: %v\n", sourceMdcFiles)

		// 6. Copy each .mdc file to the destination directory, preserving headers
		for _, srcFile := range sourceMdcFiles {
			fileName := filepath.Base(srcFile)
			dstFile := filepath.Join(destRulesDir, fileName)
			fmt.Printf("Copying %s to %s...\n", srcFile, dstFile)
			if err := copyFileWithHeaderPreservation(srcFile, dstFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error copying file %s: %v\n", srcFile, err)
				// Optionally, decide to continue with other files or stop execution
				// os.Exit(1) // Uncomment to stop on error
			}
		}
		fmt.Println("Pull command finished successfully.")
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Pushes rules from the current git project's .cursor/rules directory to the source directory and commits changes",
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
			fmt.Fprintf(os.Stderr, "Error: Directory %s not found. Ensure 'pull' command was run or create the directory manually.\n", rulesSourceDirInProject)
			os.Exit(1)
		}

		// 5. Get all .mdc files from <project_git_root>/.cursor/rules
		projectMdcFiles, err := findMdcFiles(rulesSourceDirInProject)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if len(projectMdcFiles) == 0 {
			fmt.Printf("No .mdc files found in %s to push.\n", rulesSourceDirInProject)
			return
		}
		fmt.Printf("Found the following .mdc files to push: %v\n", projectMdcFiles)

		// 6. Create the destination directory (from env) if it doesn't exist
		if err := os.MkdirAll(rulesEnvDir, os.ModePerm); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating destination directory %s: %v\n", rulesEnvDir, err)
			os.Exit(1)
		}

		// 7. Copy each .mdc file to the destination directory (from env), preserving headers
		for _, srcFile := range projectMdcFiles {
			fileName := filepath.Base(srcFile)
			dstFile := filepath.Join(rulesEnvDir, fileName)
			fmt.Printf("Copying %s to %s...\n", srcFile, dstFile)
			if err := copyFileWithHeaderPreservation(srcFile, dstFile); err != nil {
				fmt.Fprintf(os.Stderr, "Error copying file %s: %v\n", srcFile, err)
				// os.Exit(1) // Uncomment to stop on error
			}
		}

		// 8. Go to the Git repository root of the CURSOR_RULES_DIR and commit changes.
		rulesEnvRepoRoot, err := getGitRootDir(rulesEnvDir) // Find git root from the rules env directory
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error finding Git repository root for %s: %v\n", rulesEnvDir, err)
			fmt.Fprintf(os.Stderr, "Commit will not be performed. Please copy files and commit manually if needed.\n")
			// Do not exit with error, as copying might have been successful
		} else {
			commitMessage := "Sync cursor rules: updated from project " + filepath.Base(projectGitRoot)
			fmt.Printf("Attempting to commit in %s with message: '%s'\n", rulesEnvRepoRoot, commitMessage)
			if err := commitChanges(rulesEnvRepoRoot, commitMessage); err != nil {
				fmt.Fprintf(os.Stderr, "Error during commit in %s: %v\n", rulesEnvRepoRoot, err)
			} else {
				fmt.Println("Changes committed successfully.") // Message updated as push is conditional
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
