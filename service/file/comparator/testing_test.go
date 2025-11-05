package comparator_test

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/yanodintsovmercuryo/cursync/service/file/comparator"
	"github.com/yanodintsovmercuryo/cursync/service/file/comparator/mocks"
)

type fixture struct {
	comparator *comparator.Comparator

	fileOpsMock *mocks.MockfileOps
}

func setUp(t *testing.T) (*fixture, func()) {
	t.Helper()
	ctrl := gomock.NewController(t)
	fileOpsMock := mocks.NewMockfileOps(ctrl)

	return &fixture{
		comparator:  comparator.NewComparator(fileOpsMock),
		fileOpsMock: fileOpsMock,
	}, ctrl.Finish
}

