//go:generate mockgen -source=service.go -destination=mocks/mocks.go -package=mocks
package copier

import (
	"os"

	"github.com/yanodintsovmercuryo/cursync/pkg/header"
)

type fileOps interface {
	ReadFileNormalized(filePath string) (string, error)
	WriteFile(filePath, content string, perm os.FileMode) error
	FileExists(filePath string) (bool, error)
	CopyFile(srcPath, dstPath string) error
}

type headerService interface {
	RemoveHeaderFromContent(content string) string
	ExtractHeaderFromContent(content string) string
}

// Copier handles file copying with header preservation support
type Copier struct {
	fileOps       fileOps
	headerService headerService
}

// NewCopier creates a new Copier instance
func NewCopier(fileOps fileOps) *Copier {
	headerService := header.NewHeader()

	return &Copier{
		fileOps:       fileOps,
		headerService: headerService,
	}
}

