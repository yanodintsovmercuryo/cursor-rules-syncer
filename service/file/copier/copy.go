package copier

import (
	"fmt"
	"path/filepath"
)

const mdcExtension = ".mdc"

// Copy copies file applying header preservation only for .mdc files
func (c *Copier) Copy(srcPath, dstPath string, overwriteHeaders bool) error {
	if filepath.Ext(srcPath) == mdcExtension {
		return c.copyFile(srcPath, dstPath, !overwriteHeaders)
	}

	return c.fileOps.CopyFile(srcPath, dstPath)
}

// copyFile copies file optionally preserving headers
func (c *Copier) copyFile(srcPath, dstPath string, preserveHeaders bool) error {
	srcContent, err := c.fileOps.ReadFileNormalized(srcPath)
	if err != nil {
		return fmt.Errorf("failed to read source file %s: %w", srcPath, err)
	}

	var finalContent string

	if preserveHeaders {
		existingHeader, extractErr := c.extractExistingHeader(dstPath)
		if extractErr != nil {
			return extractErr
		}

		if existingHeader != "" {
			srcContentWithoutHeader := c.headerService.RemoveHeaderFromContent(srcContent)
			finalContent = existingHeader + srcContentWithoutHeader
		} else {
			finalContent = srcContent
		}
	} else {
		finalContent = srcContent
	}

	err = c.fileOps.WriteFile(dstPath, finalContent, 0644)
	if err != nil {
		return fmt.Errorf("failed to write destination file %s: %w", dstPath, err)
	}

	return nil
}

// extractExistingHeader extracts YAML header from existing file
func (c *Copier) extractExistingHeader(dstPath string) (string, error) {
	if exists, err := c.fileOps.FileExists(dstPath); err != nil || !exists {
		return "", nil
	}

	content, err := c.fileOps.ReadFileNormalized(dstPath)
	if err != nil {
		return "", nil
	}

	return c.headerService.ExtractHeaderFromContent(content), nil
}
