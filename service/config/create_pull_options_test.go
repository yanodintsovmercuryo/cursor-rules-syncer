package config_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yanodintsovmercuryo/cursync/models"
	"github.com/yanodintsovmercuryo/cursync/pkg/config"
	cfgService "github.com/yanodintsovmercuryo/cursync/service/config"
)

func TestCfgService_CreatePullOptions(t *testing.T) {
	t.Parallel()

	t.Run("uses flag values when set", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{}
		f.configRepositoryMock.EXPECT().
			LoadOrDefault().
			Return(cfg).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagRulesDir:         "/custom/rules",
			cfgService.FlagFilePatterns:     "*.mdc",
			cfgService.FlagOverwriteHeaders: true,
		})

		result := f.cfgService.CreatePullOptions(ctx)

		expected := &models.SyncOptions{
			RulesDir:         "/custom/rules",
			FilePatterns:     "*.mdc",
			OverwriteHeaders: true,
			GitWithoutPush:   false,
		}

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("uses config values when flags not set", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{
			RulesDir:         "/default/rules",
			FilePatterns:     "default.mdc",
			OverwriteHeaders: true,
		}
		f.configRepositoryMock.EXPECT().
			LoadOrDefault().
			Return(cfg).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{})

		result := f.cfgService.CreatePullOptions(ctx)

		expected := &models.SyncOptions{
			RulesDir:         "/default/rules",
			FilePatterns:     "default.mdc",
			OverwriteHeaders: true,
			GitWithoutPush:   false,
		}

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("flag values override config values", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{
			RulesDir:         "/default/rules",
			FilePatterns:     "default.mdc",
			OverwriteHeaders: false,
		}
		f.configRepositoryMock.EXPECT().
			LoadOrDefault().
			Return(cfg).
			Times(1)

		ctx := createCLIContext(t, map[string]interface{}{
			cfgService.FlagRulesDir: "/override/rules",
		})

		result := f.cfgService.CreatePullOptions(ctx)

		expected := &models.SyncOptions{
			RulesDir:         "/override/rules",
			FilePatterns:     "default.mdc",
			OverwriteHeaders: false,
			GitWithoutPush:   false,
		}

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})
}
