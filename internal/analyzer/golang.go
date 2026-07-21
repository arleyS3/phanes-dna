package analyzer

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/arley/phanes-dna/internal/dna"
)

type GoAnalyzer struct{}

func NewGoAnalyzer() *GoAnalyzer {
	return &GoAnalyzer{}
}

func (a *GoAnalyzer) Language() string {
	return "go"
}

func (a *GoAnalyzer) CanAnalyze(ext string) bool {
	return ext == ".go"
}

func (a *GoAnalyzer) AnalyzeFile(path string) (*dna.SourceFile, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".go" {
		return nil, fmt.Errorf("unsupported file extension %q for go analyzer", ext)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read go file: %w", err)
	}

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse go file: %w", err)
	}

	sf := &dna.SourceFile{
		Path:     path,
		Language: "go",
		AST:      []dna.ASTNode{},
		Layers:   []string{},
	}

	// Classify layer from package name or file path
	pkgName := node.Name.Name
	lowerPath := strings.ToLower(path)
	if strings.Contains(pkgName, "controller") || strings.Contains(lowerPath, "controller") || strings.Contains(lowerPath, "handler") {
		sf.Layers = append(sf.Layers, "controller")
	} else if strings.Contains(pkgName, "service") || strings.Contains(lowerPath, "service") || strings.Contains(lowerPath, "usecase") {
		sf.Layers = append(sf.Layers, "service")
	} else if strings.Contains(pkgName, "store") || strings.Contains(pkgName, "repo") || strings.Contains(lowerPath, "store") {
		sf.Layers = append(sf.Layers, "repository")
	} else {
		sf.Layers = append(sf.Layers, "domain")
	}

	// Extract AST nodes (structs, interfaces, functions, methods)
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		switch x := n.(type) {
		case *ast.TypeSpec:
			sf.AST = append(sf.AST, dna.ASTNode{
				Type: "struct_or_interface",
				Name: x.Name.Name,
				Line: fset.Position(x.Pos()).Line,
			})
		case *ast.FuncDecl:
			name := x.Name.Name
			if x.Recv != nil && len(x.Recv.List) > 0 {
				name = fmt.Sprintf("Method:%s", name)
			}
			sf.AST = append(sf.AST, dna.ASTNode{
				Type: "function",
				Name: name,
				Line: fset.Position(x.Pos()).Line,
			})
		}
		return true
	})

	return sf, nil
}
