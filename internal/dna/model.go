package dna

import "time"

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
