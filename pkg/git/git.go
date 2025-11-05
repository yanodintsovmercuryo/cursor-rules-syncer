package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const cursorDirName = ".cursor"

// Git handles git operations
type Git struct{}

// NewGit creates a new Git instance
func NewGit() *Git {
	return &Git{}
}

// GetGitRootDir returns root directory by recursively searching for either git project root or .cursor folder
func (g *Git) GetGitRootDir(startDir string) (string, error) {
	absStartDir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}

	currentDir := absStartDir

	for {
		gitDir := filepath.Join(currentDir, ".git")
		if info, err := os.Stat(gitDir); err == nil && info.IsDir() {
			return currentDir, nil
		}

		cursorDir := filepath.Join(currentDir, cursorDirName)
		if info, err := os.Stat(cursorDir); err == nil && info.IsDir() {
			return currentDir, nil
		}

		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			return "", fmt.Errorf("failed to find git root or .cursor directory starting from %s: neither .git directory nor .cursor directory found in parent directories", absStartDir)
		}

		currentDir = parentDir
	}
}

// CommitChanges executes git add, commit and optionally push
func (g *Git) CommitChanges(repoDir, commitMessage string, withoutPush bool) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	defer func() {
		if chdirErr := os.Chdir(currentDir); chdirErr != nil {
			// Ignore error on restore - we've already done the work
			_ = chdirErr
		}
	}()

	if err := os.Chdir(repoDir); err != nil {
		return fmt.Errorf("failed to change directory to %s: %w", repoDir, err)
	}

	cmd := exec.Command("git", "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git add: %w", err)
	}

	cmd = exec.Command("git", "commit", "-m", commitMessage)
	if err := cmd.Run(); err != nil {
		exitError, ok := err.(*exec.ExitError)
		if ok && exitError.ExitCode() == 1 {
			return nil
		}
		return fmt.Errorf("failed to git commit: %w", err)
	}

	if !withoutPush {
		if err := g.pushIfRemoteExists(); err != nil {
			return err
		}
	}

	return nil
}

// pushIfRemoteExists executes git push only if remote origin exists
func (g *Git) pushIfRemoteExists() error {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	if err := cmd.Run(); err != nil {
		return nil
	}

	cmd = exec.Command("git", "push")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to git push: %w", err)
	}

	return nil
}
