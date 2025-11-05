package copier_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopier_Copy(t *testing.T) {
	t.Run("non-mdc file", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "file.txt"
		dstPath := "dest.txt"

		f.fileOpsMock.EXPECT().
			CopyFile(srcPath, dstPath).
			Return(nil).
			Times(1)

		err := f.copier.Copy(srcPath, dstPath, false)
		require.NoError(t, err)
	})

	t.Run("mdc file with overwrite headers", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "file.mdc"
		dstPath := "dest.mdc"
		content := "content"

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(srcPath).
			Return(content, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			WriteFile(dstPath, content, os.FileMode(0644)).
			Return(nil).
			Times(1)

		err := f.copier.Copy(srcPath, dstPath, true)
		require.NoError(t, err)
	})

	t.Run("mdc file without overwrite headers with existing file", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "file.mdc"
		dstPath := "dest.mdc"
		srcContent := "---\nheader1\n---\ncontent"
		dstContent := "---\nheader2\n---\nold"
		expectedContent := "---\nheader2\n---\ncontent"

		f.fileOpsMock.EXPECT().
			FileExists(dstPath).
			Return(true, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(dstPath).
			Return(dstContent, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(srcPath).
			Return(srcContent, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			WriteFile(dstPath, expectedContent, os.FileMode(0644)).
			Return(nil).
			Times(1)

		err := f.copier.Copy(srcPath, dstPath, false)
		require.NoError(t, err)
	})

	t.Run("mdc file without overwrite headers without existing file", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "file.mdc"
		dstPath := "dest.mdc"
		content := "content"

		f.fileOpsMock.EXPECT().
			FileExists(dstPath).
			Return(false, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(srcPath).
			Return(content, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			WriteFile(dstPath, content, os.FileMode(0644)).
			Return(nil).
			Times(1)

		err := f.copier.Copy(srcPath, dstPath, false)
		require.NoError(t, err)
	})

	t.Run("error reading source file", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "file.mdc"
		dstPath := "dest.mdc"
		expectedErr := errors.New("read error")

		// FileExists is called only if preserveHeaders == true
		// But in this case it is called in extractExistingHeader after ReadFileNormalized
		// However if ReadFileNormalized returns error, extractExistingHeader is not called
		// Therefore FileExists should not be called in this scenario
		f.fileOpsMock.EXPECT().
			ReadFileNormalized(srcPath).
			Return("", expectedErr).
			Times(1)

		err := f.copier.Copy(srcPath, dstPath, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to read source file")
	})

	t.Run("error writing destination file", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "file.mdc"
		dstPath := "dest.mdc"
		content := "content"
		expectedErr := errors.New("write error")

		f.fileOpsMock.EXPECT().
			FileExists(dstPath).
			Return(false, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			ReadFileNormalized(srcPath).
			Return(content, nil).
			Times(1)

		f.fileOpsMock.EXPECT().
			WriteFile(dstPath, content, os.FileMode(0644)).
			Return(expectedErr).
			Times(1)

		err := f.copier.Copy(srcPath, dstPath, false)
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to write destination file")
	})

	t.Run("error copying non-mdc file", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "file.txt"
		dstPath := "dest.txt"
		expectedErr := errors.New("copy error")

		f.fileOpsMock.EXPECT().
			CopyFile(srcPath, dstPath).
			Return(expectedErr).
			Times(1)

		err := f.copier.Copy(srcPath, dstPath, false)
		require.ErrorIs(t, err, expectedErr)
	})
}

func TestCopier_Copy_MdcExtension(t *testing.T) {
	t.Run("case sensitivity test", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		srcPath := "file.MDC"
		dstPath := "dest.MDC"

		isMdc := filepath.Ext(srcPath) == ".mdc"
		if !isMdc {
			f.fileOpsMock.EXPECT().
				CopyFile(srcPath, dstPath).
				Return(nil).
				Times(1)
		}

		err := f.copier.Copy(srcPath, dstPath, false)
		require.NoError(t, err)
	})
}
