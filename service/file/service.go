//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
package file

import (
	"os"

	"github.com/yanodintsovmercuryo/cursync/pkg/output"
	"github.com/yanodintsovmercuryo/cursync/pkg/path"
	"github.com/yanodintsovmercuryo/cursync/service/file/comparator"
	"github.com/yanodintsovmercuryo/cursync/service/file/copier"
	"github.com/yanodintsovmercuryo/cursync/service/file/filter"
)

type fileOps interface {
	FindAllFiles(dir string) ([]string, error)
	ReadFileNormalized(filePath string) (string, error)
	WriteFile(filePath, content string, perm os.FileMode) error
	FileExists(filePath string) (bool, error)
	CopyFile(srcPath, dstPath string) error
	RemoveFile(filePath string) error
	MkdirAll(path string, perm os.FileMode) error
	GetCurrentDir() (string, error)
	Stat(filePath string) (os.FileInfo, error)
}

type outputService interface {
	PrintErrorf(format string, args ...interface{})
	PrintOperation(operationType, relativePath string)
}

type pathUtils interface {
	RecreateDirectoryStructure(srcPath, srcBase, dstBase string) (string, error)
	GetRelativePath(filePath, baseDir string) (string, error)
	NormalizePath(filePath string) string
	GetDirectory(filePath string) string
	GetBaseName(filePath string) string
}

type comparatorService interface {
	AreEqual(file1, file2 string, overwriteHeaders bool) (bool, error)
}

type copierService interface {
	Copy(srcPath, dstPath string, overwriteHeaders bool) error
}

type filterService interface {
	GetFilePatterns(flagValue, envVarName string) ([]string, error)
	FindFilesByPatterns(dir string, patterns []string) ([]string, error)
	CleanupExtraFilesByPatterns(srcFiles []string, srcBase, dstBase string, patterns []string) error
}

// FileService is a facade for file operations
type FileService struct {
	comparator comparatorService
	copier     copierService
	filter     filterService
}

// NewFileService creates a new FileService
func NewFileService(output *output.Output, fileOps fileOps, pathUtils *path.PathUtils) *FileService {
	comparatorImpl := comparator.NewComparator(fileOps)
	copierImpl := copier.NewCopier(fileOps)
	filterImpl := filter.NewFilter(output, fileOps, pathUtils)

	return &FileService{
		comparator: comparatorImpl,
		copier:     copierImpl,
		filter:     filterImpl,
	}
}

// NewFileServiceWithMocks creates a new FileService with provided mocks for testing
func NewFileServiceWithMocks(comparator comparatorService, copier copierService, filter filterService) *FileService {
	return &FileService{
		comparator: comparator,
		copier:     copier,
		filter:     filter,
	}
}

// AreEqual compares files using header-aware comparison only for .mdc files
func (f *FileService) AreEqual(file1, file2 string, overwriteHeaders bool) (bool, error) {
	return f.comparator.AreEqual(file1, file2, overwriteHeaders)
}

// Copy copies file applying header preservation only for .mdc files
func (f *FileService) Copy(srcPath, dstPath string, overwriteHeaders bool) error {
	return f.copier.Copy(srcPath, dstPath, overwriteHeaders)
}

// GetFilePatterns returns file patterns from various sources
func (f *FileService) GetFilePatterns(flagValue, envVarName string) ([]string, error) {
	return f.filter.GetFilePatterns(flagValue, envVarName)
}

// FindFilesByPatterns finds all files matching patterns in directory
func (f *FileService) FindFilesByPatterns(dir string, patterns []string) ([]string, error) {
	return f.filter.FindFilesByPatterns(dir, patterns)
}

// CleanupExtraFilesByPatterns removes files that exist in destination but not in source, considering patterns
func (f *FileService) CleanupExtraFilesByPatterns(srcFiles []string, srcBase, dstBase string, patterns []string) error {
	return f.filter.CleanupExtraFilesByPatterns(srcFiles, srcBase, dstBase, patterns)
}
