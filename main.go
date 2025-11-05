package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/yanodintsovmercuryo/cursync/models"
	"github.com/yanodintsovmercuryo/cursync/pkg/file_ops"
	"github.com/yanodintsovmercuryo/cursync/pkg/git"
	"github.com/yanodintsovmercuryo/cursync/pkg/output"
	"github.com/yanodintsovmercuryo/cursync/pkg/path"
	"github.com/yanodintsovmercuryo/cursync/service/file"
	"github.com/yanodintsovmercuryo/cursync/service/sync"
)

const version = "v1.0.0"

func main() {
	outputService := output.NewOutput()
	fileOpsImpl := file_ops.NewFileOps()
	pathUtilsImpl := path.NewPathUtils()
	gitOpsImpl := git.NewGit()
	fileServiceImpl := file.NewFileService(outputService, fileOpsImpl, pathUtilsImpl)

	syncService := sync.NewSyncService(
		outputService,
		fileOpsImpl,
		pathUtilsImpl,
		gitOpsImpl,
		fileServiceImpl,
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
						Name:    "rules-dir",
						Aliases: []string{"d"},
						Usage:   "Path to rules directory (overrides CURSOR_RULES_DIR env var)",
					},
					&cli.BoolFlag{
						Name:    "overwrite-headers",
						Aliases: []string{"o"},
						Usage:   "Overwrite headers instead of preserving them",
					},
					&cli.StringFlag{
						Name:    "file-patterns",
						Aliases: []string{"p"},
						Usage:   "Comma-separated file patterns to sync (e.g., 'local_*.mdc,translate/*.md') (overrides CURSOR_RULES_PATTERNS env var)",
					},
				},
				Action: func(c *cli.Context) error {
					options := &models.SyncOptions{
						RulesDir:         c.String("rules-dir"),
						GitWithoutPush:   false,
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
						Name:    "rules-dir",
						Aliases: []string{"d"},
						Usage:   "Path to rules directory (overrides CURSOR_RULES_DIR env var)",
					},
					&cli.BoolFlag{
						Name:    "git-without-push",
						Aliases: []string{"w"},
						Usage:   "Commit changes but don't push to remote",
					},
					&cli.BoolFlag{
						Name:    "overwrite-headers",
						Aliases: []string{"o"},
						Usage:   "Overwrite headers instead of preserving them",
					},
					&cli.StringFlag{
						Name:    "file-patterns",
						Aliases: []string{"p"},
						Usage:   "Comma-separated file patterns to sync (e.g., 'local_*.mdc,translate/*.md') (overrides CURSOR_RULES_PATTERNS env var)",
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
