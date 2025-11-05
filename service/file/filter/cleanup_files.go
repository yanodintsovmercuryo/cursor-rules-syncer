package filter

import (
	"fmt"
)

// CleanupExtraFilesByPatterns removes files that exist in destination but not in source, considering patterns
func (f *Filter) CleanupExtraFilesByPatterns(srcFiles []string, srcBase, dstBase string, patterns []string) error {
	srcFilesMap := make(map[string]bool)
	for _, srcFile := range srcFiles {
		relativePath, err := f.pathUtils.GetRelativePath(srcFile, srcBase)
		if err != nil {
			continue
		}
		srcFilesMap[relativePath] = true
	}

	destFiles, err := f.fileOps.FindAllFiles(dstBase)
	if err != nil {
		return fmt.Errorf("error walking destination directory: %w", err)
	}

	destFiles = f.filterFilesByPatterns(destFiles, dstBase, patterns)

	for _, destFile := range destFiles {
		relativePath, err := f.pathUtils.GetRelativePath(destFile, dstBase)
		if err != nil {
			continue
		}

		if !srcFilesMap[relativePath] {
			if err := f.fileOps.RemoveFile(destFile); err != nil {
				f.output.PrintErrorf("Error deleting file %s: %v\n", relativePath, err)
			} else {
				f.output.PrintOperation("delete", relativePath)
			}
		}
	}

	return nil
}
