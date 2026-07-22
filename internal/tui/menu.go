package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/arleyS3/phanes-dna/internal/ai"
	"github.com/arleyS3/phanes-dna/internal/dna"
	"github.com/charmbracelet/huh"
)

type MenuOption struct {
	Title       string
	Description string
	ActionKey   string
}

var menuTranslations = map[string]map[string]string{
	"es": {
		"title": "🧬 Phanes DNA — CLI de Arquitecto de Software Residente",
		"select": "Seleccione una acción a ejecutar",
		"analyze": "Analizar Proyecto — Escanea e indexa archivos fuente",
		"review": "Revisar Arquitectura — Audita el cumplimiento de capas",
		"commit": "Smart Git Commit — Genera conventional commit con IA",
		"config-ai": "Configurar Proveedor de IA — Establece API Keys/Ollama URL",
		"setup": "Configurar MCP para Agentes de IA — Instala en editores",
		"setup-rules": "Configurar Reglas del Proyecto — Define PHANES_RULES.md",
		"export": "Exportar Paquete de Sincronización (.dna) — Exporta a bundle",
		"import": "Importar Paquete de Sincronización (.dna) — Importa a la DB",
		"hooks": "Instalar Git Hooks — Instala bloqueadores de arquitectura",
		"doctor": "Ejecutar Doctor — Diagnósticos de salud del ecosistema",
		"exit": "Salir — Cerrar CLI de Phanes DNA",
	},
	"en": {
		"title": "🧬 Phanes DNA — Resident Software Architect CLI",
		"select": "Select an action to execute",
		"analyze": "Analyze Project — Scan & index source files",
		"review": "Review Architecture — Audit layer compliance",
		"commit": "Smart Git Commit — Generate conventional commit",
		"config-ai": "Configure AI Provider — Set API Keys/Ollama URL",
		"setup": "Setup MCP for AI Agents — Install into editors",
		"setup-rules": "Setup Project Rules — Define PHANES_RULES.md",
		"export": "Export Sync Bundle (.dna) — Export rules to bundle",
		"import": "Import Sync Bundle (.dna) — Import rules to store",
		"hooks": "Install Git Hooks — Install architecture gates",
		"doctor": "Run Doctor — Execute health check diagnostics",
		"exit": "Exit — Close Phanes DNA CLI",
	},
}

// RunInteractiveMenu presents an interactive CLI interface with arrow navigation using Huh.
func RunInteractiveMenu() string {
	var choice string

	lang := dna.DetectLanguage()
	t := menuTranslations[lang]

	fmt.Println("\n" + t["title"])
	fmt.Println("==================================================")

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(t["select"]).
				Options(
					huh.NewOption(t["analyze"], "analyze"),
					huh.NewOption(t["review"], "review"),
					huh.NewOption(t["commit"], "commit"),
					huh.NewOption(t["config-ai"], "config-ai"),
					huh.NewOption(t["setup"], "setup"),
					huh.NewOption(t["setup-rules"], "setup-rules"),
					huh.NewOption(t["export"], "export"),
					huh.NewOption(t["import"], "import"),
					huh.NewOption(t["hooks"], "hooks"),
					huh.NewOption(t["doctor"], "doctor"),
					huh.NewOption(t["exit"], "exit"),
				).
				Value(&choice),
		),
	)

	err := form.Run()
	if err != nil {
		return "exit"
	}

	if choice == "config-ai" {
		RunConfigureAIInteractive()
		return "exit"
	}

	return choice
}

func RunConfigureAIInteractive() {
	reader := bufio.NewReader(os.Stdin)
	cfg, _ := ai.LoadUserConfig()

	fmt.Println("\n🤖 AI Provider Interactive Setup")
	fmt.Println("--------------------------------------------------")
	fmt.Printf("1. Ollama URL (current: %s): ", cfg.OllamaURL)
	ollamaInput, _ := reader.ReadString('\n')
	ollamaInput = strings.TrimSpace(ollamaInput)
	if ollamaInput != "" {
		cfg.OllamaURL = ollamaInput
	}

	fmt.Printf("2. Gemini API Key (current: %s): ", maskKey(cfg.GeminiKey))
	geminiInput, _ := reader.ReadString('\n')
	geminiInput = strings.TrimSpace(geminiInput)
	if geminiInput != "" {
		cfg.GeminiKey = geminiInput
	}

	fmt.Printf("3. Anthropic API Key (current: %s): ", maskKey(cfg.AnthropicKey))
	anthropicInput, _ := reader.ReadString('\n')
	anthropicInput = strings.TrimSpace(anthropicInput)
	if anthropicInput != "" {
		cfg.AnthropicKey = anthropicInput
	}

	if err := ai.SaveUserConfig(cfg); err != nil {
		fmt.Printf("❌ Failed to save config: %v\n", err)
	} else {
		fmt.Println("✅ AI Provider configuration saved to ~/.phanes-dna/config.json")
	}
}

func maskKey(k string) string {
	if k == "" {
		return "not set"
	}
	if len(k) < 8 {
		return "***"
	}
	return k[:4] + "..." + k[len(k)-4:]
}
