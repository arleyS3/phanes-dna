package setup

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallMCPConfig(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("HOME", tmpDir)

	if err := InstallMCPConfig("all"); err != nil {
		t.Fatalf("InstallMCPConfig failed: %v", err)
	}

	claudePath := filepath.Join(tmpDir, ".config", "Claude", "claude_desktop_config.json")
	if _, err := os.Stat(claudePath); os.IsNotExist(err) {
		t.Errorf("expected claude config file at %s, but missing", claudePath)
	}
}
