package gitflow

import (
	"context"
	"testing"
)

func TestGenerateConventionalCommit_Execution(t *testing.T) {
	msg, err := GenerateConventionalCommit(context.Background(), nil, "es")
	if err != nil {
		// No staged changes error is expected when clean git index
		if err.Error() != "no staged changes found (run 'git add' first)" {
			t.Fatalf("unexpected error: %v", err)
		}
	} else if msg == "" {
		t.Errorf("expected non-empty commit message string")
	}
}
