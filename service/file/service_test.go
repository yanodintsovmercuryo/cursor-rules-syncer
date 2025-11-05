package file_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestFileService_AreEqual(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.txt"
		file2 := "file2.txt"
		overwriteHeaders := false
		expected := true

		f.comparatorMock.EXPECT().
			AreEqual(file1, file2, overwriteHeaders).
			Return(expected, nil).
			Times(1)

		result, err := f.fileService.AreEqual(file1, file2, overwriteHeaders)
		require.NoError(t, err)

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error from comparator", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.txt"
		file2 := "file2.txt"
		overwriteHeaders := false
		expectedErr := errors.New("comparator error")

		f.comparatorMock.EXPECT().
			AreEqual(file1, file2, overwriteHeaders).
			Return(false, expectedErr).
			Times(1)

		result, err := f.fileService.AreEqual(file1, file2, overwriteHeaders)
		require.ErrorIs(t, err, expectedErr)

		if diff := cmp.Diff(false, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestFileService_Copy(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "src.txt"
		dstPath := "dst.txt"
		overwriteHeaders := false

		f.copierMock.EXPECT().
			Copy(srcPath, dstPath, overwriteHeaders).
			Return(nil).
			Times(1)

		err := f.fileService.Copy(srcPath, dstPath, overwriteHeaders)
		require.NoError(t, err)
	})

	t.Run("error from copier", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "src.txt"
		dstPath := "dst.txt"
		overwriteHeaders := false
		expectedErr := errors.New("copier error")

		f.copierMock.EXPECT().
			Copy(srcPath, dstPath, overwriteHeaders).
			Return(expectedErr).
			Times(1)

		err := f.fileService.Copy(srcPath, dstPath, overwriteHeaders)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestFileService_GetFilePatterns(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		flagValue := "*.txt"
		envVarName := "TEST_VAR"
		expected := []string{"*.txt"}

		f.filterMock.EXPECT().
			GetFilePatterns(flagValue, envVarName).
			Return(expected, nil).
			Times(1)

		result, err := f.fileService.GetFilePatterns(flagValue, envVarName)
		require.NoError(t, err)

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error from filter", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		flagValue := "*.txt"
		envVarName := "TEST_VAR"
		expectedErr := errors.New("filter error")

		f.filterMock.EXPECT().
			GetFilePatterns(flagValue, envVarName).
			Return(nil, expectedErr).
			Times(1)

		result, err := f.fileService.GetFilePatterns(flagValue, envVarName)
		require.ErrorIs(t, err, expectedErr)
		require.Empty(t, result)
	})
}

func TestFileService_FindFilesByPatterns(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		dir := "/test"
		patterns := []string{"*.txt"}
		expected := []string{"/test/file.txt"}

		f.filterMock.EXPECT().
			FindFilesByPatterns(dir, patterns).
			Return(expected, nil).
			Times(1)

		result, err := f.fileService.FindFilesByPatterns(dir, patterns)
		require.NoError(t, err)

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error from filter", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		dir := "/test"
		patterns := []string{"*.txt"}
		expectedErr := errors.New("filter error")

		f.filterMock.EXPECT().
			FindFilesByPatterns(dir, patterns).
			Return(nil, expectedErr).
			Times(1)

		result, err := f.fileService.FindFilesByPatterns(dir, patterns)
		require.ErrorIs(t, err, expectedErr)
		require.Empty(t, result)
	})
}

func TestFileService_CleanupExtraFilesByPatterns(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcFiles := []string{"/src/file.txt"}
		srcBase := "/src"
		dstBase := "/dst"
		patterns := []string{"*.txt"}

		f.filterMock.EXPECT().
			CleanupExtraFilesByPatterns(srcFiles, srcBase, dstBase, patterns).
			Return(nil).
			Times(1)

		err := f.fileService.CleanupExtraFilesByPatterns(srcFiles, srcBase, dstBase, patterns)
		require.NoError(t, err)
	})

	t.Run("error from filter", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcFiles := []string{"/src/file.txt"}
		srcBase := "/src"
		dstBase := "/dst"
		patterns := []string{"*.txt"}
		expectedErr := errors.New("filter error")

		f.filterMock.EXPECT().
			CleanupExtraFilesByPatterns(srcFiles, srcBase, dstBase, patterns).
			Return(expectedErr).
			Times(1)

		err := f.fileService.CleanupExtraFilesByPatterns(srcFiles, srcBase, dstBase, patterns)
		require.ErrorIs(t, err, expectedErr)
	})
}
