package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/yanodintsovmercuryo/cursor-rules-syncer/models"
	"github.com/yanodintsovmercuryo/cursor-rules-syncer/pkg/file_ops"
	"github.com/yanodintsovmercuryo/cursor-rules-syncer/pkg/path"
	"github.com/yanodintsovmercuryo/cursor-rules-syncer/service/file_filter"
	"github.com/yanodintsovmercuryo/cursor-rules-syncer/service/output"
	"github.com/yanodintsovmercuryo/cursor-rules-syncer/service/sync"
)

// version
const version = "v1.0.0"

// Адаптеры для реальных реализаций из pkg пакетов
type fileOpsAdapter struct {
	file_ops.FileOps
}

type pathUtilsAdapter struct {
	*path.PathUtils
}

// Реализация file_filter.fileOps интерфейса
func (a *fileOpsAdapter) FindAllFiles(dir string) ([]string, error) {
	return a.FileOps.FindAllFiles(dir)
}

func (a *fileOpsAdapter) ReadFileNormalized(filePath string) (string, error) {
	return a.FileOps.ReadFileNormalized(filePath)
}

func (a *fileOpsAdapter) WriteFile(filePath, content string, perm os.FileMode) error {
	return a.FileOps.WriteFile(filePath, content, perm)
}

func (a *fileOpsAdapter) FileExists(filePath string) (bool, error) {
	return a.FileOps.FileExists(filePath)
}

func (a *fileOpsAdapter) CopyFile(srcPath, dstPath string) error {
	return a.FileOps.CopyFile(srcPath, dstPath)
}

func (a *fileOpsAdapter) RemoveFile(filePath string) error {
	return a.FileOps.RemoveFile(filePath)
}

func (a *fileOpsAdapter) MkdirAll(path string, perm os.FileMode) error {
	return a.FileOps.MkdirAll(path, perm)
}

func (a *fileOpsAdapter) GetCurrentDir() (string, error) {
	return a.FileOps.GetCurrentDir()
}

func (a *fileOpsAdapter) Stat(filePath string) (os.FileInfo, error) {
	return a.FileOps.Stat(filePath)
}

// Реализация file_filter.pathUtils интерфейса
func (a *pathUtilsAdapter) RecreateDirectoryStructure(srcPath, srcBase, dstBase string) (string, error) {
	return a.PathUtils.RecreateDirectoryStructure(srcPath, srcBase, dstBase)
}

func (a *pathUtilsAdapter) GetRelativePath(filePath, baseDir string) (string, error) {
	return a.PathUtils.GetRelativePath(filePath, baseDir)
}

func (a *pathUtilsAdapter) NormalizePath(filePath string) string {
	return a.PathUtils.NormalizePath(filePath)
}

func (a *pathUtilsAdapter) GetDirectory(filePath string) string {
	return a.PathUtils.GetDirectory(filePath)
}

func (a *pathUtilsAdapter) GetBaseName(filePath string) string {
	return a.PathUtils.GetBaseName(filePath)
}

// Wrapper структура для sync.fileFilter интерфейса
type syncFileFilterWrapper struct {
	ff *file_filter.FileFilterService
}

func (s *syncFileFilterWrapper) GetFilePatterns(flagValue, envVarName string) ([]string, error) {
	return s.ff.GetFilePatterns(flagValue, envVarName)
}

func (s *syncFileFilterWrapper) FindFilesByPatterns(dir string, patterns []string) ([]string, error) {
	return s.ff.FindFilesByPatterns(dir, patterns)
}

func (s *syncFileFilterWrapper) CleanupExtraFilesByPatterns(srcFiles []string, srcBase, dstBase string, patterns []string) error {
	return s.ff.CleanupExtraFilesByPatterns(srcFiles, srcBase, dstBase, patterns)
}

func (s *syncFileFilterWrapper) GetEffectivePatterns(patterns []string) []string {
	return s.ff.GetEffectivePatterns(patterns)
}

func (s *syncFileFilterWrapper) ValidatePatterns(patterns []string) error {
	return s.ff.ValidatePatterns(patterns)
}

func (s *syncFileFilterWrapper) AnalyzePatternMatching(files []string, baseDir string, patterns []string) interface{} {
	// Возвращаем простую структуру stats без привязки к конкретному типу
	return struct {
		TotalFiles      int
		MatchedFiles    int
		MatchedPatterns map[string]int
	}{
		TotalFiles:      len(files),
		MatchedFiles:    0,
		MatchedPatterns: make(map[string]int),
	}
}

// Simple implementations for sync interface stubs
type simpleGitOps struct{}

func (g *simpleGitOps) GetGitRootDir(startDir string) (string, error) {
	return startDir, nil
}

func (g *simpleGitOps) CommitChanges(repoDir, commitMessage string, withoutPush bool) error {
	return nil
}

func main() {
	// Initialize services
	outputService := output.NewOutputService()
	
	// Create implementations using real pkg implementations
	fileOpsImpl := file_ops.NewFileOps()
	pathUtilsImpl := path.NewPathUtils()
	
	fileFilterService := file_filter.NewFileFilterService(outputService, &fileOpsAdapter{fileOpsImpl}, &pathUtilsAdapter{pathUtilsImpl})
	syncFileFilterService := &syncFileFilterWrapper{fileFilterService}
	
	syncService := sync.NewSyncService(
		outputService,        // outputService interface
		syncFileFilterService, // sync.fileFilter interface
		&fileOpsAdapter{fileOpsImpl}, // fileOps interface
		&pathUtilsAdapter{pathUtilsImpl}, // pathUtils interface
		&simpleGitOps{},       // gitOps interface (still stub)
	)

	app := &cli.App{
		Name:  "cursor-rules-syncer",
		Usage: "A CLI tool to sync cursor rules",
		Commands: []*cli.Command{
			{
				Name:  "pull",
				Usage: "Pulls rules from the source directory to the current git project's .cursor/rules directory, deleting extra files in the project.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "rules-dir",
						Usage: "Path to rules directory (overrides CURSOR_RULES_DIR env var)",
					},
					&cli.BoolFlag{
						Name:  "overwrite-headers",
						Usage: "Overwrite headers instead of preserving them",
					},
					&cli.StringFlag{
						Name:  "file-patterns",
						Usage: "Comma-separated file patterns to sync (e.g., 'local_*.mdc,translate/*.md') (overrides CURSOR_RULES_PATTERNS env var)",
					},
				},
				Action: func(c *cli.Context) error {
					options := &models.SyncOptions{
						RulesDir:         c.String("rules-dir"),
						GitWithoutPush:   false, // Not used in pull
						OverwriteHeaders: c.Bool("overwrite-headers"),
						FilePatterns:     c.String("file-patterns"),
					}

					_, err := syncService.PullRules(options)
					if err != nil {
						outputService.PrintFatalf("Error: %v", err)
					}
					return nil
				},
			},
			{
				Name:  "push",
				Usage: "Pushes rules from the current git project's .cursor/rules directory to the source directory, deleting extra files in the source, and commits changes",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "rules-dir",
						Usage: "Path to rules directory (overrides CURSOR_RULES_DIR env var)",
					},
					&cli.BoolFlag{
						Name:  "git-without-push",
						Usage: "Commit changes but don't push to remote",
					},
					&cli.BoolFlag{
						Name:  "overwrite-headers",
						Usage: "Overwrite headers instead of preserving them",
					},
					&cli.StringFlag{
						Name:  "file-patterns",
						Usage: "Comma-separated file patterns to sync (e.g., 'local_*.mdc,translate/*.md') (overrides CURSOR_RULES_PATTERNS env var)",
					},
				},
				Action: func(c *cli.Context) error {
					options := &models.SyncOptions{
						RulesDir:         c.String("rules-dir"),
						GitWithoutPush:   c.Bool("git-without-push"),
						OverwriteHeaders: c.Bool("overwrite-headers"),
						FilePatterns:     c.String("file-patterns"),
					}

					_, err := syncService.PushRules(options)
					if err != nil {
						outputService.PrintFatalf("Error: %v", err)
					}
					return nil
				},
			},
			{
				Name:  "version",
				Usage: "Print the version number",
				Action: func(c *cli.Context) error {
					fmt.Println(version)
					return nil
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		os.Exit(1)
	}
}
