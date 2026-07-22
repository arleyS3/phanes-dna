package dna

import (
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DetectFramework inspects configuration files in the root path to identify the framework.
func DetectFramework(rootPath string) string {
	// 1. Check Node.js project (package.json)
	packageJsonPath := filepath.Join(rootPath, "package.json")
	if data, err := os.ReadFile(packageJsonPath); err == nil {
		content := string(data)
		if strings.Contains(content, "@nestjs/core") {
			return "NestJS"
		}
		if strings.Contains(content, "next") {
			return "Next.js"
		}
		if strings.Contains(content, "react") {
			return "React"
		}
		if strings.Contains(content, "@angular/core") {
			return "Angular"
		}
		return "Node.js (Generic)"
	}

	// 2. Check Java Project (pom.xml / build.gradle)
	pomPath := filepath.Join(rootPath, "pom.xml")
	gradlePath := filepath.Join(rootPath, "build.gradle")
	if _, err := os.Stat(pomPath); err == nil {
		if data, errReadFile := os.ReadFile(pomPath); errReadFile == nil && strings.Contains(string(data), "spring-boot") {
			return "Spring Boot"
		}
		return "Java Maven"
	}
	if _, err := os.Stat(gradlePath); err == nil {
		if data, errReadFile := os.ReadFile(gradlePath); errReadFile == nil && strings.Contains(string(data), "spring-boot") {
			return "Spring Boot"
		}
		return "Java Gradle"
	}

	// 3. Check Go Project (go.mod)
	goModPath := filepath.Join(rootPath, "go.mod")
	if _, err := os.Stat(goModPath); err == nil {
		return "Go"
	}

	return "Generic Stack"
}

// DetectLanguage checks the system environment to determine the preferred language.
// It returns "es" for Spanish locales and "en" for English/other locales.
func DetectLanguage() string {
	// 1. Check environment variable PHANES_LANGUAGE
	if envLang := os.Getenv("PHANES_LANGUAGE"); envLang != "" {
		if strings.HasPrefix(strings.ToLower(envLang), "es") {
			return "es"
		}
		return "en"
	}

	// 2. Check system LANG environment variable (standard in POSIX systems)
	lang := os.Getenv("LANG")
	if lang == "" {
		lang = os.Getenv("LC_ALL")
	}
	if lang == "" {
		lang = os.Getenv("LC_MESSAGES")
	}

	if strings.HasPrefix(strings.ToLower(lang), "es") {
		return "es"
	}

	return "en"
}

// Project represents a project root captured at analysis time.
type Project struct {
	ID         int64
	Name       string
	RootPath   string
	DetectedAt time.Time
}

// SourceFile represents a single source file with its AST and layer tags.
type SourceFile struct {
	Path     string
	Language string
	AST      []ASTNode
	Layers   []string
}

// ASTNode is a tree-sitter derived node: class, method, field, annotation, etc.
type ASTNode struct {
	Type     string
	Name     string
	Line     int
	Col      int
	Children []ASTNode
}

// Dependency is a directed edge between two code elements.
// Type describes the relationship, e.g. "Controller→Repository".
type Dependency struct {
	Source string
	Target string
	Type   string
}

// LayerRule defines allowed and denied dependencies for a given layer.
type LayerRule struct {
	Layer       string
	TargetLayer string
	Allowed     bool
	Severity    string
	AllowedDeps []string
	DeniedDeps  []string
}

// Chunk is a contiguous source fragment with its embedding vector.
type Chunk struct {
	ID        string
	FilePath  string
	StartLine int
	EndLine   int
	Content   string
	Embedding []float32
}

// Violation represents a single rule breach found during review.
type Violation struct {
	Rule     string
	Severity string
	Location string
	Message  string
}
