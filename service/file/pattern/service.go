//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
package pattern

type pathUtils interface {
	GetRelativePath(filePath, baseDir string) (string, error)
}

// PatternFilterService handles file filtering by patterns
type PatternFilterService struct {
	pathUtils pathUtils
}

// NewPatternFilterService creates a new PatternFilterService instance
func NewPatternFilterService(pathUtils pathUtils) *PatternFilterService {
	return &PatternFilterService{
		pathUtils: pathUtils,
	}
}

