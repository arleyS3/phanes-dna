package ai

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// GeminiProvider implements Provider using Google Gemini API.
type GeminiProvider struct {
	apiKey     string
	modelName  string
	embedModel string
}

func NewGeminiProvider(apiKey, modelName, embedModel string) *GeminiProvider {
	if modelName == "" {
		modelName = "gemini-1.5-flash"
	}
	if embedModel == "" {
		embedModel = "text-embedding-004"
	}
	return &GeminiProvider{
		apiKey:     apiKey,
		modelName:  modelName,
		embedModel: embedModel,
	}
}

func (p *GeminiProvider) Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error) {
	if p.apiKey == "" {
		return nil, fmt.Errorf("gemini api key not configured")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	defer client.Close()

	model := client.GenerativeModel(p.modelName)
	if req.SystemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Parts: []genai.Part{genai.Text(req.SystemPrompt)},
		}
	}
	if req.Temperature > 0 {
		model.SetTemperature(req.Temperature)
	}

	resp, err := model.GenerateContent(ctx, genai.Text(req.Prompt))
	if err != nil {
		return nil, fmt.Errorf("generate content: %w", err)
	}

	var content string
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				if txt, ok := part.(genai.Text); ok {
					content += string(txt)
				}
			}
		}
	}

	var tokensIn, tokensOut int
	if resp.UsageMetadata != nil {
		tokensIn = int(resp.UsageMetadata.PromptTokenCount)
		tokensOut = int(resp.UsageMetadata.CandidatesTokenCount)
	}

	return &AnalyzeResponse{
		Content:   content,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
	}, nil
}

func (p *GeminiProvider) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, nil
	}
	if p.apiKey == "" {
		return nil, fmt.Errorf("gemini api key not configured")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	defer client.Close()

	em := client.EmbeddingModel(p.embedModel)
	var result [][]float32
	for _, t := range texts {
		res, err := em.EmbedContent(ctx, genai.Text(t))
		if err != nil {
			return nil, fmt.Errorf("embed content: %w", err)
		}
		if res.Embedding != nil {
			result = append(result, res.Embedding.Values)
		}
	}
	return result, nil
}

func (p *GeminiProvider) Ping(ctx context.Context) error {
	if p.apiKey == "" {
		return fmt.Errorf("no api key")
	}
	client, err := genai.NewClient(ctx, option.WithAPIKey(p.apiKey))
	if err != nil {
		return err
	}
	defer client.Close()
	return nil
}
