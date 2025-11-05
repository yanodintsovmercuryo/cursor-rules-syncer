package path_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"github.com/yanodintsovmercuryo/cursync/pkg/path"
)

func TestPathUtils_GetRelativePath(t *testing.T) {
	t.Parallel()

	p := path.NewPathUtils()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		baseDir := "/base"
		filePath := "/base/sub/file.txt"
		expected := "sub/file.txt"

		result, err := p.GetRelativePath(filePath, baseDir)
		require.NoError(t, err)

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("error on invalid path", func(t *testing.T) {
		t.Parallel()

		baseDir := "/base"
		filePath := "/other/file.txt"

		_, err := p.GetRelativePath(filePath, baseDir)
		require.Error(t, err)
	})
}

func TestPathUtils_NormalizePath(t *testing.T) {
	t.Parallel()

	p := path.NewPathUtils()

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "windows path",
			filePath: "dir\\file.txt",
			expected: "dir/file.txt",
		},
		{
			name:     "unix path",
			filePath: "dir/file.txt",
			expected: "dir/file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := p.NormalizePath(tt.filePath)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPathUtils_GetDirectory(t *testing.T) {
	t.Parallel()

	p := path.NewPathUtils()

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "file in subdirectory",
			filePath: "dir/sub/file.txt",
			expected: "dir/sub",
		},
		{
			name:     "file in root",
			filePath: "file.txt",
			expected: ".",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := p.GetDirectory(tt.filePath)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPathUtils_GetBaseName(t *testing.T) {
	t.Parallel()

	p := path.NewPathUtils()

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "file in subdirectory",
			filePath: "dir/sub/file.txt",
			expected: "file.txt",
		},
		{
			name:     "file in root",
			filePath: "file.txt",
			expected: "file.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := p.GetBaseName(tt.filePath)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestPathUtils_RecreateDirectoryStructure(t *testing.T) {
	t.Parallel()

	p := path.NewPathUtils()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		srcBase := filepath.Join(tmpDir, "src")
		dstBase := filepath.Join(tmpDir, "dst")
		srcPath := filepath.Join(srcBase, "sub", "file.txt")

		require.NoError(t, os.MkdirAll(filepath.Dir(srcPath), os.ModePerm))
		require.NoError(t, os.WriteFile(srcPath, []byte("content"), 0644))

		result, err := p.RecreateDirectoryStructure(srcPath, srcBase, dstBase)
		require.NoError(t, err)

		expected := filepath.Join(dstBase, "sub", "file.txt")
		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}

		require.DirExists(t, filepath.Dir(result))
	})

	t.Run("error on invalid relative path", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		srcBase := filepath.Join(tmpDir, "src")
		dstBase := filepath.Join(tmpDir, "dst")
		srcPath := filepath.Join(tmpDir, "other", "file.txt")

		_, err := p.RecreateDirectoryStructure(srcPath, srcBase, dstBase)
		require.Error(t, err)
	})
}

