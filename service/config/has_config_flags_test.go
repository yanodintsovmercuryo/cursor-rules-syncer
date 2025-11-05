package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	cfgService "github.com/yanodintsovmercuryo/cursync/service/config"
)

func TestCfgService_HasConfigFlags(t *testing.T) {
	t.Parallel()

	t.Run("returns true when rules-dir is set", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagRulesDir: "/test/rules",
		})

		result := f.cfgService.HasConfigFlags(ctx)

		require.True(t, result)
	})

	t.Run("returns true when file-patterns is set", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagFilePatterns: "*.mdc",
		})

		result := f.cfgService.HasConfigFlags(ctx)

		require.True(t, result)
	})

	t.Run("returns true when overwrite-headers is set", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagOverwriteHeaders: true,
		})

		result := f.cfgService.HasConfigFlags(ctx)

		require.True(t, result)
	})

	t.Run("returns true when git-without-push is set", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagGitWithoutPush: true,
		})

		result := f.cfgService.HasConfigFlags(ctx)

		require.True(t, result)
	})

	t.Run("returns false when no flags are set", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		ctx := createCLIContext(t, map[string]interface{}{})

		result := f.cfgService.HasConfigFlags(ctx)

		require.False(t, result)
	})
}
