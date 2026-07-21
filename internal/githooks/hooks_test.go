package githooks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallAndUninstallHooks(t *testing.T) {
	tmpDir := t.TempDir()
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatalf("failed to create temp .git dir: %v", err)
	}

	// Install
	if err := InstallHooks(tmpDir, "all"); err != nil {
		t.Fatalf("InstallHooks failed: %v", err)
	}

	preCommit := filepath.Join(gitDir, "hooks", "pre-commit")
	if _, err := os.Stat(preCommit); os.IsNotExist(err) {
		t.Errorf("expected pre-commit hook file at %s", preCommit)
	}

	prePush := filepath.Join(gitDir, "hooks", "pre-push")
	if _, err := os.Stat(prePush); os.IsNotExist(err) {
		t.Errorf("expected pre-push hook file at %s", prePush)
	}

	// Uninstall
	if err := UninstallHooks(tmpDir); err != nil {
		t.Fatalf("UninstallHooks failed: %v", err)
	}

	if _, err := os.Stat(preCommit); !os.IsNotExist(err) {
		t.Errorf("expected pre-commit hook to be removed")
	}
}
