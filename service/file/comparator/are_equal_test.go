package comparator_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestComparator_AreEqual(t *testing.T) {
	t.Run("mdc files with overwrite headers", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.mdc"
		file2 := "file2.mdc"
		content := "content"

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file1).
			Return(content, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file2).
			Return(content, nil).
			Times(1)

		result, err := f.comparator.AreEqual(file1, file2, true)
		require.NoError(t, err)

		if diff := cmp.Diff(true, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("mdc files without overwrite headers", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.mdc"
		file2 := "file2.mdc"
		content1 := "---\nheader\n---\ncontent"
		content2 := "---\nother\n---\ncontent"

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file1).
			Return(content1, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file2).
			Return(content2, nil).
			Times(1)

		result, err := f.comparator.AreEqual(file1, file2, false)
		require.NoError(t, err)

		if diff := cmp.Diff(true, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("non-mdc files", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.txt"
		file2 := "file2.txt"
		content := "content"

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file1).
			Return(content, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file2).
			Return(content, nil).
			Times(1)

		result, err := f.comparator.AreEqual(file1, file2, false)
		require.NoError(t, err)

		if diff := cmp.Diff(true, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("different content", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.txt"
		file2 := "file2.txt"
		content1 := "content1"
		content2 := "content2"

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file1).
			Return(content1, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file2).
			Return(content2, nil).
			Times(1)

		result, err := f.comparator.AreEqual(file1, file2, false)
		require.NoError(t, err)

		if diff := cmp.Diff(false, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error reading first file", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.txt"
		file2 := "file2.txt"
		expectedErr := errors.New("read error")

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file1).
			Return("", expectedErr).
			Times(1)

		result, err := f.comparator.AreEqual(file1, file2, false)
		require.ErrorIs(t, err, expectedErr)

		if diff := cmp.Diff(false, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error reading second file", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.txt"
		file2 := "file2.txt"
		content1 := "content1"
		expectedErr := errors.New("read error")

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file1).
			Return(content1, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file2).
			Return("", expectedErr).
			Times(1)

		result, err := f.comparator.AreEqual(file1, file2, false)
		require.ErrorIs(t, err, expectedErr)

		if diff := cmp.Diff(false, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})
}

func TestComparator_AreEqual_MdcExtension(t *testing.T) {
	t.Run("one mdc one non-mdc", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.mdc"
		file2 := "file2.txt"
		content := "content"

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file1).
			Return(content, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file2).
			Return(content, nil).
			Times(1)

		result, err := f.comparator.AreEqual(file1, file2, false)
		require.NoError(t, err)

		if diff := cmp.Diff(true, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("case sensitivity test", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		file1 := "file1.MDC"
		file2 := "file2.mdc"
		content := "content"

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file1).
			Return(content, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(file2).
			Return(content, nil).
			Times(1)

		result, err := f.comparator.AreEqual(file1, file2, false)
		require.NoError(t, err)

		// Check that both files have .mdc extension (case-insensitive)
		ext1 := strings.ToLower(filepath.Ext(file1))
		ext2 := strings.ToLower(filepath.Ext(file2))
		expected := ext1 == ".mdc" && ext2 == ".mdc"
		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})
}
