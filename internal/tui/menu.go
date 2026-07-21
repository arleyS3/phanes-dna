package tui

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/arleyS3/phanes-dna/internal/ai"
)

type MenuOption struct {
	Title       string
	Description string
	ActionKey   string
}

// RunInteractiveMenu presents an interactive CLI interface similar to Gentle-AI.
func RunInteractiveMenu() string {
	fmt.Println("\n🧬 Phanes DNA — Resident Software Architect CLI")
	fmt.Println("==================================================")
	fmt.Println("Select an action to execute:")
	fmt.Println()

	options := []MenuOption{
		{"Analyze Project", "Scan & index multi-stack source files into SQLite", "analyze"},
		{"Review Architecture", "Audit layer rules & compliance", "review"},
		{"Dev Onboarding Mentor", "Ask how features or conventions are built in this project", "onboard"},
		{"Smart Git Commit", "Generate Conventional Commit from branch & staged diff", "commit"},
		{"Configure AI Provider", "Interactively set Ollama URL, Gemini API Key, or Anthropic Key", "config-ai"},
		{"Setup MCP for AI Agents", "Configure MCP stdio server in Cursor/Claude/Antigravity/OpenCode", "setup"},
		{"Export Sync Bundle (.dna)", "Export project rules to ultralight .dna bundle", "export"},
		{"Import Sync Bundle (.dna)", "Import .dna sync bundle into local store", "import"},
		{"Install Git Hooks", "Install pre-commit & pre-push architecture gates", "hooks"},
		{"Run Doctor", "Execute ecosystem health & environment diagnostics", "doctor"},
		{"Exit", "Close Phanes DNA CLI", "exit"},
	}

	for i, opt := range options {
		fmt.Printf("  [%d] %-28s — %s\n", i+1, opt.Title, opt.Description)
	}
	fmt.Println()
	fmt.Print("Enter choice [1-11]: ")

	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(options) {
		return "exit"
	}

	if options[choice-1].ActionKey == "config-ai" {
		RunConfigureAIInteractive()
		return "exit"
	}

	return options[choice-1].ActionKey
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
