package config_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yanodintsovmercuryo/cursync/pkg/config"
	cfgService "github.com/yanodintsovmercuryo/cursync/service/config"
)

func TestCfgService_UpdateConfig(t *testing.T) {
	t.Parallel()

	t.Run("updates rules-dir", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{}
		f.configRepositoryMock.EXPECT().
			Load().
			Return(cfg, nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Set(cfg, cfgService.ConfigKeyRulesDir, "/test/rules").
			Return(nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Save(cfg).
			Return(nil).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagRulesDir: "/test/rules",
		})

		err := f.cfgService.UpdateConfig(ctx)
		require.NoError(t, err)
	})

	t.Run("clears rules-dir when empty value", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{}
		f.configRepositoryMock.EXPECT().
			Load().
			Return(cfg, nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Set(cfg, cfgService.ConfigKeyRulesDir, "").
			Return(nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Save(cfg).
			Return(nil).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagRulesDir: "",
		})

		err := f.cfgService.UpdateConfig(ctx)
		require.NoError(t, err)
	})

	t.Run("updates file-patterns", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{}
		f.configRepositoryMock.EXPECT().
			Load().
			Return(cfg, nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Set(cfg, cfgService.ConfigKeyFilePatterns, "*.mdc").
			Return(nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Save(cfg).
			Return(nil).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagFilePatterns: "*.mdc",
		})

		err := f.cfgService.UpdateConfig(ctx)
		require.NoError(t, err)
	})

	t.Run("updates overwrite-headers", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{}
		f.configRepositoryMock.EXPECT().
			Load().
			Return(cfg, nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Set(cfg, cfgService.ConfigKeyOverwriteHeaders, true).
			Return(nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Save(cfg).
			Return(nil).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagOverwriteHeaders: "true",
		})

		err := f.cfgService.UpdateConfig(ctx)
		require.NoError(t, err)
	})

	t.Run("updates overwrite-headers with false", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{}
		f.configRepositoryMock.EXPECT().
			Load().
			Return(cfg, nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Set(cfg, cfgService.ConfigKeyOverwriteHeaders, false).
			Return(nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Save(cfg).
			Return(nil).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagOverwriteHeaders: "false",
		})

		err := f.cfgService.UpdateConfig(ctx)
		require.NoError(t, err)
	})

	t.Run("updates git-without-push", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{}
		f.configRepositoryMock.EXPECT().
			Load().
			Return(cfg, nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Set(cfg, cfgService.ConfigKeyGitWithoutPush, true).
			Return(nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			Save(cfg).
			Return(nil).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagGitWithoutPush: "1",
		})

		err := f.cfgService.UpdateConfig(ctx)
		require.NoError(t, err)
	})

	t.Run("returns error when load fails", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		loadErr := errors.New("load error")
		f.configRepositoryMock.EXPECT().
			Load().
			Return(nil, loadErr).
			Times(1)
		f.outputMock.EXPECT().
			PrintErrorf("Failed to load config: %v", loadErr).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagRulesDir: "/test/rules",
		})

		err := f.cfgService.UpdateConfig(ctx)
		require.Error(t, err)
		require.Equal(t, loadErr, err)
	})
}
