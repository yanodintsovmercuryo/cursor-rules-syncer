package sync_test

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/yanodintsovmercuryo/cursync/models"
)

func TestSyncService_PullRules(t *testing.T) {
	t.Run("error getting rules source dir", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir: "",
		}

		os.Unsetenv("CURSOR_RULES_DIR")

		result, err := f.syncService.PullRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get rules source dir")
		require.Nil(t, result)
	})

	t.Run("error getting current directory", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir: "/test/rules",
		}
		expectedErr := errors.New("get current dir error")

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return("", expectedErr).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get current directory")
		require.Nil(t, result)
	})

	t.Run("error finding git root", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir: "/test/rules",
		}
		currentDir := "/test/current"
		expectedErr := errors.New("git root error")

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return("", expectedErr).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to find git root")
		require.Nil(t, result)
	})

	t.Run("error getting file patterns", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:     "/test/rules",
			FilePatterns: "",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		expectedErr := errors.New("file patterns error")

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return(nil, expectedErr).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get file patterns")
		require.Nil(t, result)
	})

	t.Run("error creating destination directory", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:     "/test/rules",
			FilePatterns: "",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		destRulesDir := "/test/git/.cursor/rules"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		expectedErr := errors.New("mkdir error")

		f.fileOpsMock.EXPECT().
			MkdirAll(destRulesDir, os.ModePerm).
			Return(expectedErr).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create destination directory")
		require.Nil(t, result)
	})

	t.Run("error finding source files without patterns", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:     "/test/rules",
			FilePatterns: "",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		destRulesDir := "/test/git/.cursor/rules"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll(destRulesDir, os.ModePerm).
			Return(nil).
			Times(1)

		expectedErr := errors.New("find files error")

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return(nil, expectedErr).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to find source files")
		require.Nil(t, result)
	})

	t.Run("error finding source files with patterns", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:     "/test/rules",
			FilePatterns: "*.mdc",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		destRulesDir := "/test/git/.cursor/rules"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("*.mdc", "CURSOR_RULES_PATTERNS").
			Return([]string{"*.mdc"}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll(destRulesDir, os.ModePerm).
			Return(nil).
			Times(1)

		expectedErr := errors.New("find files by patterns error")

		f.fileServiceMock.EXPECT().
			FindFilesByPatterns("/test/rules", []string{"*.mdc"}).
			Return(nil, expectedErr).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to find files by patterns")
		require.Nil(t, result)
	})

	t.Run("error cleaning up extra files", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:     "/test/rules",
			FilePatterns: "*.mdc",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		destRulesDir := "/test/git/.cursor/rules"
		sourceFiles := []string{"/test/rules/file1.mdc"}

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("*.mdc", "CURSOR_RULES_PATTERNS").
			Return([]string{"*.mdc"}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll(destRulesDir, os.ModePerm).
			Return(nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			FindFilesByPatterns("/test/rules", []string{"*.mdc"}).
			Return(sourceFiles, nil).
			Times(1)

		expectedErr := errors.New("cleanup error")

		f.fileServiceMock.EXPECT().
			CleanupExtraFilesByPatterns(sourceFiles, "/test/rules", destRulesDir, []string{"*.mdc"}).
			Return(expectedErr).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to cleanup extra files")
		require.Nil(t, result)
	})

	t.Run("success with file copy", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:         "/test/rules",
			FilePatterns:     "",
			OverwriteHeaders: false,
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		destRulesDir := "/test/git/.cursor/rules"
		sourceFiles := []string{"/test/rules/file1.mdc"}
		srcFile := "/test/rules/file1.mdc"
		dstFile := "/test/git/.cursor/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll(destRulesDir, os.ModePerm).
			Return(nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return(sourceFiles, nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, "/test/rules").
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(destRulesDir).
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, "/test/rules", destRulesDir).
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, "/test/rules").
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			Stat(dstFile).
			Return(nil, os.ErrNotExist).
			Times(1)

		f.fileServiceMock.EXPECT().
			Copy(srcFile, dstFile, false).
			Return(nil).
			Times(1)

		f.outputMock.EXPECT().
			PrintOperation("add", relativePath).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.HasChanges)
		require.Len(t, result.Operations, 1)

		if diff := cmp.Diff(models.OperationAdd, result.Operations[0].Type); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success with file update", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:         "/test/rules",
			FilePatterns:     "",
			OverwriteHeaders: false,
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		destRulesDir := "/test/git/.cursor/rules"
		sourceFiles := []string{"/test/rules/file1.mdc"}
		srcFile := "/test/rules/file1.mdc"
		dstFile := "/test/git/.cursor/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll(destRulesDir, os.ModePerm).
			Return(nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return(sourceFiles, nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, "/test/rules").
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(destRulesDir).
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, "/test/rules", destRulesDir).
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, "/test/rules").
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			Stat(dstFile).
			Return(nil, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			AreEqual(srcFile, dstFile, false).
			Return(false, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			Copy(srcFile, dstFile, false).
			Return(nil).
			Times(1)

		f.outputMock.EXPECT().
			PrintOperation("update", relativePath).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.HasChanges)
		require.Len(t, result.Operations, 1)

		if diff := cmp.Diff(models.OperationUpdate, result.Operations[0].Type); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success skip identical files", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:         "/test/rules",
			FilePatterns:     "",
			OverwriteHeaders: false,
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		destRulesDir := "/test/git/.cursor/rules"
		sourceFiles := []string{"/test/rules/file1.mdc"}
		srcFile := "/test/rules/file1.mdc"
		dstFile := "/test/git/.cursor/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll(destRulesDir, os.ModePerm).
			Return(nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return(sourceFiles, nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, "/test/rules").
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(destRulesDir).
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, "/test/rules", destRulesDir).
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, "/test/rules").
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			Stat(dstFile).
			Return(nil, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			AreEqual(srcFile, dstFile, false).
			Return(true, nil).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.HasChanges)
		require.Empty(t, result.Operations)
	})

	t.Run("error checking destination file", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:         "/test/rules",
			FilePatterns:     "",
			OverwriteHeaders: false,
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		destRulesDir := "/test/git/.cursor/rules"
		sourceFiles := []string{"/test/rules/file1.mdc"}
		srcFile := "/test/rules/file1.mdc"
		dstFile := "/test/git/.cursor/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll(destRulesDir, os.ModePerm).
			Return(nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return(sourceFiles, nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, "/test/rules").
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(destRulesDir).
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, "/test/rules", destRulesDir).
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, "/test/rules").
			Return(relativePath, nil).
			Times(1)

		expectedErr := errors.New("stat error")

		f.fileOpsMock.EXPECT().
			Stat(dstFile).
			Return(nil, expectedErr).
			Times(1)

		f.outputMock.EXPECT().
			PrintErrorf("Error checking destination file %s: %v\n", relativePath, expectedErr).
			Times(1)

		result, err := f.syncService.PullRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.HasChanges)
		require.Empty(t, result.Operations)
	})
}
