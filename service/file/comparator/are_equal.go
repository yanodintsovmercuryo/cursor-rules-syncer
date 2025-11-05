package comparator

import (
	"path/filepath"
	"strings"
)

const mdcExtension = ".mdc"

// isMdcFile checks if file is .mdc file (case-insensitive)
func isMdcFile(filePath string) bool {
	return strings.EqualFold(filepath.Ext(filePath), mdcExtension)
}

// AreEqual compares files using header-aware comparison only for .mdc files
func (c *Comparator) AreEqual(file1, file2 string, overwriteHeaders bool) (bool, error) {
	if isMdcFile(file1) && isMdcFile(file2) {
		if overwriteHeaders {
			return c.areEqualNormalized(file1, file2)
		}

		return c.areEqualNormalizedWithoutHeaders(file1, file2)
	}

	return c.areEqualNormalized(file1, file2)
}

// areEqualNormalized compares two files after normalizing their content
func (c *Comparator) areEqualNormalized(file1, file2 string) (bool, error) {
	content1, err := c.fileOps.ReadFileNormalized(file1)
	if err != nil {
		return false, err
	}

	content2, err := c.fileOps.ReadFileNormalized(file2)
	if err != nil {
		return false, err
	}

	return content1 == content2, nil
}

// areEqualNormalizedWithoutHeaders compares two files ignoring YAML headers
func (c *Comparator) areEqualNormalizedWithoutHeaders(file1, file2 string) (bool, error) {
	content1, err := c.fileOps.ReadFileNormalized(file1)
	if err != nil {
		return false, err
	}

	content2, err := c.fileOps.ReadFileNormalized(file2)
	if err != nil {
		return false, err
	}

	contentWithoutHeader1 := c.headerService.RemoveHeaderFromContent(content1)
	contentWithoutHeader2 := c.headerService.RemoveHeaderFromContent(content2)

	return contentWithoutHeader1 == contentWithoutHeader2, nil
}
