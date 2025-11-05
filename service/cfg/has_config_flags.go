package cfg

import (
	"github.com/urfave/cli/v2"
)

// HasConfigFlags checks if any config flags are set
func (s *CfgService) HasConfigFlags(ctx *cli.Context) bool {
	return ctx.IsSet(FlagRulesDir) || ctx.IsSet(FlagFilePatterns) || ctx.IsSet(FlagOverwriteHeaders) || ctx.IsSet(FlagGitWithoutPush)
}
