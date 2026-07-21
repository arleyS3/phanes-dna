package ai

import (
	"context"
	"fmt"
	"os"
	"time"
)

// AutoDetectProvider probes available AI providers according to design decision D5:
// 1. Ollama on localhost:11434 (if reachable)
// 2. GEMINI_API_KEY environment variable
// 3. ANTHROPIC_API_KEY environment variable
func AutoDetectProvider(cfg Config) (Provider, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	userCfg, _ := LoadUserConfig()
	if cfg.OllamaURL == "" {
		cfg.OllamaURL = userCfg.OllamaURL
	}
	if cfg.GeminiKey == "" {
		cfg.GeminiKey = userCfg.GeminiKey
	}
	if cfg.AnthropicKey == "" {
		cfg.AnthropicKey = userCfg.AnthropicKey
	}

	// 1. Try Ollama
	ollamaURL := cfg.OllamaURL
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	ollama := NewOllamaProvider(ollamaURL, "", "")
	if err := ollama.Ping(ctx); err == nil {
		return ollama, "ollama", nil
	}

	// 2. Try Gemini
	geminiKey := cfg.GeminiKey
	if geminiKey == "" {
		geminiKey = os.Getenv("GEMINI_API_KEY")
	}
	if geminiKey != "" {
		gemini := NewGeminiProvider(geminiKey, "", "")
		return gemini, "gemini", nil
	}

	// 3. Try Anthropic
	anthropicKey := cfg.AnthropicKey
	if anthropicKey == "" {
		anthropicKey = os.Getenv("ANTHROPIC_API_KEY")
	}
	if anthropicKey != "" {
		anthropic := NewAnthropicProvider(anthropicKey, "")
		return anthropic, "anthropic", nil
	}

	return nil, "", fmt.Errorf("no AI provider reachable (Ollama down and GEMINI_API_KEY/ANTHROPIC_API_KEY not set)")
}
