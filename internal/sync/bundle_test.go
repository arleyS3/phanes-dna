package sync

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/arleyS3/phanes-dna/internal/dna"
	"github.com/arleyS3/phanes-dna/internal/store"
)

func TestExportAndImportBundle(t *testing.T) {
	st, err := store.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open memory store: %v", err)
	}
	defer st.Close()

	ctx := context.Background()

	// Insert test project, rule, and violation
	projID, err := st.InsertProject(ctx, &dna.Project{Name: "SocialImpactProject", RootPath: "/tmp/social"})
	if err != nil {
		t.Fatalf("failed to insert project: %v", err)
	}

	_, _ = st.InsertLayerRule(ctx, &dna.LayerRule{
		Layer:       "controller",
		TargetLayer: "service",
		Allowed:     true,
		Severity:    "info",
	})

	_, _ = st.InsertViolation(ctx, &dna.Violation{
		Rule:     "NoDirectDBAccessInController",
		Severity: "error",
		Location: "controllers/user.go",
		Message:  "Controllers must not execute raw SQL queries",
	})

	tmpDir := t.TempDir()
	dnaPath := filepath.Join(tmpDir, "project.dna")

	// Export bundle
	if err := ExportBundle(ctx, st, projID, dnaPath); err != nil {
		t.Fatalf("ExportBundle failed: %v", err)
	}

	// Create second empty store for import
	st2, err := store.NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to open target memory store: %v", err)
	}
	defer st2.Close()

	// Import bundle
	bundle, err := ImportBundle(ctx, st2, dnaPath)
	if err != nil {
		t.Fatalf("ImportBundle failed: %v", err)
	}

	if bundle.ProjectName != "SocialImpactProject" {
		t.Errorf("expected ProjectName 'SocialImpactProject', got %q", bundle.ProjectName)
	}

	rules, err := st2.GetLayerRules(ctx)
	if err != nil || len(rules) == 0 {
		t.Fatalf("expected imported layer rules, got %v (err: %v)", rules, err)
	}

	violations, err := st2.GetViolations(ctx)
	if err != nil || len(violations) == 0 {
		t.Fatalf("expected imported violations, got %v (err: %v)", violations, err)
	}
}
