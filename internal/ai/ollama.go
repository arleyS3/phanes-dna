package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaProvider implements Provider for local Ollama instances.
type OllamaProvider struct {
	baseURL      string
	genModel     string
	embedModel   string
	httpClient   *http.Client
}

func NewOllamaProvider(baseURL, genModel, embedModel string) *OllamaProvider {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if genModel == "" {
		genModel = "llama3"
	}
	if embedModel == "" {
		embedModel = "nomic-embed-text"
	}
	return &OllamaProvider{
		baseURL:    baseURL,
		genModel:   genModel,
		embedModel: embedModel,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}
}

type ollamaGenerateReq struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	System string `json:"system,omitempty"`
	Stream bool   `json:"stream"`
}

type ollamaGenerateResp struct {
	Response           string `json:"response"`
	PromptEvalCount   int    `json:"prompt_eval_count"`
	EvalCount         int    `json:"eval_count"`
}

type ollamaEmbedReq struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type ollamaEmbedResp struct {
	Embeddings [][]float32 `json:"embeddings"`
}

func (p *OllamaProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	body, err := json.Marshal(ollamaGenerateReq{
		Model:  p.genModel,
		Prompt: req.Prompt,
		System: req.SystemPrompt,
		Stream: false,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal generate req: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create req: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama generate call: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama generate status %d", httpResp.StatusCode)
	}

	var res ollamaGenerateResp
	if err := json.NewDecoder(httpResp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode generate resp: %w", err)
	}

	return &AnalyzeResponse{
		Content:   res.Response,
		TokensIn:  res.PromptEvalCount,
		TokensOut: res.EvalCount,
	}, nil
}

func (p *OllamaProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}

	body, err := json.Marshal(ollamaEmbedReq{
		Model: p.embedModel,
		Input: texts,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal embed req: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create embed req: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	httpResp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama embed call: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama embed status %d", httpResp.StatusCode)
	}

	var res ollamaEmbedResp
	if err := json.NewDecoder(httpResp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("decode embed resp: %w", err)
	}

	return res.Embeddings, nil
}

func (p *OllamaProvider) Ping(ctx context.Context) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/", nil)
	if err != nil {
		return err
	}
	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return nil
}
