package config_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/yanodintsovmercuryo/cursync/pkg/config"
)

func TestCfgService_ShowConfig(t *testing.T) {
	t.Parallel()

	t.Run("displays config values", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{
			RulesDir:         "/test/rules",
			FilePatterns:     "*.mdc",
			OverwriteHeaders: true,
			GitWithoutPush:   false,
		}
		f.configRepositoryMock.EXPECT().
			Load().
			Return(cfg, nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			GetAll(cfg).
			Return(map[string]interface{}{
				"rules_dir":         "/test/rules",
				"file_patterns":     "*.mdc",
				"overwrite_headers": true,
				"git_without_push":  false,
			}).
			Times(1)

		err := f.cfgService.ShowConfig()
		require.NoError(t, err)
	})

	t.Run("displays no values message when config is empty", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		cfg := &config.Config{}
		f.configRepositoryMock.EXPECT().
			Load().
			Return(cfg, nil).
			Times(1)
		f.configRepositoryMock.EXPECT().
			GetAll(cfg).
			Return(map[string]interface{}{
				"rules_dir":         "",
				"file_patterns":     "",
				"overwrite_headers": false,
				"git_without_push":  false,
			}).
			Times(1)

		err := f.cfgService.ShowConfig()
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

		err := f.cfgService.ShowConfig()
		require.Error(t, err)
		require.ErrorIs(t, err, loadErr)
	})
}
