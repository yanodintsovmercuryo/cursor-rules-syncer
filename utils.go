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

// readFileNormalized читает файл и нормализует line endings
func readFileNormalized(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	// Нормализуем line endings - приводим к LF
	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	normalized = strings.ReplaceAll(normalized, "\r", "\n")

	// Убираем trailing whitespace в конце файла
	normalized = strings.TrimRight(normalized, " \t\n")

	return normalized, nil
}

// copyFileWithHeaderPreservation copies a file, preserving the header of the destination file if it exists.
// The header is defined by lines between --- at the beginning of the file.
func copyFileWithHeaderPreservation(srcPath, dstPath string) error {
	// Читаем содержимое source файла полностью
	srcContent, err := os.ReadFile(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", srcPath, err)
	}

	// Нормализуем line endings в source content
	srcContentStr := strings.ReplaceAll(string(srcContent), "\r\n", "\n")
	srcContentStr = strings.ReplaceAll(srcContentStr, "\r", "\n")

	// Проверяем существование destination файла и извлекаем header
	existingHeader, err := extractExistingHeader(dstPath)
	if err != nil {
		return err
	}

	// Если есть существующий header в destination, извлекаем content из source без его header
	var finalContent string
	if existingHeader != "" {
		srcContentWithoutHeader := removeHeaderFromContent(srcContentStr)
		finalContent = existingHeader + srcContentWithoutHeader
	} else {
		finalContent = srcContentStr
	}

	// Убеждаемся что файл заканчивается переводом строки
	if !strings.HasSuffix(finalContent, "\n") {
		finalContent += "\n"
	}

	// Записываем финальное содержимое
	err = os.WriteFile(dstPath, []byte(finalContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dstPath, err)
	}

	return nil
}

// removeHeaderFromContent удаляет header из content string
func removeHeaderFromContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	// Проверяем начинается ли с header separator
	if lines[0] != headerSeparator {
		return content // Нет header, возвращаем как есть
	}

	// Ищем закрывающий separator
	for i := 1; i < len(lines); i++ {
		if lines[i] == headerSeparator {
			// Найден закрывающий separator, возвращаем содержимое после него
			if i+1 < len(lines) {
				remainingLines := lines[i+1:]
				// Убираем leading empty lines
				for len(remainingLines) > 0 && strings.TrimSpace(remainingLines[0]) == "" {
					remainingLines = remainingLines[1:]
				}
				return strings.Join(remainingLines, "\n")
			}
			return "" // Header занимает весь файл
		}
		// Ограничиваем поиск header разумным количеством строк
		if i > 20 {
			break
		}
	}

	// Не найден закрывающий separator, возвращаем весь content
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

// extractHeaderFromContent извлекает header из content string
func extractHeaderFromContent(content string) string {
	// Нормализуем line endings
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")

	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != headerSeparator {
		return "" // Нет header
	}

	var headerLines []string
	headerLines = append(headerLines, lines[0]) // Добавляем первый separator

	// Ищем закрывающий separator
	for i := 1; i < len(lines) && i <= 20; i++ { // Ограничиваем поиск 20 строками
		headerLines = append(headerLines, lines[i])
		if lines[i] == headerSeparator {
			// Найден закрывающий separator, возвращаем header с переводом строки
			return strings.Join(headerLines, "\n") + "\n"
		}
	}

	return "" // Не найден правильный header
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
	// Используем git add -A вместо git add . для добавления всех изменений включая удаления
	addCmd := exec.Command("git", "add", "-A")
	addCmd.Dir = repoDir
	output, err := addCmd.CombinedOutput()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running 'git add -A' in %s: %s\n%v\n", repoDir, string(output), err)
		return err
	}

	// Проверяем статус для отладки
	statusCmd := exec.Command("git", "status", "--porcelain")
	statusCmd.Dir = repoDir
	statusOutput, statusErr := statusCmd.CombinedOutput()
	if statusErr != nil {
		fmt.Fprintf(os.Stderr, "Warning: could not check git status in %s: %v\n", repoDir, statusErr)
	} else {
		statusLines := strings.TrimSpace(string(statusOutput))
		if statusLines == "" {
			return nil // Нет изменений для коммита
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
