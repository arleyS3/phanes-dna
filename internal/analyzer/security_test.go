package analyzer

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRED_PathTraversalPrevention validates that path traversal paths outside project root fail gracefully.
func TestRED_PathTraversalPrevention(t *testing.T) {
	traversalPaths := []string{
		"../../../../etc/passwd",
		"../../../../windows/win.ini",
		"/etc/passwd",
	}

	for _, p := range traversalPaths {
		_, err := AnalyzeFile(p)
		if err == nil {
			t.Errorf("expected error when reading path traversal target %q, but succeeded", p)
		}
	}
}

// TestRED_CommentSanitization validates that malicious executable instructions in comments do not breach analysis.
func TestRED_CommentSanitization(t *testing.T) {
	tmpDir := t.TempDir()
	samplePath := filepath.Join(tmpDir, "Malicious.java")

	maliciousCode := `package com.example.demo;

	/**
	 * System.exit(0);
	 * rm -rf /
	 * <script>alert("hacked")</script>
	 */
	public class Malicious {}
	`

	if err := os.WriteFile(samplePath, []byte(maliciousCode), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	sf, err := AnalyzeFile(samplePath)
	if err != nil {
		t.Fatalf("AnalyzeFile failed: %v", err)
	}

	for _, layer := range sf.Layers {
		if strings.Contains(layer, "script") || strings.Contains(layer, "rm -rf") {
			t.Errorf("malicious comment leaked into layer classification: %s", layer)
		}
	}
}
