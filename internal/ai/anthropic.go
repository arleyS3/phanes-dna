package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AnthropicProvider implements Provider via Anthropic Messages API.
type AnthropicProvider struct {
	apiKey     string
	modelName  string
	httpClient *http.Client
}

func NewAnthropicProvider(apiKey, modelName string) *AnthropicProvider {
	if modelName == "" {
		modelName = "claude-3-5-sonnet-20241022"
	}
	return &AnthropicProvider{
		apiKey:     apiKey,
		modelName:  modelName,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicReq struct {
	Model     string         `json:"model"`
	MaxTokens int            `json:"max_tokens"`
	System    string         `json:"system,omitempty"`
	Messages  []anthropicMsg `json:"messages"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicResp struct {
	Content []anthropicContentBlock `json:"content"`
	Usage   anthropicUsage          `json:"usage"`
}

func (p *AnthropicProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("anthropic api key not configured")
	}

	body, err := json.Marshal(anthropicReq{
		Model:     p.modelName,
		MaxTokens: 4096,
		System:    req.SystemPrompt,
		Messages: []anthropicMsg{
			{Role: "user", Content: req.Prompt},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create http request: %w", err)
	}
	httpReq.Header.Set("x-api-key", p.apiKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("anthropic call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic status %d", resp.StatusCode)
	}

	var res anthropicResp
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	var content string
	for _, blk := range res.Content {
		if blk.Type == "text" {
			content += blk.Text
		}
	}

	return &AnalyzeResponse{
		Content:   content,
		TokensIn:  res.Usage.InputTokens,
		TokensOut: res.Usage.OutputTokens,
	}, nil
}

func (p *AnthropicProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	// Anthropic API does not offer direct embeddings; fallback to zero/passthrough or error
	return nil, fmt.Errorf("anthropic does not provide native vector embeddings endpoint; use Ollama or Gemini for embeddings")
}

func (p *AnthropicProvider) Ping(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("no anthropic api key")
	}
	return nil
}
