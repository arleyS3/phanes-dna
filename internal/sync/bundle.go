package sync

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/arleyS3/phanes-dna/internal/dna"
	"github.com/arleyS3/phanes-dna/internal/store"
)

// DNABundle is an ultralight, Git-syncable representation of project rules and architecture.
type DNABundle struct {
	Version     string            `json:"version"`
	ExportedAt  time.Time         `json:"exported_at"`
	ProjectName string            `json:"project_name"`
	LayerRules  []*dna.LayerRule  `json:"layer_rules"`
	Violations  []*dna.Violation `json:"violations"`
}

// ExportBundle serialises project layer rules and violations to a gzip-compressed .dna file.
func ExportBundle(ctx context.Context, st *store.Store, projectID int64, outputPath string) error {
	proj, err := st.GetProject(ctx, projectID)
	projName := "unknown"
	if err == nil && proj != nil {
		projName = proj.Name
	}

	rules, err := st.GetLayerRules(ctx)
	if err != nil {
		return fmt.Errorf("get layer rules: %w", err)
	}

	violations, err := st.GetViolations(ctx)
	if err != nil {
		return fmt.Errorf("get violations: %w", err)
	}

	bundle := DNABundle{
		Version:     "0.1.0",
		ExportedAt:  time.Now().UTC(),
		ProjectName: projName,
		LayerRules:  rules,
		Violations:  violations,
	}

	data, err := json.MarshalIndent(bundle, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal bundle: %w", err)
	}

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("create bundle file: %w", err)
	}
	defer outFile.Close()

	gw := gzip.NewWriter(outFile)
	defer gw.Close()

	if _, err := gw.Write(data); err != nil {
		return fmt.Errorf("write compressed bundle: %w", err)
	}

	return nil
}

// ImportBundle reads a .dna bundle and syncs layer rules & violations into the local SQLite store.
func ImportBundle(ctx context.Context, st *store.Store, inputPath string) (*DNABundle, error) {
	inFile, err := os.Open(inputPath)
	if err != nil {
		return nil, fmt.Errorf("open bundle file: %w", err)
	}
	defer inFile.Close()

	gr, err := gzip.NewReader(inFile)
	if err != nil {
		return nil, fmt.Errorf("read gzip bundle: %w", err)
	}
	defer gr.Close()

	data, err := io.ReadAll(gr)
	if err != nil {
		return nil, fmt.Errorf("decompress bundle: %w", err)
	}

	var bundle DNABundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		return nil, fmt.Errorf("unmarshal bundle: %w", err)
	}

	// Sync LayerRules
	for _, r := range bundle.LayerRules {
		_, _ = st.InsertLayerRule(ctx, r)
	}

	// Sync Violations
	for _, v := range bundle.Violations {
		_, _ = st.InsertViolation(ctx, v)
	}

	return &bundle, nil
}
