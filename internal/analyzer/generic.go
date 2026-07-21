package analyzer

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arleyS3/phanes-dna/internal/dna"
)

type GenericAnalyzer struct{}

func NewGenericAnalyzer() *GenericAnalyzer {
	return &GenericAnalyzer{}
}

func (a *GenericAnalyzer) Language() string {
	return "generic"
}

func (a *GenericAnalyzer) CanAnalyze(ext string) bool {
	switch strings.ToLower(ext) {
	case ".py", ".ts", ".js", ".jsx", ".tsx":
		return true
	default:
		return false
	}
}

var (
	pyClassRe = regexp.MustCompile(`^\s*class\s+([A-Za-z0-9_]+)`)
	pyDefRe   = regexp.MustCompile(`^\s*def\s+([A-Za-z0-9_]+)`)
	tsClassRe = regexp.MustCompile(`^\s*(?:export\s+)?class\s+([A-Za-z0-9_]+)`)
	tsFuncRe  = regexp.MustCompile(`^\s*(?:export\s+)?(?:function|const)\s+([A-Za-z0-9_]+)`)
)

func (a *GenericAnalyzer) AnalyzeFile(path string) (*dna.SourceFile, error) {
	ext := strings.ToLower(filepath.Ext(path))
	if !a.CanAnalyze(ext) {
		return nil, fmt.Errorf("unsupported file extension %q for generic analyzer", ext)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open generic file: %w", err)
	}
	defer file.Close()

	lang := "python"
	if ext == ".ts" || ext == ".tsx" {
		lang = "typescript"
	} else if ext == ".js" || ext == ".jsx" {
		lang = "javascript"
	}

	sf := &dna.SourceFile{
		Path:     path,
		Language: lang,
		AST:      []dna.ASTNode{},
		Layers:   []string{},
	}

	// Classify layer from path
	lowerPath := strings.ToLower(path)
	if strings.Contains(lowerPath, "controller") || strings.Contains(lowerPath, "routes") || strings.Contains(lowerPath, "api") {
		sf.Layers = append(sf.Layers, "controller")
	} else if strings.Contains(lowerPath, "service") || strings.Contains(lowerPath, "usecase") {
		sf.Layers = append(sf.Layers, "service")
	} else if strings.Contains(lowerPath, "repository") || strings.Contains(lowerPath, "models") || strings.Contains(lowerPath, "db") {
		sf.Layers = append(sf.Layers, "repository")
	} else {
		sf.Layers = append(sf.Layers, "utility")
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0
	var currentDoc []string
	inDocBlock := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Capture JSDoc / Docstring blocks
		if strings.HasPrefix(trimmed, "/**") || strings.HasPrefix(trimmed, `"""`) {
			inDocBlock = true
			currentDoc = append(currentDoc, trimmed)
			if strings.HasSuffix(trimmed, "*/") && len(trimmed) > 3 {
				inDocBlock = false
			}
			continue
		}
		if inDocBlock {
			currentDoc = append(currentDoc, trimmed)
			if strings.HasSuffix(trimmed, "*/") || (strings.HasSuffix(trimmed, `"""`) && len(trimmed) > 3) {
				inDocBlock = false
			}
			continue
		}

		docComment := strings.Join(currentDoc, "\n")

		if m := pyClassRe.FindStringSubmatch(line); len(m) > 1 {
			node := dna.ASTNode{Type: "class", Name: m[1], Line: lineNum}
			if docComment != "" {
				node.Children = append(node.Children, dna.ASTNode{Type: "docstring", Name: docComment})
			}
			sf.AST = append(sf.AST, node)
			currentDoc = nil
		} else if m := tsClassRe.FindStringSubmatch(line); len(m) > 1 {
			node := dna.ASTNode{Type: "class", Name: m[1], Line: lineNum}
			if docComment != "" {
				node.Children = append(node.Children, dna.ASTNode{Type: "jsdoc", Name: docComment})
			}
			sf.AST = append(sf.AST, node)
			currentDoc = nil
		} else if m := pyDefRe.FindStringSubmatch(line); len(m) > 1 {
			node := dna.ASTNode{Type: "function", Name: m[1], Line: lineNum}
			if docComment != "" {
				node.Children = append(node.Children, dna.ASTNode{Type: "docstring", Name: docComment})
			}
			sf.AST = append(sf.AST, node)
			currentDoc = nil
		} else if m := tsFuncRe.FindStringSubmatch(line); len(m) > 1 {
			node := dna.ASTNode{Type: "function", Name: m[1], Line: lineNum}
			if docComment != "" {
				node.Children = append(node.Children, dna.ASTNode{Type: "jsdoc", Name: docComment})
			}
			sf.AST = append(sf.AST, node)
			currentDoc = nil
		} else if trimmed != "" && !strings.HasPrefix(trimmed, "//") && !strings.HasPrefix(trimmed, "#") {
			// Reset doc if non-matching code line
			currentDoc = nil
		}
	}

	return sf, scanner.Err()
}
