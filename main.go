package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
	"github.com/yanodintsovmercuryo/cursync/pkg/config"
	"github.com/yanodintsovmercuryo/cursync/pkg/file_ops"
	"github.com/yanodintsovmercuryo/cursync/pkg/git"
	"github.com/yanodintsovmercuryo/cursync/pkg/output"
	"github.com/yanodintsovmercuryo/cursync/pkg/path"
	cfgService "github.com/yanodintsovmercuryo/cursync/service/config"
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

	cfgServiceInstance := cfgService.NewCfgService(config.NewConfigRepository(), outputService)

	app := &cli.App{
		Name:  "cursor-rules-syncer",
		Usage: "A CLI tool to sync cursor rules",
		Commands: []*cli.Command{
			{
				Name:  "pull",
				Usage: "Pulls rules from the source directory to the current git project's .cursor/rules directory, deleting extra files in the project.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    cfgService.FlagRulesDir,
						Aliases: []string{cfgService.FlagAliasRulesDir},
						Usage:   "Path to rules directory (overrides CURSOR_RULES_DIR env var)",
					},
					&cli.BoolFlag{
						Name:    cfgService.FlagOverwriteHeaders,
						Aliases: []string{cfgService.FlagAliasOverwriteHeaders},
						Usage:   "Overwrite headers instead of preserving them",
					},
					&cli.StringFlag{
						Name:    cfgService.FlagFilePatterns,
						Aliases: []string{cfgService.FlagAliasFilePatterns},
						Usage:   "Comma-separated file patterns to sync (e.g., 'local_*.mdc,translate/*.md') (overrides CURSOR_RULES_PATTERNS env var)",
					},
				},
				Action: func(c *cli.Context) error {
					options := cfgServiceInstance.CreatePullOptions(c)

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
						Name:    cfgService.FlagRulesDir,
						Aliases: []string{cfgService.FlagAliasRulesDir},
						Usage:   "Path to rules directory (overrides CURSOR_RULES_DIR env var)",
					},
					&cli.BoolFlag{
						Name:    cfgService.FlagGitWithoutPush,
						Aliases: []string{cfgService.FlagAliasGitWithoutPush},
						Usage:   "Commit changes but don't push to remote",
					},
					&cli.BoolFlag{
						Name:    cfgService.FlagOverwriteHeaders,
						Aliases: []string{cfgService.FlagAliasOverwriteHeaders},
						Usage:   "Overwrite headers instead of preserving them",
					},
					&cli.StringFlag{
						Name:    cfgService.FlagFilePatterns,
						Aliases: []string{cfgService.FlagAliasFilePatterns},
						Usage:   "Comma-separated file patterns to sync (e.g., 'local_*.mdc,translate/*.md') (overrides CURSOR_RULES_PATTERNS env var)",
					},
				},
				Action: func(c *cli.Context) error {
					options := cfgServiceInstance.CreatePushOptions(c)

					_, err := syncService.PushRules(options)
					if err != nil {
						outputService.PrintFatalf("Error: %v", err)
					}
					return nil
				},
			},
			{
				Name:        "cfg",
				Usage:       "Manage configuration values",
				Description: "Set, get, or clear configuration values. For string flags, empty value clears the default. For bool flags, use --flag=false to clear.",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    cfgService.FlagRulesDir,
						Aliases: []string{cfgService.FlagAliasRulesDir},
						Usage:   "Set default rules directory (empty value clears it)",
					},
					&cli.StringFlag{
						Name:    cfgService.FlagFilePatterns,
						Aliases: []string{cfgService.FlagAliasFilePatterns},
						Usage:   "Set default file patterns (empty value clears it)",
					},
					&cli.StringFlag{
						Name:    cfgService.FlagOverwriteHeaders,
						Aliases: []string{cfgService.FlagAliasOverwriteHeaders},
						Usage:   "Set default overwrite-headers flag (use 'true', '1', 'false', '0', or empty to clear)",
					},
					&cli.StringFlag{
						Name:    cfgService.FlagGitWithoutPush,
						Aliases: []string{cfgService.FlagAliasGitWithoutPush},
						Usage:   "Set default git-without-push flag (use 'true', '1', 'false', '0', or empty to clear)",
					},
				},
				Action: func(c *cli.Context) error {
					if !cfgServiceInstance.HasConfigFlags(c) {
						return cfgServiceInstance.ShowConfig()
					}
					return cfgServiceInstance.UpdateConfig(c)
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
