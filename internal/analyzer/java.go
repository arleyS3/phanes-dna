package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/java"

	"github.com/arley/phanes-dna/internal/dna"
)

type JavaAnalyzer struct{}

func NewJavaAnalyzer() *JavaAnalyzer {
	return &JavaAnalyzer{}
}

func (a *JavaAnalyzer) Language() string {
	return "java"
}

func (a *JavaAnalyzer) CanAnalyze(ext string) bool {
	return ext == ".java"
}

func (a *JavaAnalyzer) AnalyzeFile(path string) (*dna.SourceFile, error) {
	return AnalyzeFile(path)
}

// AnalyzeFile parses a Java source file using tree-sitter and returns a structured
// SourceFile containing package info, imports, class/interface declarations,
// annotations, methods, and fields.
func AnalyzeFile(path string) (*dna.SourceFile, error) {
	if filepath.Ext(path) != ".java" {
		return nil, fmt.Errorf("unsupported file extension %q (only .java supported)", filepath.Ext(path))
	}
	source, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}

	parser := sitter.NewParser()
	defer parser.Close()
	parser.SetLanguage(java.GetLanguage())

	tree := parser.Parse(nil, source)
	if tree == nil {
		return nil, fmt.Errorf("tree-sitter parse failed for %s", path)
	}
	defer tree.Close()

	root := tree.RootNode()

	file := &dna.SourceFile{
		Path:     path,
		Language: "java",
		AST:      []dna.ASTNode{},
		Layers:   []string{},
	}

	var pkgName string
	walk(root, source, file, &pkgName)

	// Fallback layer classification from package or file name if no class had
	// an explicit layer tag.
	if len(file.Layers) == 0 {
		if layer := classifyLayer(pkgName, path); layer != "" {
			file.Layers = append(file.Layers, layer)
		}
	}

	return file, nil
}

// walk recursively traverses a tree-sitter node and populates file.AST.
func walk(node *sitter.Node, source []byte, file *dna.SourceFile, pkgName *string) {
	switch node.Type() {

	case "package_declaration":
		name := extractContent(node, source)
		*pkgName = name
		file.AST = append(file.AST, dna.ASTNode{
			Type: "package",
			Name: name,
			Line: int(node.StartPoint().Row) + 1,
		})

	case "import_declaration":
		name := extractContent(node, source)
		file.AST = append(file.AST, dna.ASTNode{
			Type: "import",
			Name: name,
			Line: int(node.StartPoint().Row) + 1,
		})

	case "class_declaration", "interface_declaration", "enum_declaration", "record_declaration":
		name := nodeName(node, source)
		anns := extractAnnotations(node, source)
		layer := classifyLayer(name, *pkgName)

		classNode := dna.ASTNode{
			Type: "class",
			Name: name,
			Line: int(node.StartPoint().Row) + 1,
		}
		for _, a := range anns {
			classNode.Children = append(classNode.Children, dna.ASTNode{
				Type: "annotation",
				Name: a,
				Line: int(node.StartPoint().Row) + 1,
			})
		}

		file.AST = append(file.AST, classNode)
		if layer != "" {
			file.Layers = append(file.Layers, layer)
		}

		// Walk the class body for methods, constructors, and fields.
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() == "class_body" || child.Type() == "interface_body" || child.Type() == "enum_body" {
				walkBody(child, source, file)
			}
		}

	case "method_declaration", "constructor_declaration":
		mtype := "method"
		if node.Type() == "constructor_declaration" {
			mtype = "constructor"
		}
		name := nodeName(node, source)
		anns := extractAnnotations(node, source)
		methodNode := dna.ASTNode{
			Type: mtype,
			Name: name,
			Line: int(node.StartPoint().Row) + 1,
		}
		for _, a := range anns {
			methodNode.Children = append(methodNode.Children, dna.ASTNode{
				Type: "annotation",
				Name: a,
				Line: int(node.StartPoint().Row) + 1,
			})
		}
		file.AST = append(file.AST, methodNode)

	case "field_declaration":
		names := fieldNames(node, source)
		anns := extractAnnotations(node, source)
		for _, fname := range names {
			fieldNode := dna.ASTNode{
				Type: "field",
				Name: fname,
				Line: int(node.StartPoint().Row) + 1,
			}
			for _, a := range anns {
				fieldNode.Children = append(fieldNode.Children, dna.ASTNode{
					Type: "annotation",
					Name: a,
					Line: int(node.StartPoint().Row) + 1,
				})
			}
			file.AST = append(file.AST, fieldNode)
		}
	}

	// Recurse into children (method_declaration etc. inside class_body are
	// handled by walkBody, not here, to avoid double-processing).
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		typ := child.Type()
		// Skip class bodies because their children are walked by walkBody.
		if typ == "class_body" || typ == "interface_body" || typ == "enum_body" {
			continue
		}
		if typ != "comment" {
			walk(child, source, file, pkgName)
		}
	}
}

// walkBody walks class/interface/enum body members directly.
func walkBody(node *sitter.Node, source []byte, file *dna.SourceFile) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "method_declaration", "constructor_declaration":
			mtype := "method"
			if child.Type() == "constructor_declaration" {
				mtype = "constructor"
			}
			name := nodeName(child, source)
			anns := extractAnnotations(child, source)
			methodNode := dna.ASTNode{
				Type: mtype,
				Name: name,
				Line: int(child.StartPoint().Row) + 1,
			}
			for _, a := range anns {
				methodNode.Children = append(methodNode.Children, dna.ASTNode{
					Type: "annotation",
					Name: a,
					Line: int(child.StartPoint().Row) + 1,
				})
			}
			file.AST = append(file.AST, methodNode)

		case "field_declaration":
			names := fieldNames(child, source)
			anns := extractAnnotations(child, source)
			for _, fname := range names {
				fieldNode := dna.ASTNode{
					Type: "field",
					Name: fname,
					Line: int(child.StartPoint().Row) + 1,
				}
				for _, a := range anns {
					fieldNode.Children = append(fieldNode.Children, dna.ASTNode{
						Type: "annotation",
						Name: a,
						Line: int(child.StartPoint().Row) + 1,
					})
				}
				file.AST = append(file.AST, fieldNode)
			}
		}
	}
}

// extractContent returns the full text of a node trimmed of whitespace.
func extractContent(node *sitter.Node, source []byte) string {
	return strings.TrimSpace(node.Content(source))
}

// nodeName extracts the name of a named declaration by finding its "name" child.
func nodeName(node *sitter.Node, source []byte) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" {
			return child.Content(source)
		}
	}
	return ""
}

// fieldNames extracts variable declarator names from a field_declaration.
func fieldNames(node *sitter.Node, source []byte) []string {
	var names []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "variable_declarator" {
			for j := 0; j < int(child.ChildCount()); j++ {
				declChild := child.Child(j)
				if declChild.Type() == "identifier" {
					names = append(names, declChild.Content(source))
				}
			}
		}
	}
	return names
}

// extractAnnotations collects annotation names from a node and its modifiers.
func extractAnnotations(node *sitter.Node, source []byte) []string {
	var anns []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		switch child.Type() {
		case "annotation", "marker_annotation":
			if name := annotationName(child, source); name != "" {
				anns = append(anns, name)
			}
		case "modifiers":
			// Annotations are wrapped in a modifiers node in tree-sitter-java.
			for j := 0; j < int(child.ChildCount()); j++ {
				mod := child.Child(j)
				if mod.Type() == "annotation" || mod.Type() == "marker_annotation" {
					if name := annotationName(mod, source); name != "" {
						anns = append(anns, name)
					}
				}
			}
		}
	}
	return anns
}

// annotationName extracts the identifier from an annotation node.
func annotationName(ann *sitter.Node, source []byte) string {
	for i := 0; i < int(ann.ChildCount()); i++ {
		child := ann.Child(i)
		if child.Type() == "identifier" || child.Type() == "scoped_identifier" {
			return child.Content(source)
		}
	}
	return ""
}

// classifyLayer maps a class/package name to an architectural layer.
func classifyLayer(name, pkgPath string) string {
	combined := name + " " + pkgPath
	switch {
	case strings.Contains(combined, "controller") || strings.Contains(combined, "Controller"):
		return "controller"
	case strings.Contains(combined, "service") || strings.Contains(combined, "Service"):
		return "service"
	case strings.Contains(combined, "repository") || strings.Contains(combined, "Repository") || strings.Contains(combined, "repo"):
		return "repository"
	default:
		return "other"
	}
}
