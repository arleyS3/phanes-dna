package githooks

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

const preCommitScript = `#!/bin/sh
# Phanes DNA pre-commit hook
# Resolve phanes-dna binary path (prioritize local dev binary)
PHANES_BIN="phanes-dna"
if [ -x "./phanes-dna" ]; then
  PHANES_BIN="./phanes-dna"
elif [ -x "$HOME/.phanes-dna/bin/phanes-dna" ]; then
  PHANES_BIN="$HOME/.phanes-dna/bin/phanes-dna"
fi

# Determine staged files
staged_files=$(git diff --cached --name-only)

if [ -n "$staged_files" ]; then
  violations_output=$($PHANES_BIN review)
  blocked=false
  
  for file in $staged_files; do
    # Search for the file in the violations output
    if echo "$violations_output" | grep -q "Location: $file"; then
      if [ "$blocked" = false ]; then
        echo "🚨 Phanes DNA: Violations detected in your staged files:"
        blocked=true
      fi
      # Print the violation block for this file
      echo "$violations_output" | grep -B 1 -A 1 "Location: $file"
    fi
  done

  if [ "$blocked" = true ]; then
    echo "🚨 Phanes DNA: Commit blocked due to architecture/naming violations in your staged files!"
    exit 1
  fi
fi
`

const commitMsgScript = `#!/bin/sh
# Phanes DNA commit-msg hook
# Resolve phanes-dna binary path (prioritize local dev binary)
PHANES_BIN="phanes-dna"
if [ -x "./phanes-dna" ]; then
  PHANES_BIN="./phanes-dna"
elif [ -x "$HOME/.phanes-dna/bin/phanes-dna" ]; then
  PHANES_BIN="$HOME/.phanes-dna/bin/phanes-dna"
fi

commit_msg_file=$1
echo "🔍 Running Phanes DNA commit message check..."
$PHANES_BIN review-commit "$commit_msg_file"
if [ $? -ne 0 ]; then
  echo "🚨 Phanes DNA: Commit blocked due to message violations!"
  exit 1
fi
`

const prePushScript = `#!/bin/sh
# Phanes DNA pre-push hook
# Resolve phanes-dna binary path (prioritize local dev binary)
PHANES_BIN="phanes-dna"
if [ -x "./phanes-dna" ]; then
  PHANES_BIN="./phanes-dna"
elif [ -x "$HOME/.phanes-dna/bin/phanes-dna" ]; then
  PHANES_BIN="$HOME/.phanes-dna/bin/phanes-dna"
fi

# Read stdin lines from Git: local_ref local_sha remote_ref remote_sha
while read local_ref local_sha remote_ref remote_sha
do
  # Skip if delete branch push
  if [ "$local_sha" = "0000000000000000000000000000000000000000" ]; then
    continue
  fi
  
  # Determine changed files in this push
  if [ "$remote_sha" = "0000000000000000000000000000000000000000" ]; then
    changed_files=$(git diff --name-only HEAD --not --remotes 2>/dev/null || git diff --name-only HEAD~1 HEAD 2>/dev/null)
    commits=$(git rev-list HEAD --not --remotes 2>/dev/null || git rev-list -n 1 HEAD)
  else
    changed_files=$(git diff --name-only $remote_sha..$local_sha)
    commits=$(git rev-list $remote_sha..$local_sha)
  fi

  # Run architecture review only if there are changed files
  if [ -n "$changed_files" ]; then
    violations_output=$($PHANES_BIN review)
    blocked=false
    
    for file in $changed_files; do
      # Search for the file in the violations output
      if echo "$violations_output" | grep -q "Location: $file"; then
        if [ "$blocked" = false ]; then
          echo "🚨 Phanes DNA: Violations detected in your pushed files:"
          blocked=true
        fi
        # Print the violation block for this file
        echo "$violations_output" | grep -B 1 -A 1 "Location: $file"
      fi
    done

    if [ "$blocked" = true ]; then
      echo "🚨 Phanes DNA: Push blocked due to architecture/naming violations in your changes!"
      exit 1
    fi
  fi

  for commit in $commits
  do
    # Run commit message validation on each commit message
    $PHANES_BIN review-commit-sha "$commit"
    if [ $? -ne 0 ]; then
      echo "🚨 Phanes DNA: Push blocked due to invalid commit messages!"
      exit 1
    fi
  done
done
exit 0
`

// InstallHooks escribe los scripts de los hooks de git (pre-commit, commit-msg, y pre-push) en la carpeta gitDir/hooks.
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

	if hookType == "all" || hookType == "commit-msg" || hookType == "" {
		p := filepath.Join(hooksDir, "commit-msg")
		if err := os.WriteFile(p, []byte(commitMsgScript), 0755); err != nil {
			return fmt.Errorf("write commit-msg hook: %w", err)
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

// UninstallHooks elimina los hooks físicos creados por Phanes DNA.
func UninstallHooks(gitDir string) error {
	hooksDir := filepath.Join(gitDir, ".git", "hooks")
	_ = os.Remove(filepath.Join(hooksDir, "pre-commit"))
	_ = os.Remove(filepath.Join(hooksDir, "commit-msg"))
	_ = os.Remove(filepath.Join(hooksDir, "pre-push"))
	fmt.Println("  🗑️ Removed Phanes DNA Git hooks.")
	return nil
}

// ValidateCommitMessageFromFile lee un mensaje de commit desde un archivo y lo valida contra las reglas de PHANES_RULES.md.
func ValidateCommitMessageFromFile(msgFilePath string, rulesFilePath string) error {
	msgBytes, err := os.ReadFile(msgFilePath)
	if err != nil {
		return fmt.Errorf("read commit message file: %w", err)
	}
	msg := strings.TrimSpace(string(msgBytes))

	// Skip validation if it's a merge commit or a revert commit that Git automatically generates
	if strings.HasPrefix(msg, "Merge branch") || strings.HasPrefix(msg, "Merge remote-tracking branch") || strings.HasPrefix(msg, "Revert ") {
		return nil
	}

	// 1. Read PHANES_RULES.md to find git conventions
	file, err := os.Open(rulesFilePath)
	if err != nil {
		// If PHANES_RULES.md does not exist, commit validation is skipped
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inGitConventions := false
	conventionalCommitsRequired := false
	gitmojiRequired := false

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "## ") {
			if strings.Contains(strings.ToLower(trimmed), "git conventions") {
				inGitConventions = true
			} else {
				inGitConventions = false
			}
			continue
		}

		if inGitConventions && strings.HasPrefix(trimmed, "-") {
			lowerLine := strings.ToLower(trimmed)
			if strings.Contains(lowerLine, "conventional commits") {
				conventionalCommitsRequired = true
			}
			if strings.Contains(lowerLine, "gitmoji") {
				gitmojiRequired = true
			}
		}
	}

	// 2. Perform validations
	if conventionalCommitsRequired {
		// Regex for Conventional Commits: type(scope): description
		re := regexp.MustCompile(`^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(?:\([a-zA-Z0-9_.-]+\))?: .+$`)
		if !re.MatchString(msg) {
			return fmt.Errorf("commit message does not follow Conventional Commits format (e.g., 'feat(auth): add login' or 'fix: resolve bug')")
		}
	}

	if gitmojiRequired {
		hasEmoji := false
		if strings.HasPrefix(msg, ":") {
			re := regexp.MustCompile(`^:[a-zA-Z0-9_-]+:`)
			hasEmoji = re.MatchString(msg)
		} else {
			runes := []rune(msg)
			if len(runes) > 0 {
				r := runes[0]
				// Basic emoji unicode ranges
				if (r >= 0x1F300 && r <= 0x1F9FF) || (r >= 0x2600 && r <= 0x26FF) || (r >= 0x2700 && r <= 0x27BF) {
					hasEmoji = true
				}
			}
		}
		if !hasEmoji {
			return fmt.Errorf("commit message does not start with a Gitmoji (e.g., '✨ feat: add login' or ':sparkles: feat: add login')")
		}
	}

	return nil
}

// ValidateCommitSha valida el mensaje de un commit ya registrado por su SHA contra las convenciones en PHANES_RULES.md.
func ValidateCommitSha(sha string, rulesFilePath string) error {
	cmd := exec.Command("git", "log", "--format=%B", "-n", "1", sha)
	msgBytes, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to read commit SHA %s: %w", sha, err)
	}

	// Create a temp file to feed ValidateCommitMessageFromFile
	tmpFile, err := os.CreateTemp("", "phanes-commit-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(msgBytes); err != nil {
		return err
	}

	return ValidateCommitMessageFromFile(tmpFile.Name(), rulesFilePath)
}
