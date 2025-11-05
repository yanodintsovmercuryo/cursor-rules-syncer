//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
package comparator

import (
	"github.com/yanodintsovmercuryo/cursync/pkg/header"
)

type fileOps interface {
	ReadFileNormalized(filePath string) (string, error)
}

type headerService interface {
	RemoveHeaderFromContent(content string) string
	ExtractHeaderFromContent(content string) string
}

// Comparator handles file comparison
type Comparator struct {
	fileOps       fileOps
	headerService headerService
}

// NewComparator creates a new Comparator instance
func NewComparator(fileOps fileOps) *Comparator {
	headerService := header.NewHeader()

	return &Comparator{
		fileOps:       fileOps,
		headerService: headerService,
	}
}
