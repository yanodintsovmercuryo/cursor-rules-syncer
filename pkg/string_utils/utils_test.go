package string_utils_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yanodintsovmercuryo/cursync/pkg/string_utils"
)

func TestSplitTrimFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		separator string
		expected  []string
	}{
		{
			name:      "simple split",
			input:     "a,b,c",
			separator: ",",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "split with spaces",
			input:     "a, b, c",
			separator: ",",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "split with empty strings",
			input:     "a,,b, c",
			separator: ",",
			expected:  []string{"a", "b", "c"},
		},
		{
			name:      "empty input",
			input:     "",
			separator: ",",
			expected:  []string{},
		},
		{
			name:      "only spaces",
			input:     " , , ",
			separator: ",",
			expected:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string_utils.SplitTrimFilter(tt.input, tt.separator)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestRemoveDuplicates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		items    []string
		expected []string
	}{
		{
			name:     "no duplicates",
			items:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "with duplicates",
			items:    []string{"a", "b", "a", "c", "b"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "empty slice",
			items:    []string{},
			expected: []string{},
		},
		{
			name:     "all duplicates",
			items:    []string{"a", "a", "a"},
			expected: []string{"a"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := string_utils.RemoveDuplicates(tt.items)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
