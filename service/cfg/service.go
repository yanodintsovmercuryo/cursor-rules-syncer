package cfg

import (
	"github.com/urfave/cli/v2"
	"github.com/yanodintsovmercuryo/cursync/pkg/config"
)

// Flag names constants
const (
	FlagRulesDir         = "rules-dir"
	FlagFilePatterns     = "file-patterns"
	FlagOverwriteHeaders = "overwrite-headers"
	FlagGitWithoutPush   = "git-without-push"
)

// Flag aliases constants
const (
	FlagAliasRulesDir         = "d"
	FlagAliasFilePatterns     = "p"
	FlagAliasOverwriteHeaders = "o"
	FlagAliasGitWithoutPush   = "w"
)

// Config keys constants (for config.Set/Get)
const (
	ConfigKeyRulesDir         = "rules-dir"
	ConfigKeyFilePatterns     = "file-patterns"
	ConfigKeyOverwriteHeaders = "overwrite-headers"
	ConfigKeyGitWithoutPush   = "git-without-push"
)

// CfgService handles configuration and options creation
type CfgService struct {
	configManager configManager
	output        outputService
}

type outputService interface {
	PrintErrorf(format string, args ...interface{})
}

// NewCfgService creates a new CfgService
func NewCfgService(output outputService) *CfgService {
	return &CfgService{
		configManager: NewConfigRepository(),
		output:        output,
	}
}

// NewCfgServiceWithManager creates a new CfgService with custom config manager
func NewCfgServiceWithManager(manager configManager, output outputService) *CfgService {
	return &CfgService{
		configManager: manager,
		output:        output,
	}
}

// getStringValue returns flag value or config value or empty string
func (s *CfgService) getStringValue(ctx *cli.Context, flagName string, cfg *config.Config) string {
	if ctx.IsSet(flagName) {
		return ctx.String(flagName)
	}
	switch flagName {
	case FlagRulesDir:
		return cfg.RulesDir
	case FlagFilePatterns:
		return cfg.FilePatterns
	}
	return ""
}

// getBoolValue returns flag value or config value or false
func (s *CfgService) getBoolValue(ctx *cli.Context, flagName string, cfg *config.Config) bool {
	if ctx.IsSet(flagName) {
		return ctx.Bool(flagName)
	}
	switch flagName {
	case FlagOverwriteHeaders:
		return cfg.OverwriteHeaders
	case FlagGitWithoutPush:
		return cfg.GitWithoutPush
	}
	return false
}
