//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
package filter

import (
	"github.com/yanodintsovmercuryo/cursync/pkg/output"
	"github.com/yanodintsovmercuryo/cursync/pkg/path"
	"github.com/yanodintsovmercuryo/cursync/service/file/pattern"
)

type fileOps interface {
	FindAllFiles(dir string) ([]string, error)
	RemoveFile(filePath string) error
}

type outputService interface {
	PrintErrorf(format string, args ...interface{})
	PrintOperation(operationType, relativePath string)
}

type pathUtils interface {
	GetRelativePath(filePath, baseDir string) (string, error)
}

type patternFilter interface {
	FilterFilesByPatterns(files []string, baseDir string, patterns []string) []string
}

// Filter handles file filtering
type Filter struct {
	output        outputService
	fileOps       fileOps
	pathUtils     pathUtils
	patternFilter patternFilter
}

// NewFilter creates a new Filter instance
func NewFilter(output *output.Output, fileOps fileOps, pathUtils *path.PathUtils) *Filter {
	patternFilter := pattern.NewPatternFilterService(pathUtils)

	return &Filter{
		output:        output,
		fileOps:       fileOps,
		pathUtils:     pathUtils,
		patternFilter: patternFilter,
	}
}
