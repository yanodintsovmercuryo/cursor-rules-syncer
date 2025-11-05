package config

import (
	"fmt"
	"strings"

	"github.com/urfave/cli/v2"
	"github.com/yanodintsovmercuryo/cursync/pkg/config"
)

// parseBoolFromString parses boolean value from string
// Returns true for: "1", "true", "True", "TRUE"
// Returns false for: "0", "false", "False", "FALSE", "", " " (empty or whitespace)
func parseBoolFromString(s string) bool {
	s = strings.TrimSpace(s)
	switch strings.ToLower(s) {
	case "1", "true":
		return true
	case "0", "false", "":
		return false
	default:
		return false
	}
}

// UpdateConfig updates configuration based on CLI context
func (s *CfgService) UpdateConfig(ctx *cli.Context) error {
	cfg, err := s.configRepository.Load()
	if err != nil {
		s.output.PrintErrorf("Failed to load config: %v", err)
		return err
	}

	updated := false
	if s.updateStringFlag(ctx, cfg, FlagRulesDir, ConfigKeyRulesDir, "rules-dir") {
		updated = true
	}
	if s.updateStringFlag(ctx, cfg, FlagFilePatterns, ConfigKeyFilePatterns, "file-patterns") {
		updated = true
	}
	if s.updateBoolFlag(ctx, cfg, FlagOverwriteHeaders, ConfigKeyOverwriteHeaders, "overwrite-headers") {
		updated = true
	}
	if s.updateBoolFlag(ctx, cfg, FlagGitWithoutPush, ConfigKeyGitWithoutPush, "git-without-push") {
		updated = true
	}

	if updated {
		if err := s.configRepository.Save(cfg); err != nil {
			s.output.PrintErrorf("Failed to save config: %v", err)
			return err
		}
	}

	return nil
}

// updateStringFlag updates a string configuration flag
func (s *CfgService) updateStringFlag(ctx *cli.Context, cfg *config.Config, flagName, configKey, displayName string) bool {
	if !ctx.IsSet(flagName) {
		return false
	}

	val := ctx.String(flagName)
	if err := s.configRepository.Set(cfg, configKey, val); err != nil {
		s.output.PrintErrorf("Failed to set %s: %v", displayName, err)
		return false
	}
	if val == "" {
		fmt.Printf("Cleared %s\n", displayName)
	} else {
		fmt.Printf("Set %s to: %s\n", displayName, val)
	}
	return true
}

// updateBoolFlag updates a boolean configuration flag
func (s *CfgService) updateBoolFlag(ctx *cli.Context, cfg *config.Config, flagName, configKey, displayName string) bool {
	if !ctx.IsSet(flagName) {
		return false
	}

	valStr := ctx.String(flagName)
	val := parseBoolFromString(valStr)
	if err := s.configRepository.Set(cfg, configKey, val); err != nil {
		s.output.PrintErrorf("Failed to set %s: %v", displayName, err)
		return false
	}
	if val {
		fmt.Printf("Set %s to: true\n", displayName)
	} else {
		fmt.Printf("Set %s to: false\n", displayName)
	}
	return true
}
