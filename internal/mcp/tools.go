package mcp

import (
	"context"
	"fmt"

	"github.com/arleyS3/phanes-dna/internal/ai"
	"github.com/arleyS3/phanes-dna/internal/dna"
	"github.com/arleyS3/phanes-dna/internal/onboard"
	"github.com/arleyS3/phanes-dna/internal/store"
	"github.com/mark3labs/mcp-go/mcp"
)

type DNAHandlers struct {
	Store     *store.Store
	Provider  ai.Provider
	ProjectID int64
}

func NewDNAHandlers(st *store.Store, prov ai.Provider, projectID int64) *DNAHandlers {
	return &DNAHandlers{
		Store:     st,
		Provider:  prov,
		ProjectID: projectID,
	}
}

func (h *DNAHandlers) HandleGetProjectDNA(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	var query string
	if args != nil {
		query, _ = args["query"].(string)
	}
	if query == "" {
		return mcp.NewToolResultError("missing required string parameter: query"), nil
	}

	// 1. Embed query if provider is available
	var queryEmb []float32
	if h.Provider != nil {
		embeddings, err := h.Provider.Embed(ctx, []string{query})
		if err == nil && len(embeddings) > 0 {
			queryEmb = embeddings[0]
		}
	}

	// 2. Vector search in Store
	topChunks, err := h.Store.SearchSimilar(ctx, queryEmb, 5)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("vector search failed: %v", err)), nil
	}

	// 3. Compress with Caveman and format context
	var formattedContext string
	for _, c := range topChunks {
		res := dna.Compress(c.Content, dna.Normal)
		formattedContext += fmt.Sprintf("--- Chunk (File: %s, Lines %d-%d) ---\n%s\n\n", c.FilePath, c.StartLine, c.EndLine, res.Text)
	}

	if formattedContext == "" {
		formattedContext = "No relevant architectural context found for query."
	}

	return mcp.NewToolResultText(formattedContext), nil
}

func (h *DNAHandlers) HandleReviewArchitecture(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	violations, err := h.Store.GetViolations(ctx)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to list violations: %v", err)), nil
	}

	if len(violations) == 0 {
		return mcp.NewToolResultText("✅ No architecture violations detected in project."), nil
	}

	var result string
	for i, v := range violations {
		result += fmt.Sprintf("[%d] %s: %s (Severity: %s)\n    Location: %s\n    Recommendation: %s\n\n",
			i+1, v.Rule, v.Message, v.Severity, v.Location, v.Message)
	}

	return mcp.NewToolResultText(result), nil
}

func (h *DNAHandlers) HandleDevOnboarding(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	var topic string
	if args != nil {
		topic, _ = args["topic"].(string)
	}

	guide, err := onboard.GenerateOnboardingGuide(ctx, h.Store, h.Provider, h.ProjectID, topic)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("onboarding guide failed: %v", err)), nil
	}

	return mcp.NewToolResultText(guide), nil
}
