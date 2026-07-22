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

// Pattern representa la regla parseada para una capa específica del scaffolding.
type Pattern struct {
	Layer string
	Rule  string
}

// Config almacena las reglas y convenciones de nomenclatura leídas del archivo de configuración.
type Config struct {
	CaseStyle string
	Patterns  []Pattern
}

// SplitIntoWords divide una cadena de entrada (soporta camelCase, snake_case, kebab-case, y espacios) en palabras individuales en minúscula.
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
			// Si es mayúscula y teníamos caracteres en minúscula antes, o si el siguiente es minúscula (ej: APIKey -> API, Key)
			if current.Len() > 0 {
				// Validar si el anterior era minúscula o el siguiente es minúscula
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

// ToCaseStyle convierte la cadena de entrada al estilo de mayúsculas/minúsculas destino.
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

// ParseRulesFile lee el archivo PHANES_RULES.md y extrae la configuración de convenciones de nomenclatura.
func ParseRulesFile(path string) (Config, error) {
	cfg := Config{
		CaseStyle: "snake_case", // valor por defecto
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

// GenerateScaffolding genera la estructura física de archivos para una funcionalidad específica.
func GenerateScaffolding(featureName string) error {
	rulesPath := "PHANES_RULES.md"
	cfg, err := ParseRulesFile(rulesPath)
	if err != nil {
		// Si falta el archivo de reglas, usar valores por defecto para Go
		cfg = Config{
			CaseStyle: "snake_case",
			Patterns: []Pattern{
				{Layer: "controller", Rule: "[name]_controller.go"},
				{Layer: "service", Rule: "[name]_service.go"},
				{Layer: "repository", Rule: "[name]_repository.go"},
			},
		}
		fmt.Printf("⚠️  No se pudo leer %s. Usando convenciones por defecto para Go.\n", rulesPath)
	}

	formattedName := ToCaseStyle(featureName, cfg.CaseStyle)
	pascalName := ToCaseStyle(featureName, "PascalCase")

	fmt.Printf("🔨 Generando scaffolding para la feature '%s' (formato: '%s')...\n", featureName, formattedName)

	for _, p := range cfg.Patterns {
		fileName := strings.ReplaceAll(p.Rule, "[name]", formattedName)
		ext := filepath.Ext(fileName)

		// Determinar el directorio base para la capa correspondiente
		targetDir := determineTargetDirectory(p.Layer)
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("falló al crear el directorio para la capa %s: %w", p.Layer, err)
		}

		filePath := filepath.Join(targetDir, fileName)

		// Validar si el archivo ya existe para evitar sobreescribir código existente
		if _, err := os.Stat(filePath); err == nil {
			fmt.Printf("  ⚠️  El archivo ya existe, omitiendo: %s\n", filePath)
			continue
		}

		// Generar el contenido base según la extensión de lenguaje detectada
		content := generateTemplate(ext, p.Layer, formattedName, pascalName)

		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			return fmt.Errorf("falló al escribir el archivo %s: %w", filePath, err)
		}
		fmt.Printf("  ✅ Creado: %s\n", filePath)
	}

	return nil
}

// determineTargetDirectory decide la ruta física para la capa basándose en heurísticas de la estructura de carpetas.
func determineTargetDirectory(layer string) string {
	// Heurísticas:
	// 1. Validar si existe "internal/<capa>"
	// 2. Validar si existe "<capa>" en la raíz
	// 3. Si existe "internal", por defecto usar "internal/<capa>"
	// 4. Por defecto usar "<capa>" en la raíz

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

// generateTemplate genera código base (boilerplate) según la extensión del lenguaje de programación y la capa.
func generateTemplate(ext, layer, formattedName, pascalName string) string {
	ext = strings.ToLower(ext)

	switch ext {
	case ".go":
		return fmt.Sprintf(`package %s

// %s%s representa la lógica de la capa %s para %s.
type %s%s struct {
	// TODO: Agregar dependencias e inicializaciones
}

// New%s%s crea una nueva instancia de %s%s.
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
    // TODO: Agregar dependencias e inicializaciones
  }

  // TODO: Implementar acciones de la capa %s
}
`, pascalName, strings.Title(layer), layer)

	case ".java":
		return fmt.Sprintf(`package com.project.%s;

public class %s%s {
    public %s%s() {
        // TODO: Implementar constructor
    }
}
`, layer, pascalName, strings.Title(layer), pascalName, strings.Title(layer))

	case ".py":
		return fmt.Sprintf(`class %s%s:
    """
    Lógica de la capa %s para %s.
    """
    def __init__(self):
        pass
`, pascalName, strings.Title(layer), strings.Title(layer), formattedName)

	default:
		// Fallback genérico
		return fmt.Sprintf("// Código base para la capa %s: %s\n", layer, formattedName)
	}
}
