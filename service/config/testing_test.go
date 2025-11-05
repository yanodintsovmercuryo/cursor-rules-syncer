package config_test

import (
	"flag"
	"testing"

	"github.com/urfave/cli/v2"
	"go.uber.org/mock/gomock"

	cfgService "github.com/yanodintsovmercuryo/cursync/service/config"
	"github.com/yanodintsovmercuryo/cursync/service/config/mocks"
)

type fixture struct {
	cfgService *cfgService.CfgService

	configRepositoryMock *mocks.MockConfigRepositoryInterface
	outputMock           *mocks.MockoutputService
}

func setUp(t *testing.T) (*fixture, func()) {
	t.Helper()
	ctrl := gomock.NewController(t)
	configRepositoryMock := mocks.NewMockConfigRepositoryInterface(ctrl)
	outputMock := mocks.NewMockoutputService(ctrl)

	cfgService := cfgService.NewCfgService(configRepositoryMock, outputMock)

	return &fixture{
		cfgService:           cfgService,
		configRepositoryMock: configRepositoryMock,
		outputMock:           outputMock,
	}, ctrl.Finish
}

func createCLIContext(t *testing.T, flags map[string]interface{}) *cli.Context {
	t.Helper()
	app := &cli.App{
		Flags: []cli.Flag{
			&cli.StringFlag{Name: cfgService.FlagRulesDir, Aliases: []string{cfgService.FlagAliasRulesDir}},
			&cli.StringFlag{Name: cfgService.FlagFilePatterns, Aliases: []string{cfgService.FlagAliasFilePatterns}},
			&cli.StringFlag{Name: cfgService.FlagOverwriteHeaders, Aliases: []string{cfgService.FlagAliasOverwriteHeaders}},
			&cli.StringFlag{Name: cfgService.FlagGitWithoutPush, Aliases: []string{cfgService.FlagAliasGitWithoutPush}},
		},
	}

	args := []string{"app"}
	for key, val := range flags {
		switch v := val.(type) {
		case string:
			if v == "" {
				args = append(args, "--"+key+"=")
			} else {
				args = append(args, "--"+key+"="+v)
			}
		case bool:
			if v {
				args = append(args, "--"+key+"=true")
			} else {
				args = append(args, "--"+key+"=false")
			}
		}
	}

	set := flag.NewFlagSet("test", 0)
	for _, f := range app.Flags {
		f.Apply(set)
	}
	set.Parse(args[1:])

	ctx := cli.NewContext(app, set, nil)
	return ctx
}
