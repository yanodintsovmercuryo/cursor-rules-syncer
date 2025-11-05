package file_test

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/yanodintsovmercuryo/cursync/service/file"
	"github.com/yanodintsovmercuryo/cursync/service/file/mocks"
)

type fixture struct {
	fileService *file.FileService

	comparatorMock *mocks.MockcomparatorService
	copierMock     *mocks.MockcopierService
	filterMock     *mocks.MockfilterService
}

func setUp(t *testing.T) (*fixture, func()) {
	t.Helper()
	ctrl := gomock.NewController(t)
	comparatorMock := mocks.NewMockcomparatorService(ctrl)
	copierMock := mocks.NewMockcopierService(ctrl)
	filterMock := mocks.NewMockfilterService(ctrl)

	fileService := file.NewFileServiceWithMocks(comparatorMock, copierMock, filterMock)

	return &fixture{
		fileService:    fileService,
		comparatorMock: comparatorMock,
		copierMock:     copierMock,
		filterMock:     filterMock,
	}, ctrl.Finish
}
