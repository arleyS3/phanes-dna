package ai

import "context"

// Provider is the interface for AI connectivity (Ollama, Gemini, Anthropic, etc.).
type Provider interface {
	Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error)
	Embed(ctx context.Context, texts []string) ([][]float32, error)
	Ping(ctx context.Context) error
}

// AnalyzeRequest carries the prompt to the AI model.
type AnalyzeRequest struct {
	Prompt       string
	SystemPrompt string
	Temperature  float32
}

// AnalyzeResponse contains the model's output and token usage.
type AnalyzeResponse struct {
	Content   string
	TokensIn  int
	TokensOut int
}

// Config holds provider connection parameters.
type Config struct {
	OllamaURL       string
	GeminiKey       string
	AnthropicKey    string
	DefaultProvider string
}
