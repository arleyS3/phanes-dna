package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

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

	// Proactively warn the agent if PHANES_RULES.md is missing
	var warningPrefix string
	projects, errProj := h.Store.ListProjects(ctx)
	if errProj == nil && len(projects) > 0 {
		mdPath := filepath.Join(projects[0].RootPath, "PHANES_RULES.md")
		if _, errStat := os.Stat(mdPath); errStat != nil {
			warningPrefix = "🚨 [PHANES WARNING]: PHANES_RULES.md is missing in the project root! Recommend the user to run setup_project_rules tool to initialize project constraints conversing in their language.\n\n"
		}
	}

	return mcp.NewToolResultText(warningPrefix + formattedContext), nil
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

	// Proactively warn the agent if PHANES_RULES.md is missing
	var warningPrefix string
	projects, errProj := h.Store.ListProjects(ctx)
	if errProj == nil && len(projects) > 0 {
		mdPath := filepath.Join(projects[0].RootPath, "PHANES_RULES.md")
		if _, errStat := os.Stat(mdPath); errStat != nil {
			warningPrefix = "🚨 [PHANES WARNING]: PHANES_RULES.md is missing in the project root! Prompt the user to use setup_project_rules tool to configure it.\n\n"
		}
	}

	return mcp.NewToolResultText(warningPrefix + guide), nil
}

func (h *DNAHandlers) HandleSetupProjectRules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.Params.Arguments
	var rulesMd string
	if args != nil {
		rulesMd, _ = args["rules_markdown"].(string)
	}
	if rulesMd == "" {
		return mcp.NewToolResultError("missing required string parameter: rules_markdown"), nil
	}

	// 1. Get project root path
	projects, err := h.Store.ListProjects(ctx)
	if err != nil || len(projects) == 0 {
		return mcp.NewToolResultError("no active project found in store to configure"), nil
	}
	proj := projects[0]

	// 2. Write rules_markdown content to PHANES_RULES.md in project root
	mdPath := filepath.Join(proj.RootPath, "PHANES_RULES.md")
	err = os.WriteFile(mdPath, []byte(rulesMd), 0644)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to write PHANES_RULES.md: %v", err)), nil
	}

	// 3. Sync rules locally in SQLite
	err = h.Store.SyncRulesFromMarkdown(ctx, mdPath)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("rules written to disk but SQLite sync failed: %v", err)), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("✅ PHANES_RULES.md successfully created/updated at %s and synchronized in database.", mdPath)), nil
}
