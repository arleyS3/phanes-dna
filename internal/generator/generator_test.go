package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSplitIntoWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"user_profile", []string{"user", "profile"}},
		{"UserProfile", []string{"user", "profile"}},
		{"userProfile", []string{"user", "profile"}},
		{"user-profile", []string{"user", "profile"}},
		{"user.profile", []string{"user", "profile"}},
		{"APIKey", []string{"api", "key"}},
		{"simple", []string{"simple"}},
	}

	for _, tt := range tests {
		result := SplitIntoWords(tt.input)
		if len(result) != len(tt.expected) {
			t.Fatalf("for %s, expected %v but got %v", tt.input, tt.expected, result)
		}
		for i, v := range result {
			if v != tt.expected[i] {
				t.Errorf("for %s at index %d, expected %s but got %s", tt.input, i, tt.expected[i], v)
			}
		}
	}
}

func TestToCaseStyle(t *testing.T) {
	tests := []struct {
		input    string
		style    string
		expected string
	}{
		{"userProfile", "snake_case", "user_profile"},
		{"user_profile", "kebab-case", "user-profile"},
		{"user-profile", "camelCase", "userProfile"},
		{"user_profile", "PascalCase", "UserProfile"},
	}

	for _, tt := range tests {
		result := ToCaseStyle(tt.input, tt.style)
		if result != tt.expected {
			t.Errorf("ToCaseStyle(%s, %s) = %s; want %s", tt.input, tt.style, result, tt.expected)
		}
	}
}

func TestParseRulesFile(t *testing.T) {
	// Create a temporary PHANES_RULES.md mock content
	tmpRules := `
# Phanes DNA Project Rules & Conventions

## 1. Layer Dependencies
- domain cannot import controller

## 4. Naming & Scaffolding Conventions
- Case Style: ` + "`" + `snake_case` + "`" + `
- Patterns:
  - controller: ` + "`" + `[name]_controller.go` + "`" + `
  - service: ` + "`" + `[name]_service.go` + "`" + `
  - repository: ` + "`" + `[name]_repository.go` + "`" + `
`
	tmpFile := filepath.Join(t.TempDir(), "PHANES_RULES.md")
	if err := os.WriteFile(tmpFile, []byte(tmpRules), 0644); err != nil {
		t.Fatalf("failed to write tmp file: %v", err)
	}

	cfg, err := ParseRulesFile(tmpFile)
	if err != nil {
		t.Fatalf("ParseRulesFile failed: %v", err)
	}

	if cfg.CaseStyle != "snake_case" {
		t.Errorf("expected CaseStyle 'snake_case', got '%s'", cfg.CaseStyle)
	}

	if len(cfg.Patterns) != 3 {
		t.Fatalf("expected 3 patterns, got %d", len(cfg.Patterns))
	}

	expected := []Pattern{
		{"controller", "[name]_controller.go"},
		{"service", "[name]_service.go"},
		{"repository", "[name]_repository.go"},
	}

	for i, p := range cfg.Patterns {
		if p.Layer != expected[i].Layer || p.Rule != expected[i].Rule {
			t.Errorf("at pattern %d, expected %+v but got %+v", i, expected[i], p)
		}
	}
}
