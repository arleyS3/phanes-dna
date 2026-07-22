package generator

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// Pattern holds the parsed rule for a specific layer.
type Pattern struct {
	Layer string
	Rule  string
}

// Config holds the parsed naming and scaffolding rules.
type Config struct {
	CaseStyle string
	Patterns  []Pattern
}

// SplitIntoWords splits an input string (camelCase, snake_case, kebab-case, space-separated) into individual words.
func SplitIntoWords(input string) []string {
	var words []string
	var current strings.Builder

	runes := []rune(input)
	for i := 0; i < len(runes); i++ {
		r := runes[i]
		if r == '_' || r == '-' || r == ' ' || r == '.' {
			if current.Len() > 0 {
				words = append(words, strings.ToLower(current.String()))
				current.Reset()
			}
			continue
		}

		if unicode.IsUpper(r) {
			// If it's capital and we had lowercase characters before, or if next is lowercase (e.g. APIKey -> API, Key)
			if current.Len() > 0 {
				// Check if previous was lowercase or next is lowercase
				prevIsLower := unicode.IsLower(runes[i-1])
				nextIsLower := i+1 < len(runes) && unicode.IsLower(runes[i+1])
				if prevIsLower || nextIsLower {
					words = append(words, strings.ToLower(current.String()))
					current.Reset()
				}
			}
		}
		current.WriteRune(r)
	}
	if current.Len() > 0 {
		words = append(words, strings.ToLower(current.String()))
	}
	return words
}

// ToCaseStyle converts the input string to the target case style.
func ToCaseStyle(input string, style string) string {
	words := SplitIntoWords(input)
	if len(words) == 0 {
		return input
	}

	switch strings.ToLower(strings.ReplaceAll(style, " ", "")) {
	case "snake_case":
		return strings.Join(words, "_")
	case "kebab-case":
		return strings.Join(words, "-")
	case "camelcase":
		result := words[0]
		for i := 1; i < len(words); i++ {
			if len(words[i]) > 0 {
				result += strings.ToUpper(words[i][:1]) + words[i][1:]
			}
		}
		return result
	case "pascalcase":
		result := ""
		for _, w := range words {
			if len(w) > 0 {
				result += strings.ToUpper(w[:1]) + w[1:]
			}
		}
		return result
	default:
		return input
	}
}

// ParseRulesFile reads PHANES_RULES.md and parses naming conventions.
func ParseRulesFile(path string) (Config, error) {
	cfg := Config{
		CaseStyle: "snake_case", // default
	}

	file, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inSection := false

	caseRegex := regexp.MustCompile(`(?i)-\s*case\s*style:\s*` + "`" + `([^` + "`" + `]+)` + "`")
	patternRegex := regexp.MustCompile(`(?i)-\s*([a-zA-Z0-9_\-]+):\s*` + "`" + `([^` + "`" + `]+)` + "`")

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "##") {
			if strings.Contains(strings.ToLower(line), "naming") || strings.Contains(strings.ToLower(line), "scaffolding") {
				inSection = true
			} else {
				inSection = false
			}
			continue
		}

		if inSection {
			if match := caseRegex.FindStringSubmatch(line); len(match) > 1 {
				cfg.CaseStyle = strings.TrimSpace(match[1])
			} else if match := patternRegex.FindStringSubmatch(line); len(match) > 1 {
				cfg.Patterns = append(cfg.Patterns, Pattern{
					Layer: strings.TrimSpace(match[1]),
					Rule:  strings.TrimSpace(match[2]),
				})
			}
		}
	}

	return cfg, scanner.Err()
}

// GenerateScaffolding generates the layer files for a given feature.
func GenerateScaffolding(featureName string) error {
	rulesPath := "PHANES_RULES.md"
	cfg, err := ParseRulesFile(rulesPath)
	if err != nil {
		// If rules file is missing or failed to parse, use sensible Go defaults
		cfg = Config{
			CaseStyle: "snake_case",
			Patterns: []Pattern{
				{Layer: "controller", Rule: "[name]_controller.go"},
				{Layer: "service", Rule: "[name]_service.go"},
				{Layer: "repository", Rule: "[name]_repository.go"},
			},
		}
		fmt.Printf("⚠️  Could not read %s. Using default Go naming conventions.\n", rulesPath)
	}

	formattedName := ToCaseStyle(featureName, cfg.CaseStyle)
	pascalName := ToCaseStyle(featureName, "PascalCase")

	fmt.Printf("🔨 Generating scaffolding for feature '%s' (formatted: '%s')...\n", featureName, formattedName)

	for _, p := range cfg.Patterns {
		fileName := strings.ReplaceAll(p.Rule, "[name]", formattedName)
		ext := filepath.Ext(fileName)

		// Determine base directory for this layer
		targetDir := determineTargetDirectory(p.Layer)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("failed to create directory for layer %s: %w", p.Layer, err)
		}

		filePath := filepath.Join(targetDir, fileName)

		// Check if file already exists to avoid overwriting
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("  ⚠️  File already exists, skipping: %s\n", filePath)
			continue
		}

		// Generate template content based on language extension
		content := generateTemplate(ext, p.Layer, formattedName, pascalName)

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file %s: %w", filePath, err)
		}
		fmt.Printf("  ✅ Created: %s\n", filePath)
	}

	return nil
}

// determineTargetDirectory determines where the file should be placed.
func determineTargetDirectory(layer string) string {
	// Heuristics:
	// 1. Check if "internal/<layer>" exists
	// 2. Check if "<layer>" exists in root
	// 3. If "internal" directory exists, default to "internal/<layer>"
	// 4. Default to "<layer>" in the root directory

	paths := []string{
		filepath.Join("internal", layer),
		layer,
	}

	for _, path := range paths {
		if info, err := os.Stat(path); err == nil && info.IsDir() {
			return path
		}
	}

	if info, err := os.Stat("internal"); err == nil && info.IsDir() {
		return filepath.Join("internal", layer)
	}

	return layer
}

// generateTemplate creates boilerplate skeleton code based on file extension and layer type.
func generateTemplate(ext, layer, formattedName, pascalName string) string {
	ext = strings.ToLower(ext)

	switch ext {
	case ".go":
		return fmt.Sprintf(`package %s

// %s%s represents the %s layer logic for %s.
type %s%s struct {
	// TODO: Add dependencies and fields
}

// New%s%s creates a new instance of %s%s.
func New%s%s() *%s%s {
	return &%s%s{}
}
`, layer, pascalName, strings.Title(layer), layer, formattedName,
			pascalName, strings.Title(layer),
			pascalName, strings.Title(layer), pascalName, strings.Title(layer),
			pascalName, strings.Title(layer), pascalName, strings.Title(layer),
			pascalName, strings.Title(layer))

	case ".ts", ".tsx":
		return fmt.Sprintf(`export class %s%s {
  constructor() {
    // TODO: Add dependencies and initialization
  }

  // TODO: Implement %s actions
}
`, pascalName, strings.Title(layer), layer)

	case ".java":
		return fmt.Sprintf(`package com.project.%s;

public class %s%s {
    public %s%s() {
        // TODO: Implement constructor
    }
}
`, layer, pascalName, strings.Title(layer), pascalName, strings.Title(layer))

	case ".py":
		return fmt.Sprintf(`class %s%s:
    """
    %s layer logic for %s.
    """
    def __init__(self):
        pass
`, pascalName, strings.Title(layer), strings.Title(layer), formattedName)

	default:
		// Generic file fallback
		return fmt.Sprintf("// Boilerplate for %s layer: %s\n", layer, formattedName)
	}
}
