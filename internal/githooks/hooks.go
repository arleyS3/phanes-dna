package githooks

import (
	"fmt"
	"os"
	"path/filepath"
)

const preCommitScript = `#!/bin/sh
# Phanes DNA pre-commit hook
echo "🔍 Running Phanes DNA architecture check..."
phanes-dna review --strict
if [ $? -ne 0 ]; then
  echo "🚨 Phanes DNA: Commit blocked due to architecture violations!"
  exit 1
fi
`

const prePushScript = `#!/bin/sh
# Phanes DNA pre-push hook
echo "🔍 Running Phanes DNA architecture push check..."
phanes-dna review --strict
if [ $? -ne 0 ]; then
  echo "🚨 Phanes DNA: Push blocked due to architecture violations!"
  exit 1
fi
`

// InstallHooks writes pre-commit and pre-push hook scripts to gitDir/hooks.
func InstallHooks(gitDir string, hookType string) error {
	hooksDir := filepath.Join(gitDir, ".git", "hooks")
	if _, err := os.Stat(filepath.Join(gitDir, ".git")); os.IsNotExist(err) {
		// Fallback if gitDir itself is already .git folder
		if filepath.Base(gitDir) == ".git" {
			hooksDir = gitDir
		} else {
			return fmt.Errorf("not a git repository (missing .git directory at %s)", gitDir)
		}
	}

	if err := os.MkdirAll(hooksDir, 0755); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}

	if hookType == "all" || hookType == "pre-commit" || hookType == "" {
		p := filepath.Join(hooksDir, "pre-commit")
		if err := os.WriteFile(p, []byte(preCommitScript), 0755); err != nil {
			return fmt.Errorf("write pre-commit hook: %w", err)
		}
		fmt.Printf("  ✅ Installed Git hook -> %s\n", p)
	}

	if hookType == "all" || hookType == "pre-push" || hookType == "" {
		p := filepath.Join(hooksDir, "pre-push")
		if err := os.WriteFile(p, []byte(prePushScript), 0755); err != nil {
			return fmt.Errorf("write pre-push hook: %w", err)
		}
		fmt.Printf("  ✅ Installed Git hook -> %s\n", p)
	}

	return nil
}

// UninstallHooks removes Phanes DNA pre-commit and pre-push hooks.
func UninstallHooks(gitDir string) error {
	hooksDir := filepath.Join(gitDir, ".git", "hooks")
	_ = os.Remove(filepath.Join(hooksDir, "pre-commit"))
	_ = os.Remove(filepath.Join(hooksDir, "pre-push"))
	fmt.Println("  🗑️ Removed Phanes DNA Git hooks.")
	return nil
}
