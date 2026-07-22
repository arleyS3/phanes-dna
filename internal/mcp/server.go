package mcp

import (
	"github.com/arleyS3/phanes-dna/internal/ai"
	"github.com/arleyS3/phanes-dna/internal/store"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ServeStdio starts an MCP stdio server registered with Phanes DNA tools.
func ServeStdio(st *store.Store, prov ai.Provider, projectID int64) error {
	s := server.NewMCPServer("phanes-dna", "0.1.0")

	handlers := NewDNAHandlers(st, prov, projectID)

	// Tool 1: get_project_dna
	toolDNA := mcp.NewTool("get_project_dna",
		mcp.WithDescription("Query indexed architecture rules, AST definitions, and design context compressed via Caveman"),
		mcp.WithString("query", mcp.Required(), mcp.Description("Architecture query or concept search")),
	)
	s.AddTool(toolDNA, handlers.HandleGetProjectDNA)

	// Tool 2: review_architecture
	toolReview := mcp.NewTool("review_architecture",
		mcp.WithDescription("Inspect project layer dependencies and report architecture violations"),
	)
	s.AddTool(toolReview, handlers.HandleReviewArchitecture)

	// Tool 3: dev_onboarding
	toolOnboard := mcp.NewTool("dev_onboarding",
		mcp.WithDescription("Interactive Dev Onboarding Mentor answering architectural patterns, feature implementation guides, and coding conventions"),
		mcp.WithString("topic", mcp.Description("Onboarding topic or feature question")),
	)
	s.AddTool(toolOnboard, handlers.HandleDevOnboarding)

	// Tool 4: setup_project_rules
	toolSetupRules := mcp.NewTool("setup_project_rules",
		mcp.WithDescription("Write or update PHANES_RULES.md in the project root containing architectural dependencies and conventions, and sync rules dynamically in SQLite"),
		mcp.WithString("rules_markdown", mcp.Required(), mcp.Description("Markdown content conforming to the PHANES_RULES.md structure")),
	)
	s.AddTool(toolSetupRules, handlers.HandleSetupProjectRules)

	return server.ServeStdio(s)
}
