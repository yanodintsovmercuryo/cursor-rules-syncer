package sync_test

import (
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/yanodintsovmercuryo/cursync/service/sync"
	syncMocks "github.com/yanodintsovmercuryo/cursync/service/sync/mocks"
)

type fixture struct {
	syncService *sync.SyncService

	outputMock      *syncMocks.MockoutputService
	fileOpsMock     *syncMocks.MockfileOps
	pathUtilsMock   *syncMocks.MockpathUtils
	gitOpsMock      *syncMocks.MockgitOps
	fileServiceMock *syncMocks.MockfileService
}

func setUp(t *testing.T) (*fixture, func()) {
	t.Helper()
	ctrl := gomock.NewController(t)
	outputMock := syncMocks.NewMockoutputService(ctrl)
	fileOpsMock := syncMocks.NewMockfileOps(ctrl)
	pathUtilsMock := syncMocks.NewMockpathUtils(ctrl)
	gitOpsMock := syncMocks.NewMockgitOps(ctrl)
	fileServiceMock := syncMocks.NewMockfileService(ctrl)

	// Use constructor for tests with mocks
	syncService := sync.NewSyncServiceWithMocks(outputMock, fileOpsMock, pathUtilsMock, gitOpsMock, fileServiceMock)

	return &fixture{
		syncService:     syncService,
		outputMock:      outputMock,
		fileOpsMock:     fileOpsMock,
		pathUtilsMock:   pathUtilsMock,
		gitOpsMock:      gitOpsMock,
		fileServiceMock: fileServiceMock,
	}, ctrl.Finish
}
