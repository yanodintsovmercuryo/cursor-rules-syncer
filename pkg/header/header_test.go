package header_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yanodintsovmercuryo/cursync/pkg/header"
)

func TestHeader_RemoveHeaderFromContent(t *testing.T) {
	t.Parallel()

	h := header.NewHeader()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "content without header",
			content:  "some content",
			expected: "some content",
		},
		{
			name:     "content with header",
			content:  "---\nkey: value\n---\ncontent",
			expected: "content",
		},
		{
			name:     "content with header and empty lines",
			content:  "---\nkey: value\n---\n\n\ncontent",
			expected: "content",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "only header",
			content:  "---\nkey: value\n---",
			expected: "",
		},
		{
			name:     "header without closing separator",
			content:  "---\nkey: value\ncontent",
			expected: "---\nkey: value\ncontent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := h.RemoveHeaderFromContent(tt.content)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestHeader_ExtractHeaderFromContent(t *testing.T) {
	t.Parallel()

	h := header.NewHeader()

	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "content without header",
			content:  "some content",
			expected: "",
		},
		{
			name:     "content with header",
			content:  "---\nkey: value\n---\ncontent",
			expected: "---\nkey: value\n---\n",
		},
		{
			name:     "empty content",
			content:  "",
			expected: "",
		},
		{
			name:     "only header",
			content:  "---\nkey: value\n---",
			expected: "---\nkey: value\n---\n",
		},
		{
			name:     "header without closing separator",
			content:  "---\nkey: value\ncontent",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := h.ExtractHeaderFromContent(tt.content)

			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
