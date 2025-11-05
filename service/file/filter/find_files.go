package filter

// FindFilesByPatterns finds all files matching patterns in directory
func (f *Filter) FindFilesByPatterns(dir string, patterns []string) ([]string, error) {
	allFiles, err := f.fileOps.FindAllFiles(dir)
	if err != nil {
		return nil, err
	}

	return f.filterFilesByPatterns(allFiles, dir, patterns), nil
}

// filterFilesByPatterns filters files based on provided patterns
func (f *Filter) filterFilesByPatterns(files []string, baseDir string, patterns []string) []string {
	return f.patternFilter.FilterFilesByPatterns(files, baseDir, patterns)
}

