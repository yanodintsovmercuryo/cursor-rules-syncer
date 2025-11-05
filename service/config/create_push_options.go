package config

import (
	"github.com/urfave/cli/v2"
	"github.com/yanodintsovmercuryo/cursync/models"
)

// CreatePushOptions creates SyncOptions for push command
func (s *CfgService) CreatePushOptions(ctx *cli.Context) *models.SyncOptions {
	cfg := s.configRepository.LoadOrDefault()
	return &models.SyncOptions{
		RulesDir:         s.getStringValue(ctx, FlagRulesDir, cfg),
		GitWithoutPush:   s.getBoolValue(ctx, FlagGitWithoutPush, cfg),
		OverwriteHeaders: s.getBoolValue(ctx, FlagOverwriteHeaders, cfg),
		FilePatterns:     s.getStringValue(ctx, FlagFilePatterns, cfg),
	}
}
