package pattern_test

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/yanodintsovmercuryo/cursync/pkg/path"
	"github.com/yanodintsovmercuryo/cursync/service/file/pattern"
	"github.com/yanodintsovmercuryo/cursync/service/file/pattern/mocks"
)

type fixture struct {
	patternFilter *pattern.PatternFilterService

	pathUtilsMock *mocks.MockpathUtils
}

func setUp(t *testing.T) (*fixture, func()) {
	t.Helper()
	ctrl := gomock.NewController(t)
	pathUtilsMock := mocks.NewMockpathUtils(ctrl)

	return &fixture{
		patternFilter: pattern.NewPatternFilterService(pathUtilsMock),
		pathUtilsMock: pathUtilsMock,
	}, ctrl.Finish
}

func setUpWithRealPathUtils(t *testing.T) (*pattern.PatternFilterService, func()) {
	t.Helper()
	pathUtilsImpl := path.NewPathUtils()

	return pattern.NewPatternFilterService(pathUtilsImpl), func() {}
}
