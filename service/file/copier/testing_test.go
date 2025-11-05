package copier_test

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/yanodintsovmercuryo/cursync/service/file/copier"
	"github.com/yanodintsovmercuryo/cursync/service/file/copier/mocks"
)

type fixture struct {
	copier *copier.Copier

	fileOpsMock *mocks.MockfileOps
}

func setUp(t *testing.T) (*fixture, func()) {
	t.Helper()
	ctrl := gomock.NewController(t)
	fileOpsMock := mocks.NewMockfileOps(ctrl)

	return &fixture{
		copier:      copier.NewCopier(fileOpsMock),
		fileOpsMock: fileOpsMock,
	}, ctrl.Finish
}
