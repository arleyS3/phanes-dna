package analyzer

import (
	"path/filepath"
	"strings"

	"github.com/arleyS3/phanes-dna/internal/dna"
)

// Analyzer parses source files for a specific language stack.
type Analyzer interface {
	Language() string
	CanAnalyze(ext string) bool
	AnalyzeFile(path string) (*dna.SourceFile, error)
}

// Registry holds all registered language analyzers.
type Registry struct {
	analyzers []Analyzer
}

func NewRegistry() *Registry {
	return &Registry{
		analyzers: []Analyzer{
			NewJavaAnalyzer(),
			NewGoAnalyzer(),
			NewGenericAnalyzer(),
		},
	}
}

func (r *Registry) GetAnalyzer(path string) Analyzer {
	ext := strings.ToLower(filepath.Ext(path))
	for _, a := range r.analyzers {
		if a.CanAnalyze(ext) {
			return a
		}
	}
	return nil
}
