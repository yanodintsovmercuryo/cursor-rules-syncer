package output_test

import (
	"bytes"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yanodintsovmercuryo/cursync/pkg/output"
)

func TestOutput_PrintInfo(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	o := output.NewOutputWithWriters(&buf, &buf)

	message := "test info"
	o.PrintInfo(message)

	expected := message + "\n"
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}

func TestOutput_PrintError(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	o := output.NewOutputWithWriters(&buf, &buf)

	message := "test error"
	o.PrintError(message)

	expected := message + "\n"
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}

func TestOutput_PrintErrorf(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	o := output.NewOutputWithWriters(&buf, &buf)

	format := "error: %s"
	arg := "test"
	o.PrintErrorf(format, arg)

	expected := "error: test\n"
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}

func TestOutput_PrintOperation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		operationType string
		relativePath  string
		expected      string
	}{
		{
			name:          "add operation",
			operationType: "add",
			relativePath:  "file.txt",
			expected:      "\033[32m+ file.txt\033[0m\n",
		},
		{
			name:          "delete operation",
			operationType: "delete",
			relativePath:  "file.txt",
			expected:      "\033[31m- file.txt\033[0m\n",
		},
		{
			name:          "update operation",
			operationType: "update",
			relativePath:  "file.txt",
			expected:      "\033[33m* file.txt\033[0m\n",
		},
		{
			name:          "unknown operation",
			operationType: "unknown",
			relativePath:  "file.txt",
			expected:      "\033[0m? file.txt\033[0m\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			o := output.NewOutputWithWriters(&buf, &buf)

			o.PrintOperation(tt.operationType, tt.relativePath)

			if diff := cmp.Diff(tt.expected, buf.String()); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOutput_PrintOperationWithTarget(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		operationType string
		relativePath  string
		target        string
		expected      string
	}{
		{
			name:          "add operation with target",
			operationType: "add",
			relativePath:  "file.txt",
			target:        "target",
			expected:      "\033[32m+ file.txt (to target)\033[0m\n",
		},
		{
			name:          "delete operation with target",
			operationType: "delete",
			relativePath:  "file.txt",
			target:        "target",
			expected:      "\033[31m- file.txt (to target)\033[0m\n",
		},
		{
			name:          "update operation with target",
			operationType: "update",
			relativePath:  "file.txt",
			target:        "target",
			expected:      "\033[33m* file.txt (to target)\033[0m\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			o := output.NewOutputWithWriters(&buf, &buf)

			o.PrintOperationWithTarget(tt.operationType, tt.relativePath, tt.target)

			if diff := cmp.Diff(tt.expected, buf.String()); diff != "" {
				t.Fatalf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestOutput_PrintSuccess(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	o := output.NewOutputWithWriters(&buf, &buf)

	message := "success"
	o.PrintSuccess(message)

	expected := "\033[32m" + message + "\033[0m\n"
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}

func TestOutput_PrintWarning(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	o := output.NewOutputWithWriters(&buf, &buf)

	message := "warning"
	o.PrintWarning(message)

	expected := "\033[33m" + message + "\033[0m\n"
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}

func TestOutput_PrintWarningf(t *testing.T) {
	t.Parallel()

	var buf bytes.Buffer
	o := output.NewOutputWithWriters(&buf, &buf)

	format := "warning: %s"
	arg := "test"
	o.PrintWarningf(format, arg)

	expected := "\033[33mwarning: test\033[0m\n"
	if diff := cmp.Diff(expected, buf.String()); diff != "" {
		t.Fatalf("mismatch (-want +got):\n%s", diff)
	}
}
