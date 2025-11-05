package file_filter

// FindFilesByPatterns находит все файлы, соответствующие паттернам в директории
func (s *FileFilterService) FindFilesByPatterns(dir string, patterns []string) ([]string, error) {
	allFiles, err := s.fileOps.FindAllFiles(dir)
	if err != nil {
		return nil, err
	}

	return s.filterFilesByPatterns(allFiles, dir, patterns), nil
}
