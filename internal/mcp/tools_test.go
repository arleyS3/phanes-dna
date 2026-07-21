package mcp

import (
	"context"
	"testing"

	"github.com/arleyS3/phanes-dna/internal/dna"
	"github.com/arleyS3/phanes-dna/internal/store"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestHandleGetProjectDNA_EmptyQuery(t *testing.T) {
	st, err := store.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory store: %v", err)
	}
	defer st.Close()

	handlers := NewDNAHandlers(st, nil, 1)
	req := mcp.CallToolRequest{}
	req.Params.Name = "get_project_dna"
	req.Params.Arguments = map[string]any{}

	res, err := handlers.HandleGetProjectDNA(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.IsError {
		t.Errorf("expected error result for empty query, got success")
	}
}

func TestHandleReviewArchitecture_NoViolations(t *testing.T) {
	st, err := store.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory store: %v", err)
	}
	defer st.Close()

	handlers := NewDNAHandlers(st, nil, 1)
	req := mcp.CallToolRequest{}
	req.Params.Name = "review_architecture"

	res, err := handlers.HandleReviewArchitecture(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.IsError {
		t.Errorf("expected clean result, got error")
	}
}

func TestHandleReviewArchitecture_WithViolations(t *testing.T) {
	st, err := store.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create memory store: %v", err)
	}
	defer st.Close()

	ctx := context.Background()
	_, _ = st.InsertViolation(ctx, &dna.Violation{
		Rule:     "LayerRuleViolation",
		Severity: "error",
		Location: "com/example/service/UserService.java",
		Message:  "Service layer cannot depend directly on Controller layer",
	})

	handlers := NewDNAHandlers(st, nil, 1)
	req := mcp.CallToolRequest{}
	req.Params.Name = "review_architecture"

	res, err := handlers.HandleReviewArchitecture(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.IsError {
		t.Errorf("expected non-error result with formatted violations, got error")
	}
}
