package header

import (
	"strings"
)

const (
	headerSeparator = "---"
	maxHeaderLines  = 20
)

// Header handles operations with YAML headers in markdown files
type Header struct{}

// NewHeader creates a new Header instance
func NewHeader() *Header {
	return &Header{}
}

// RemoveHeaderFromContent removes YAML header from markdown content
func (h *Header) RemoveHeaderFromContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return content
	}

	if lines[0] != headerSeparator {
		return content
	}

	for i := 1; i < len(lines); i++ {
		if lines[i] == headerSeparator {
			if i+1 < len(lines) {
				remainingLines := lines[i+1:]
				remainingLines = removeLeadingEmptyLines(remainingLines)
				return strings.Join(remainingLines, "\n")
			}
			return ""
		}
		if i > maxHeaderLines {
			break
		}
	}

	return content
}

// ExtractHeaderFromContent extracts YAML header from markdown content
func (h *Header) ExtractHeaderFromContent(content string) string {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || lines[0] != headerSeparator {
		return ""
	}

	var headerLines []string
	headerLines = append(headerLines, lines[0])

	for i := 1; i < len(lines) && i <= maxHeaderLines; i++ {
		headerLines = append(headerLines, lines[i])
		if lines[i] == headerSeparator {
			return strings.Join(headerLines, "\n") + "\n"
		}
	}

	return ""
}

// removeLeadingEmptyLines removes leading empty lines from slice
func removeLeadingEmptyLines(lines []string) []string {
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}

	return lines
}

