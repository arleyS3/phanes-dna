package gitflow

import (
	"context"
	"testing"
)

func TestGenerateConventionalCommit_NoStagedChanges(t *testing.T) {
	_, err := GenerateConventionalCommit(context.Background(), nil, "es")
	if err == nil {
		t.Errorf("expected error when no staged changes present, got nil")
	}
}
