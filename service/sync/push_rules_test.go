package sync_test

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/yanodintsovmercuryo/cursync/models"
)

func TestSyncService_PushRules(t *testing.T) {
	t.Run("error getting rules source dir", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir: "",
		}

		os.Unsetenv("CURSOR_RULES_DIR")

		result, err := f.syncService.PushRules(options)
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

		result, err := f.syncService.PushRules(options)
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

		result, err := f.syncService.PushRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to find git root")
		require.Nil(t, result)
	})

	t.Run("error project rules directory not found", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir: "/test/rules",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		rulesSourceDirInProject := "/test/git/.cursor/rules"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(false, nil).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "project rules directory")
		require.Contains(t, err.Error(), "not found")
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
		rulesSourceDirInProject := "/test/git/.cursor/rules"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		expectedErr := errors.New("file patterns error")

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return(nil, expectedErr).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to get file patterns")
		require.Nil(t, result)
	})

	t.Run("error finding project files without patterns", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:     "/test/rules",
			FilePatterns: "",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		rulesSourceDirInProject := "/test/git/.cursor/rules"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		expectedErr := errors.New("find files error")

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(nil, expectedErr).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to find files")
		require.Nil(t, result)
	})

	t.Run("error finding project files with patterns", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:     "/test/rules",
			FilePatterns: "*.mdc",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		rulesSourceDirInProject := "/test/git/.cursor/rules"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("*.mdc", "CURSOR_RULES_PATTERNS").
			Return([]string{"*.mdc"}, nil).
			Times(1)

		expectedErr := errors.New("find files by patterns error")

		f.fileServiceMock.EXPECT().
			FindFilesByPatterns(rulesSourceDirInProject, []string{"*.mdc"}).
			Return(nil, expectedErr).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to find files by patterns")
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
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(projectFiles, nil).
			Times(1)

		expectedErr := errors.New("mkdir error")

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(expectedErr).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to create destination directory")
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
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("*.mdc", "CURSOR_RULES_PATTERNS").
			Return([]string{"*.mdc"}, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			FindFilesByPatterns(rulesSourceDirInProject, []string{"*.mdc"}).
			Return(projectFiles, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		expectedErr := errors.New("cleanup error")

		f.fileServiceMock.EXPECT().
			CleanupExtraFilesByPatterns(projectFiles, rulesSourceDirInProject, "/test/rules", []string{"*.mdc"}).
			Return(expectedErr).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to cleanup extra files")
		require.Nil(t, result)
	})

	t.Run("success with file copy and commit", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:         "/test/rules",
			FilePatterns:     "",
			OverwriteHeaders: false,
			GitWithoutPush:   false,
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}
		srcFile := "/test/git/.cursor/rules/file1.mdc"
		dstFile := "/test/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(projectFiles, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, rulesSourceDirInProject, "/test/rules").
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(dstFile).
			Return(false, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			Copy(srcFile, dstFile, false).
			Return(nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			GetBaseName("/test/rules").
			Return("rules").
			Times(1)

		f.outputMock.EXPECT().
			PrintOperationWithTarget("add", relativePath, "rules").
			Times(1)

		f.pathUtilsMock.EXPECT().
			GetBaseName(gitRoot).
			Return("git").
			Times(1)

		f.gitOpsMock.EXPECT().
			CommitChanges("/test/rules", "Sync cursor rules: updated from project git", false).
			Return(nil).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.HasChanges)
		require.Len(t, result.Operations, 1)

		if diff := cmp.Diff(models.OperationAdd, result.Operations[0].Type); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success with file update and commit", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:         "/test/rules",
			FilePatterns:     "",
			OverwriteHeaders: false,
			GitWithoutPush:   true,
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}
		srcFile := "/test/git/.cursor/rules/file1.mdc"
		dstFile := "/test/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(projectFiles, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, rulesSourceDirInProject, "/test/rules").
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(dstFile).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			AreEqual(srcFile, dstFile, false).
			Return(false, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			Copy(srcFile, dstFile, false).
			Return(nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			GetBaseName("/test/rules").
			Return("rules").
			Times(1)

		f.outputMock.EXPECT().
			PrintOperationWithTarget("update", relativePath, "rules").
			Times(1)

		f.pathUtilsMock.EXPECT().
			GetBaseName(gitRoot).
			Return("git").
			Times(1)

		f.gitOpsMock.EXPECT().
			CommitChanges("/test/rules", "Sync cursor rules: updated from project git", true).
			Return(nil).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.HasChanges)
		require.Len(t, result.Operations, 1)

		if diff := cmp.Diff(models.OperationUpdate, result.Operations[0].Type); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("success skip identical files no commit", func(t *testing.T) {
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
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}
		srcFile := "/test/git/.cursor/rules/file1.mdc"
		dstFile := "/test/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(projectFiles, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, rulesSourceDirInProject, "/test/rules").
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(dstFile).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			AreEqual(srcFile, dstFile, false).
			Return(true, nil).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.HasChanges)
		require.Empty(t, result.Operations)
	})

	t.Run("success with no files to process", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:     "/test/rules",
			FilePatterns: "",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		rulesSourceDirInProject := "/test/git/.cursor/rules"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return([]string{}, nil).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.HasChanges)
		require.Empty(t, result.Operations)
	})

	t.Run("error recreating directory structure", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		options := &models.SyncOptions{
			RulesDir:     "/test/rules",
			FilePatterns: "",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}
		srcFile := "/test/git/.cursor/rules/file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(projectFiles, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return([]string{}, nil).
			Times(1)

		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return("file1.mdc", nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, rulesSourceDirInProject, "/test/rules").
			Return("", errors.New("recreate error")).
			Times(1)

		f.outputMock.EXPECT().
			PrintErrorf("Error recreating directory structure for %s: %v\n", srcFile, errors.New("recreate error")).
			Times(1)

		result, err := f.syncService.PushRules(options)
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
			RulesDir:     "/test/rules",
			FilePatterns: "",
		}
		currentDir := "/test/current"
		gitRoot := "/test/git"
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}
		srcFile := "/test/git/.cursor/rules/file1.mdc"
		dstFile := "/test/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(projectFiles, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, rulesSourceDirInProject, "/test/rules").
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(dstFile).
			Return(false, errors.New("file exists error")).
			Times(1)

		f.outputMock.EXPECT().
			PrintErrorf("Error checking destination file %s in %s: %v\n", relativePath, "/test/rules", errors.New("file exists error")).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.HasChanges)
		require.Empty(t, result.Operations)
	})

	t.Run("error copying file", func(t *testing.T) {
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
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}
		srcFile := "/test/git/.cursor/rules/file1.mdc"
		dstFile := "/test/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(projectFiles, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, rulesSourceDirInProject, "/test/rules").
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(dstFile).
			Return(false, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			Copy(srcFile, dstFile, false).
			Return(errors.New("copy error")).
			Times(1)

		f.outputMock.EXPECT().
			PrintErrorf("Error synchronizing file %s to %s: %v\n", relativePath, "/test/rules", errors.New("copy error")).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.False(t, result.HasChanges)
		require.Empty(t, result.Operations)
	})

	t.Run("error comparing files", func(t *testing.T) {
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
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}
		srcFile := "/test/git/.cursor/rules/file1.mdc"
		dstFile := "/test/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(projectFiles, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, rulesSourceDirInProject, "/test/rules").
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(dstFile).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			AreEqual(srcFile, dstFile, false).
			Return(false, errors.New("compare error")).
			Times(1)

		f.outputMock.EXPECT().
			PrintErrorf("Error comparing files %s: %v\n", relativePath, errors.New("compare error")).
			Times(1)

		f.fileServiceMock.EXPECT().
			Copy(srcFile, dstFile, false).
			Return(nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			GetBaseName("/test/rules").
			Return("rules").
			Times(1)

		f.outputMock.EXPECT().
			PrintOperationWithTarget("update", relativePath, "rules").
			Times(1)

		f.pathUtilsMock.EXPECT().
			GetBaseName(gitRoot).
			Return("git").
			Times(1)

		f.gitOpsMock.EXPECT().
			CommitChanges("/test/rules", "Sync cursor rules: updated from project git", false).
			Return(nil).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.HasChanges)
		require.Len(t, result.Operations, 1)
	})

	t.Run("error commit changes", func(t *testing.T) {
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
		rulesSourceDirInProject := "/test/git/.cursor/rules"
		projectFiles := []string{"/test/git/.cursor/rules/file1.mdc"}
		srcFile := "/test/git/.cursor/rules/file1.mdc"
		dstFile := "/test/rules/file1.mdc"
		relativePath := "file1.mdc"

		f.fileOpsMock.EXPECT().
			GetCurrentDir().
			Return(currentDir, nil).
			Times(1)

		f.gitOpsMock.EXPECT().
			GetGitRootDir(currentDir).
			Return(gitRoot, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(rulesSourceDirInProject).
			Return(true, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			GetFilePatterns("", "CURSOR_RULES_PATTERNS").
			Return([]string{}, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles(rulesSourceDirInProject).
			Return(projectFiles, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			MkdirAll("/test/rules", os.ModePerm).
			Return(nil).
			Times(1)

		// cleanupExtraFiles is called and searches for files in destination
		// cleanupExtraFiles calls GetRelativePath for each source file
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FindAllFiles("/test/rules").
			Return([]string{}, nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			RecreateDirectoryStructure(srcFile, rulesSourceDirInProject, "/test/rules").
			Return(dstFile, nil).
			Times(1)

		// GetRelativePath is called again in main loop for display
		f.pathUtilsMock.EXPECT().
			GetRelativePath(srcFile, rulesSourceDirInProject).
			Return(relativePath, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			FileExists(dstFile).
			Return(false, nil).
			Times(1)

		f.fileServiceMock.EXPECT().
			Copy(srcFile, dstFile, false).
			Return(nil).
			Times(1)

		f.pathUtilsMock.EXPECT().
			GetBaseName("/test/rules").
			Return("rules").
			Times(1)

		f.outputMock.EXPECT().
			PrintOperationWithTarget("add", relativePath, "rules").
			Times(1)

		f.pathUtilsMock.EXPECT().
			GetBaseName(gitRoot).
			Return("git").
			Times(1)

		f.gitOpsMock.EXPECT().
			CommitChanges("/test/rules", "Sync cursor rules: updated from project git", false).
			Return(errors.New("commit error")).
			Times(1)

		f.outputMock.EXPECT().
			PrintErrorf("Commit failed for %s: %v\n", "/test/rules", errors.New("commit error")).
			Times(1)

		result, err := f.syncService.PushRules(options)
		require.NoError(t, err)
		require.NotNil(t, result)
		require.True(t, result.HasChanges)
		require.Len(t, result.Operations, 1)
	})
}

