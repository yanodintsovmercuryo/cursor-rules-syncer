package config

import (
	"github.com/urfave/cli/v2"
	"github.com/yanodintsovmercuryo/cursync/models"
)

// CreatePullOptions creates SyncOptions for pull command
func (s *CfgService) CreatePullOptions(ctx *cli.Context) *models.SyncOptions {
	cfg := s.configRepository.LoadOrDefault()
	return &models.SyncOptions{
		RulesDir:         s.getStringValue(ctx, FlagRulesDir, cfg),
		GitWithoutPush:   false,
		OverwriteHeaders: s.getBoolValue(ctx, FlagOverwriteHeaders, cfg),
		FilePatterns:     s.getStringValue(ctx, FlagFilePatterns, cfg),
	}
}
