package filter_test

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/yanodintsovmercuryo/cursync/pkg/output"
	"github.com/yanodintsovmercuryo/cursync/pkg/path"
	"github.com/yanodintsovmercuryo/cursync/service/file/filter"
	"github.com/yanodintsovmercuryo/cursync/service/file/filter/mocks"
)

type fixture struct {
	filter *filter.Filter

	fileOpsMock *mocks.MockfileOps
}

func setUp(t *testing.T) (*fixture, func()) {
	t.Helper()
	ctrl := gomock.NewController(t)
	fileOpsMock := mocks.NewMockfileOps(ctrl)

	outputImpl := output.NewOutput()
	pathUtilsImpl := path.NewPathUtils()

	filterImpl := filter.NewFilter(outputImpl, fileOpsMock, pathUtilsImpl)

	return &fixture{
		filter:      filterImpl,
		fileOpsMock: fileOpsMock,
	}, ctrl.Finish
}
