package pattern_test

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPatternFilterService_FilterFilesByPatterns(t *testing.T) {
	t.Run("empty patterns returns all files", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		files := []string{"/test/file1.txt", "/test/file2.txt"}
		patterns := []string{}

		result := f.patternFilter.FilterFilesByPatterns(files, "/test", patterns)

		if diff := cmp.Diff(files, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("filters files by pattern", func(t *testing.T) {
		t.Parallel()
		patternFilter, finish := setUpWithRealPathUtils(t)
		defer finish()

		files := []string{"/test/file1.txt", "/test/file2.md", "/test/file3.txt"}
		patterns := []string{"*.txt"}
		expected := []string{"/test/file1.txt", "/test/file3.txt"}

		result := patternFilter.FilterFilesByPatterns(files, "/test", patterns)

		if diff := cmp.Diff(expected, result); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("uses base name when relative path fails", func(t *testing.T) {
		t.Parallel()
		f, finish := setUp(t)
		defer finish()

		files := []string{"/test/file.txt"}
		patterns := []string{"file.txt"}

		f.pathUtilsMock.EXPECT().
			GetRelativePath("/test/file.txt", "/test").
			Return("", errors.New("relative path error")).
			Times(1)

		result := f.patternFilter.FilterFilesByPatterns(files, "/test", patterns)

		if diff := cmp.Diff(len(files), len(result)); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})

	t.Run("empty baseDir uses base name", func(t *testing.T) {
		t.Parallel()
		patternFilter, finish := setUpWithRealPathUtils(t)
		defer finish()

		files := []string{"/test/file.txt"}
		patterns := []string{"file.txt"}

		result := patternFilter.FilterFilesByPatterns(files, "", patterns)

		if diff := cmp.Diff(len(files), len(result)); diff != "" {
			t.Fatalf("mismatch (-want +got):\n%s", diff)
		}
	})
}

